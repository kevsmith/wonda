#!/usr/bin/env python3
"""
Verify that the ONNX export produces embeddings matching the original model.
"""

import numpy as np
from pathlib import Path
from sentence_transformers import SentenceTransformer
from optimum.onnxruntime import ORTModelForFeatureExtraction
from transformers import AutoTokenizer
import torch

def mean_pooling(model_output, attention_mask):
    """Mean pooling to get sentence embedding."""
    token_embeddings = model_output[0]
    input_mask_expanded = attention_mask.unsqueeze(-1).expand(token_embeddings.size()).float()
    return torch.sum(token_embeddings * input_mask_expanded, 1) / torch.clamp(input_mask_expanded.sum(1), min=1e-9)

def verify_onnx_export(onnx_dir="./gtr-t5-base-onnx"):
    """Compare ONNX embeddings with original model embeddings."""
    model_name = "sentence-transformers/gtr-t5-base"
    onnx_path = Path(onnx_dir)

    if not onnx_path.exists():
        print(f"Error: ONNX directory not found: {onnx_path}")
        print("Run export script first: ./scripts/run_export.sh")
        return False

    print("Loading original sentence-transformer model...")
    original_model = SentenceTransformer(model_name)

    print("Loading ONNX model...")
    onnx_model = ORTModelForFeatureExtraction.from_pretrained(onnx_path)
    tokenizer = AutoTokenizer.from_pretrained(onnx_path)

    # Test sentences
    test_sentences = [
        "Hello, world!",
        "The quick brown fox jumps over the lazy dog.",
        "My mother was a sensitive but loving person.",
        "Artificial intelligence is transforming technology.",
    ]

    print("\nTesting embeddings...")
    print("-" * 80)

    all_match = True
    for i, sentence in enumerate(test_sentences, 1):
        print(f"\n[{i}/{len(test_sentences)}] Testing: '{sentence}'")

        # Original embedding
        orig_embedding = original_model.encode(sentence, convert_to_numpy=True)

        # ONNX embedding
        encoded = tokenizer(sentence, return_tensors="pt", padding=True, truncation=True, max_length=512)
        with torch.no_grad():
            # For T5 encoder-only models, we may need to handle this differently
            try:
                onnx_output = onnx_model(**encoded)
            except Exception as e:
                print(f"  ✗ Error calling ONNX model: {e}")
                print(f"  Trying encoder-only approach...")
                # Some exports need decoder_input_ids even for encoder-only
                encoded['decoder_input_ids'] = torch.zeros((1, 1), dtype=torch.long)
                onnx_output = onnx_model(**encoded)

        onnx_embedding = mean_pooling(onnx_output, encoded['attention_mask'])
        onnx_embedding = onnx_embedding.squeeze().numpy()

        # Compare
        cosine_sim = np.dot(orig_embedding, onnx_embedding) / (
            np.linalg.norm(orig_embedding) * np.linalg.norm(onnx_embedding)
        )
        l2_distance = np.linalg.norm(orig_embedding - onnx_embedding)

        print(f"  Dimensions: {orig_embedding.shape}")
        print(f"  Cosine similarity: {cosine_sim:.6f}")
        print(f"  L2 distance: {l2_distance:.6f}")

        # Check if embeddings are close enough
        if cosine_sim > 0.999:
            print(f"  ✓ PASS - Embeddings match!")
        else:
            print(f"  ✗ FAIL - Embeddings differ significantly!")
            all_match = False

    print("\n" + "=" * 80)
    if all_match:
        print("✓ All tests passed! ONNX export is correct.")
        print(f"\nYou can now package and upload {onnx_path} to your Digital Ocean server.")
        return True
    else:
        print("✗ Some tests failed. Check the ONNX export.")
        return False

if __name__ == "__main__":
    success = verify_onnx_export()
    exit(0 if success else 1)
