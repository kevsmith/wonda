#!/usr/bin/env python3
"""
Export sentence-transformers/gtr-t5-base to ONNX format.

Requirements:
    pip install sentence-transformers optimum[onnxruntime] onnx transformers torch
"""

import os
import json
import torch
from pathlib import Path
from sentence_transformers import SentenceTransformer
from transformers import AutoTokenizer, AutoModel

def export_to_onnx(output_dir="./gtr-t5-base-onnx"):
    """Export gtr-t5-base to ONNX format using manual export."""
    model_name = "sentence-transformers/gtr-t5-base"
    output_path = Path(output_dir)
    output_path.mkdir(parents=True, exist_ok=True)

    print(f"Exporting {model_name} to ONNX...")
    print(f"Output directory: {output_path.absolute()}")

    # Load the model and tokenizer
    print("Loading model and tokenizer...")
    model = AutoModel.from_pretrained(model_name)
    tokenizer = AutoTokenizer.from_pretrained(model_name)

    # Set to eval mode
    model.eval()

    # Create dummy input for export
    print("Creating dummy input...")
    dummy_text = "This is a test sentence for export."
    inputs = tokenizer(dummy_text, return_tensors="pt", padding=True, truncation=True, max_length=512)

    # Export encoder to ONNX
    print("Exporting encoder to ONNX format...")
    onnx_path = output_path / "model.onnx"

    with torch.no_grad():
        torch.onnx.export(
            model.encoder,  # Export only the encoder
            (inputs['input_ids'], inputs['attention_mask']),
            str(onnx_path),
            input_names=['input_ids', 'attention_mask'],
            output_names=['last_hidden_state'],
            dynamic_axes={
                'input_ids': {0: 'batch_size', 1: 'sequence_length'},
                'attention_mask': {0: 'batch_size', 1: 'sequence_length'},
                'last_hidden_state': {0: 'batch_size', 1: 'sequence_length'}
            },
            opset_version=14,
            do_constant_folding=True,
        )

    # Save tokenizer
    print("Saving tokenizer...")
    tokenizer.save_pretrained(output_path)

    # Create metadata file for Wonda
    metadata = {
        "model_name": model_name,
        "version": "1.0.0",
        "dimensions": 768,
        "max_tokens": 512,
        "format": "onnx",
        "description": "GTR-T5-Base embedding model for vec2text compatibility",
        "vec2text_compatible": True
    }

    with open(output_path / "metadata.json", "w") as f:
        json.dump(metadata, f, indent=2)

    print(f"\n✓ Export complete!")
    print(f"  Model saved to: {output_path}")
    print(f"  Files created:")
    for f in output_path.iterdir():
        print(f"    - {f.name}")

if __name__ == "__main__":
    export_to_onnx()
