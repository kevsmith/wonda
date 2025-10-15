#!/bin/bash
# Verify the ONNX export in the virtualenv

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VENV_DIR="$SCRIPT_DIR/venv"

# Check if virtualenv exists
if [ ! -d "$VENV_DIR" ]; then
    echo "Virtual environment not found. Running setup..."
    "$SCRIPT_DIR/setup_export_env.sh"
fi

# Activate virtualenv
source "$VENV_DIR/bin/activate"

# Run verification script
echo "Running ONNX verification..."
python "$SCRIPT_DIR/verify_onnx_export.py" "$@"

# Deactivate
deactivate

echo ""
echo "Virtual environment deactivated."
