#!/bin/bash
# Run the ONNX export script in the virtualenv

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

# Run export script
echo "Running ONNX export..."
python "$SCRIPT_DIR/export_gtr_t5_to_onnx.py" "$@"

# Deactivate
deactivate

echo ""
echo "✓ Export complete! Virtual environment deactivated."
