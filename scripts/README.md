# ONNX Model Export Scripts

Scripts for exporting sentence-transformers/gtr-t5-base to ONNX format for in-process embeddings.

## Quick Start

```bash
# 1. Export the model to ONNX
./scripts/run_export.sh

# 2. Verify the export
./scripts/run_verify.sh

# 3. Package for distribution (after verification passes)
tar -czf gtr-t5-base-onnx.tar.gz gtr-t5-base-onnx/
```

## Scripts

### `setup_export_env.sh`
Creates a Python virtual environment and installs dependencies:
- sentence-transformers
- optimum[onnxruntime]
- onnx
- transformers

**Note:** This script is automatically called by `run_export.sh` and `run_verify.sh` if the venv doesn't exist.

### `run_export.sh`
Exports gtr-t5-base to ONNX format in `./gtr-t5-base-onnx/` directory.

Creates:
- `model.onnx` - ONNX model file
- `tokenizer.json` - Tokenizer configuration
- `config.json` - Model configuration
- `metadata.json` - Wonda-specific metadata

### `run_verify.sh`
Verifies that ONNX embeddings match the original model by:
- Loading both original and ONNX models
- Generating embeddings for test sentences
- Comparing via cosine similarity and L2 distance
- Reports PASS/FAIL for each test

### `export_gtr_t5_to_onnx.py`
Python script that performs the ONNX export (called by run_export.sh).

### `verify_onnx_export.py`
Python script that verifies the export (called by run_verify.sh).

## Output

After successful export and verification:

```
gtr-t5-base-onnx/
├── model.onnx              # ONNX model (~270MB)
├── tokenizer.json          # Tokenizer config
├── config.json             # Model config
├── metadata.json           # Wonda metadata (dimensions, vec2text compatible, etc.)
└── ... (other files)
```

## Uploading to Digital Ocean

After verification passes:

```bash
# Create tarball
tar -czf gtr-t5-base-onnx.tar.gz gtr-t5-base-onnx/

# Upload to your server
scp gtr-t5-base-onnx.tar.gz your-server:/path/to/downloads/

# Make available at URL
# https://your-domain.com/models/gtr-t5-base-onnx.tar.gz
```

## Why ONNX?

- **Vec2text compatible**: gtr-t5-base works with vec2text for cognitive distortion features
- **In-process**: No external services needed (Ollama, LM Studio, etc.)
- **Cross-platform**: Works on Linux, macOS, Windows via Go ONNX runtime
- **Deterministic**: Consistent embeddings across platforms

## Requirements

- Python 3.8+
- ~500MB free disk space for model export
- Internet connection for first-time dependency download
