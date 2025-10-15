package memory

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// DefaultModelURL is the URL to download the gtr-t5-base ONNX model
	DefaultModelURL = "https://downloads.poiesic.com/wonda/gtr-t5-base-onnx-1.0.0.tar.gz"

	// ModelVersion is the current model version
	ModelVersion = "1.0.0"

	// ModelDirName is the name of the model directory
	ModelDirName = "gtr-t5-base-onnx"
)

// ModelDownloader handles downloading and caching the ONNX model.
type ModelDownloader struct {
	modelURL  string
	cacheDir  string
	modelDir  string
	showProgress bool
}

// NewModelDownloader creates a new model downloader.
// cacheDir is typically ~/.config/wonda/models/
func NewModelDownloader(cacheDir string, modelURL string) *ModelDownloader {
	if modelURL == "" {
		modelURL = DefaultModelURL
	}

	modelDir := filepath.Join(cacheDir, ModelDirName)

	return &ModelDownloader{
		modelURL:     modelURL,
		cacheDir:     cacheDir,
		modelDir:     modelDir,
		showProgress: true,
	}
}

// EnsureModelAvailable checks if the model is cached, downloads if needed.
// Returns the path to the model directory.
func (d *ModelDownloader) EnsureModelAvailable() (string, error) {
	// Check if model directory exists and has required files
	if d.isModelCached() {
		return d.modelDir, nil
	}

	// Need to download
	fmt.Printf("Downloading ONNX embedding model from %s...\n", d.modelURL)
	fmt.Printf("This is a one-time download (~200MB).\n")

	if err := d.downloadAndExtract(); err != nil {
		return "", fmt.Errorf("failed to download model: %w", err)
	}

	// Verify extraction
	if !d.isModelCached() {
		return "", fmt.Errorf("model download succeeded but files are missing")
	}

	fmt.Printf("✓ Model downloaded and cached to %s\n", d.modelDir)
	return d.modelDir, nil
}

// isModelCached checks if the model is already downloaded and extracted.
func (d *ModelDownloader) isModelCached() bool {
	requiredFiles := []string{
		filepath.Join(d.modelDir, "model.onnx"),
		filepath.Join(d.modelDir, "tokenizer.json"),
		filepath.Join(d.modelDir, "metadata.json"),
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); err != nil {
			return false
		}
	}

	// TODO: Add version checking by reading metadata.json
	// For now, just check if files exist

	return true
}

// downloadAndExtract downloads the model tar.gz and extracts it.
func (d *ModelDownloader) downloadAndExtract() error {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(d.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Download to temp file
	tempFile := filepath.Join(d.cacheDir, "model-download.tar.gz")
	defer os.Remove(tempFile)

	if err := d.downloadFile(tempFile); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Extract
	fmt.Println("Extracting model files...")
	if err := d.extractTarGz(tempFile); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	return nil
}

// downloadFile downloads from the URL to the destination path.
func (d *ModelDownloader) downloadFile(dest string) error {
	resp, err := http.Get(d.modelURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Show progress if enabled
	if d.showProgress {
		totalBytes := resp.ContentLength
		return d.copyWithProgress(out, resp.Body, totalBytes)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

// copyWithProgress copies data while showing progress.
func (d *ModelDownloader) copyWithProgress(dst io.Writer, src io.Reader, totalBytes int64) error {
	buf := make([]byte, 32*1024) // 32KB buffer
	var written int64

	for {
		nr, err := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				return ew
			}
			if nr != nw {
				return io.ErrShortWrite
			}

			// Print progress
			if totalBytes > 0 {
				percent := float64(written) / float64(totalBytes) * 100
				fmt.Printf("\rDownloading: %.1f%% (%d/%d MB)",
					percent,
					written/(1024*1024),
					totalBytes/(1024*1024))
			}
		}
		if err != nil {
			if err == io.EOF {
				fmt.Println() // newline after progress
				return nil
			}
			return err
		}
	}
}

// extractTarGz extracts a .tar.gz file to the cache directory.
func (d *ModelDownloader) extractTarGz(archivePath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Construct destination path
		target := filepath.Join(d.cacheDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create parent directory if needed
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			// Create file
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
