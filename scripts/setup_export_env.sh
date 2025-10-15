#!/bin/bash
# Setup Python virtual environment for ONNX export

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VENV_DIR="$SCRIPT_DIR/venv"

echo "Setting up Python virtual environment for ONNX export..."

# Create virtualenv if it doesn't exist
if [ ! -d "$VENV_DIR" ]; then
    echo "Creating virtual environment..."
    python3 -m venv "$VENV_DIR"
fi

# Activate virtualenv
echo "Activating virtual environment..."
source "$VENV_DIR/bin/activate"

# Install dependencies
echo "Installing dependencies..."
pip install --upgrade pip
pip install sentence-transformers optimum[onnxruntime] onnx transformers

echo ""
echo "✓ Setup complete!"
echo ""
echo "To activate the environment manually:"
echo "  source $VENV_DIR/bin/activate"
echo ""
echo "To run the export script:"
echo "  ./scripts/run_export.sh"
