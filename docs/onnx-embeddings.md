# ONNX In-Process Embeddings

## Overview

Wonda uses in-process ONNX embeddings for the memory system, eliminating the need for external services like Ollama or LM Studio. The embedding model (`gtr-t5-base`) is vec2text-compatible, enabling advanced cognitive distortion features.

## Architecture

```
User runs Wonda
    ↓
Check if model cached (~/.config/wonda/models/)
    ↓
If not cached: Download from https://downloads.poiesic.com/wonda/
    ↓
Load model with ONNXRuntime
    ↓
Generate embeddings (768 dimensions)
```

## Components

### 1. Embedding Model

- **Model**: `sentence-transformers/gtr-t5-base` (exported to ONNX)
- **Version**: 1.0.0
- **Size**: ~197MB compressed, ~420MB extracted
- **Dimensions**: 768
- **Max tokens**: 512
- **Vec2text compatible**: Yes (required for cognitive distortions)

**Download URL**: `https://downloads.poiesic.com/wonda/gtr-t5-base-onnx-1.0.0.tar.gz`

**Cache location**: `~/.config/wonda/models/gtr-t5-base-onnx/`

**Files**:
- `model.onnx` - ONNX model file (~419MB)
- `tokenizer.json` - HuggingFace tokenizer
- `metadata.json` - Model metadata including version

### 2. Runtime Libraries

Wonda requires two shared libraries at runtime:

#### ONNXRuntime (v1.22.0)
- **Purpose**: Run ONNX model inference
- **Size**: ~21MB
- **License**: MIT
- **Source**: https://github.com/microsoft/onnxruntime

**Platform-specific files**:
- Linux: `libonnxruntime.so.1.22.0`, `libonnxruntime_providers_shared.so`
- macOS: `libonnxruntime.dylib`, `libonnxruntime_providers_shared.dylib`
- Windows: `onnxruntime.dll`, `onnxruntime_providers_shared.dll`

#### HuggingFace Tokenizers
- **Purpose**: Tokenize text for embedding
- **Size**: ~48MB
- **License**: Apache 2.0
- **Source**: https://github.com/huggingface/tokenizers

**Platform-specific files**:
- Linux: `libtokenizers.a` (static library)
- macOS: `libtokenizers.a`
- Windows: `tokenizers.lib`

## Distribution Strategy

### Option 1: Bundle Libraries with Binary (Recommended)

Include runtime libraries in the release package:

```
wonda-linux-amd64/
├── wonda                          # Binary
└── lib/
    ├── libonnxruntime.so.1.22.0
    ├── libonnxruntime_providers_shared.so
    └── libtokenizers.a
```

**Pros**:
- Works immediately after extraction
- No manual setup required
- Consistent experience across installs

**Cons**:
- Larger download size (~50MB extra)
- Need separate packages per platform

### Option 2: Download on First Run

Download libraries when first needed:

**Pros**:
- Smaller initial binary
- Can update libraries independently

**Cons**:
- First-run network dependency
- More complex error handling
- Platform detection required

### Recommended: Option 1 (Bundle)

For Wonda MVP, bundle libraries for simplicity. Users just extract and run.

## Build Process

### Development

Set library path when building/running:

```bash
# Linux/macOS
export LD_LIBRARY_PATH="$(pwd)/lib:$LD_LIBRARY_PATH"
export CGO_LDFLAGS="-L$(pwd)/lib"

go build ./cmd/wonda
./wonda
```

### Release Builds

1. **Download runtime libraries** for target platform
2. **Place in `lib/` directory** alongside binary
3. **Package together** in tar.gz/zip

**Linux example**:
```bash
# Get ONNXRuntime (GPU-enabled with CUDA support)
# Note: Requires CUDA 12.x installed on your system
curl -L -o onnxruntime.tgz \
  https://github.com/microsoft/onnxruntime/releases/download/v1.22.0/onnxruntime-linux-x64-gpu-1.22.0.tgz
tar -xzf onnxruntime.tgz
cp onnxruntime-linux-x64-gpu-1.22.0/lib/libonnxruntime* ./lib/

# Get Tokenizers
curl -L -o libtokenizers.tar.gz \
  https://github.com/daulet/tokenizers/releases/download/v1.23.0/libtokenizers.linux-amd64.tar.gz
tar -xzf libtokenizers.tar.gz
cp libtokenizers.a ./lib/

# Build
CGO_ENABLED=1 CGO_LDFLAGS="-L$(pwd)/lib" go build -o wonda ./cmd/main.go

# Package
tar -czf wonda-linux-amd64.tar.gz wonda lib/
```

## Usage

### Go API

```go
import "github.com/poiesic/wonda/internal/memory"

// Auto-downloads model if not cached
embedder, err := memory.NewONNXEmbedderWithDownload(
    "~/.config/wonda/models",
    "" // uses default URL
)
if err != nil {
    log.Fatal(err)
}
defer embedder.Destroy()

// Generate embedding
ctx := context.Background()
embedding, err := embedder.Embed(ctx, "Hello, world!")
// Returns []float32 with 768 dimensions
```

### Environment Variables

**Runtime library path** (if not bundled in standard location):
```bash
export LD_LIBRARY_PATH=/path/to/wonda/lib:$LD_LIBRARY_PATH
```

**Custom model URL** (for self-hosting):
```bash
export WONDA_MODEL_URL=https://your-server.com/models/gtr-t5-base-onnx-1.0.0.tar.gz
```

## Platform Support

| Platform | Architecture | Status | Notes |
|----------|-------------|--------|-------|
| Linux | x86_64 | ✅ Supported | Primary development platform |
| macOS | x86_64 | ⚠️ Untested | Should work, needs testing |
| macOS | ARM64 (M1/M2) | ⚠️ Untested | Requires ARM64 runtime libraries |
| Windows | x86_64 | ⚠️ Untested | Requires .dll files |

## Troubleshooting

### "cannot find -ltokenizers"
**Cause**: Tokenizers library not found at compile time
**Solution**: Set `CGO_LDFLAGS="-L$(pwd)/lib"` when building

### "onnxruntime.so: cannot open shared object file"
**Cause**: ONNXRuntime library not found at runtime
**Solution**:
- Bundle libraries in `lib/` directory
- Set `LD_LIBRARY_PATH` to include `lib/` directory
- Or install system-wide: `/usr/local/lib/`

### "API version [22] is not available"
**Cause**: Wrong ONNXRuntime version
**Solution**: Use version 1.22.0 (matches Go binding requirements)

### "failed to download model"
**Cause**: Network connectivity or URL incorrect
**Solution**:
- Check internet connection
- Verify URL: https://downloads.poiesic.com/wonda/gtr-t5-base-onnx-1.0.0.tar.gz
- Or manually download and extract to `~/.config/wonda/models/`

## Updating the Model

To release a new model version:

1. **Update version in code**:
   ```go
   // internal/memory/model_downloader.go
   const ModelVersion = "1.1.0"
   const DefaultModelURL = "https://downloads.poiesic.com/wonda/gtr-t5-base-onnx-1.1.0.tar.gz"
   ```

2. **Re-export model** (if needed):
   ```bash
   ./scripts/run_export.sh
   ```

3. **Update metadata**:
   ```json
   {
     "version": "1.1.0",
     ...
   }
   ```

4. **Package and upload**:
   ```bash
   tar -czf gtr-t5-base-onnx-1.1.0.tar.gz gtr-t5-base-onnx/
   # Upload to downloads.poiesic.com/wonda/
   ```

## Performance

**First run**:
- Model download: ~30 seconds (depending on network)
- Model extraction: ~2 seconds
- Initialization: ~1 second

**Subsequent runs**:
- Model load: ~1 second (from cache)
- Inference: ~50-100ms per embedding (CPU)

**Memory usage**:
- Model in RAM: ~500MB
- Runtime overhead: ~50MB

## Cognitive Distortions

The gtr-t5-base model is vec2text-compatible, enabling vector arithmetic for cognitive distortions:

```python
# Example: Delusional shift
reality_vector = embed("my mother was drunk and abusive")
delusion_vector = embed("my mom was sensitive but loving")

shift_magnitude = cosine_distance(reality_vector, delusion_vector)
distorted = reality_vector + (delusion_vector - reality_vector) * distortion_factor

distorted_text = vec2text.invert(distorted)
# "my mother had a unique parenting style..."
```

This technique is planned for Phase 3 of Wonda development.

## References

- [ONNX Runtime](https://github.com/microsoft/onnxruntime)
- [yalue/onnxruntime_go](https://github.com/yalue/onnxruntime_go)
- [HuggingFace Tokenizers](https://github.com/huggingface/tokenizers)
- [daulet/tokenizers](https://github.com/daulet/tokenizers)
- [gtr-t5-base model](https://huggingface.co/sentence-transformers/gtr-t5-base)
- [vec2text](https://github.com/jxmorris12/vec2text)
