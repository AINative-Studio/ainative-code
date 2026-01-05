# ZeroDB Quantum Features - Practical Examples

This document provides real-world examples and complete workflows for using ZeroDB quantum features.

## Table of Contents

1. [Basic Operations](#basic-operations)
2. [Knowledge Graph Construction](#knowledge-graph-construction)
3. [Storage Optimization](#storage-optimization)
4. [Enhanced Search Workflows](#enhanced-search-workflows)
5. [Production Use Cases](#production-use-cases)

## Basic Operations

### Example 1: Simple Vector Entanglement

Create a relationship between two related product vectors.

```bash
# Step 1: Insert first product vector
ainative-code zerodb vector insert \
  --collection products \
  --vector '[0.23, 0.45, 0.67, 0.12, 0.89]' \
  --metadata '{"product":"laptop","category":"electronics","price":999}'

# Response: vec_laptop_001

# Step 2: Insert related product vector
ainative-code zerodb vector insert \
  --collection products \
  --vector '[0.21, 0.43, 0.69, 0.15, 0.87]' \
  --metadata '{"product":"laptop_charger","category":"accessories","price":49}'

# Response: vec_charger_001

# Step 3: Entangle the related products
ainative-code zerodb quantum entangle \
  --vector-id-1 vec_laptop_001 \
  --vector-id-2 vec_charger_001

# Output:
# Vectors entangled successfully!
# Entanglement ID:     ent_abc123
# Correlation Score:   0.9234
```

### Example 2: Vector State Measurement

Analyze a vector's quantum properties before making optimization decisions.

```bash
# Measure a vector's state
ainative-code zerodb quantum measure --vector-id vec_document_001

# Output:
# Quantum Measurement Results:
#
# Vector ID:         vec_document_001
# Dimension:         1536
# Quantum State:     superposition
# Entropy:           0.652143
# Coherence:         0.891234
# Entangled:         Yes
# Entangled With:    vec_document_002, vec_document_003
# Compression Ratio: 0.00

# Interpretation:
# - Medium entropy (0.65) - good candidate for compression
# - High coherence (0.89) - quantum boost will be effective
# - Already entangled with 2 other vectors
```

### Example 3: Vector Compression

Compress a large vector to save storage costs.

```bash
# Step 1: Measure to check suitability
ainative-code zerodb quantum measure --vector-id vec_large_001

# Step 2: Compress to 50% if entropy is favorable
ainative-code zerodb quantum compress \
  --vector-id vec_large_001 \
  --compression-ratio 0.5

# Output:
# Vector compressed successfully!
#
# Vector ID:             vec_large_001
# Original Dimension:    1536
# Compressed Dimension:  768
# Compression Ratio:     0.50
# Information Loss:      2.34%
# Storage Savings:       50.0%

# Step 3: Verify search quality still acceptable
ainative-code zerodb quantum search \
  --query-vector '[0.1, 0.2, ...]' \
  --limit 5
```

## Knowledge Graph Construction

### Example 4: Building a Tutorial Knowledge Graph

Create interconnected tutorial documents with semantic relationships.

```bash
#!/bin/bash

# Tutorial 1: Getting Started
VEC1=$(ainative-code zerodb vector insert \
  --collection tutorials \
  --vector '[0.12, 0.34, 0.56, ...]' \
  --metadata '{"title":"Getting Started","level":"beginner","topic":"basics"}' \
  --json | jq -r '.id')

# Tutorial 2: Advanced Concepts
VEC2=$(ainative-code zerodb vector insert \
  --collection tutorials \
  --vector '[0.15, 0.37, 0.59, ...]' \
  --metadata '{"title":"Advanced Concepts","level":"advanced","topic":"basics"}' \
  --json | jq -r '.id')

# Tutorial 3: Best Practices
VEC3=$(ainative-code zerodb vector insert \
  --collection tutorials \
  --vector '[0.14, 0.36, 0.58, ...]' \
  --metadata '{"title":"Best Practices","level":"intermediate","topic":"basics"}' \
  --json | jq -r '.id')

# Create learning path: Getting Started -> Best Practices
ainative-code zerodb quantum entangle \
  --vector-id-1 $VEC1 \
  --vector-id-2 $VEC3

# Create learning path: Best Practices -> Advanced Concepts
ainative-code zerodb quantum entangle \
  --vector-id-1 $VEC3 \
  --vector-id-2 $VEC2

# Now searches will surface related tutorials automatically
ainative-code zerodb quantum search \
  --query-vector '[0.13, 0.35, 0.57, ...]' \
  --include-entangled \
  --limit 10
```

### Example 5: Product Recommendation Graph

Build a product recommendation system using entanglement.

```bash
#!/bin/bash

# Define products and their relationships
declare -A products=(
  ["laptop"]='{"id":"vec_laptop","vector":"[0.2,0.4,0.6]","price":999}'
  ["monitor"]='{"id":"vec_monitor","vector":"[0.19,0.41,0.61]","price":299}'
  ["keyboard"]='{"id":"vec_keyboard","vector":"[0.21,0.39,0.59]","price":79}'
  ["mouse"]='{"id":"vec_mouse","vector":"[0.22,0.38,0.58]","price":49}'
)

# Insert products
for product in "${!products[@]}"; do
  ainative-code zerodb vector insert \
    --collection products \
    --vector "$(echo ${products[$product]} | jq -r '.vector')" \
    --metadata "{\"product\":\"$product\"}"
done

# Entangle frequently-bought-together items
# Laptop + Monitor
ainative-code zerodb quantum entangle \
  --vector-id-1 vec_laptop \
  --vector-id-2 vec_monitor

# Laptop + Keyboard
ainative-code zerodb quantum entangle \
  --vector-id-1 vec_laptop \
  --vector-id-2 vec_keyboard

# Laptop + Mouse
ainative-code zerodb quantum entangle \
  --vector-id-1 vec_laptop \
  --vector-id-2 vec_mouse

# Keyboard + Mouse (often bought together)
ainative-code zerodb quantum entangle \
  --vector-id-1 vec_keyboard \
  --vector-id-2 vec_mouse

# Search for laptop will now show related accessories
ainative-code zerodb quantum search \
  --query-vector '[0.2,0.4,0.6]' \
  --include-entangled \
  --limit 10
```

## Storage Optimization

### Example 6: Tiered Storage Strategy

Implement hot/warm/cold storage tiers using compression.

```bash
#!/bin/bash

# Tier 1: Hot data (accessed frequently) - No compression
# Tier 2: Warm data (accessed occasionally) - 30% compression
# Tier 3: Cold data (archived) - 50% compression

# Function to compress warm data
compress_warm_data() {
  local vector_id=$1
  echo "Compressing warm data: $vector_id"

  ainative-code zerodb quantum compress \
    --vector-id "$vector_id" \
    --compression-ratio 0.7
}

# Function to compress cold data
compress_cold_data() {
  local vector_id=$1
  echo "Compressing cold data: $vector_id"

  ainative-code zerodb quantum compress \
    --vector-id "$vector_id" \
    --compression-ratio 0.5
}

# Example: Move data to tiers based on age
# Vectors older than 30 days -> warm
# Vectors older than 90 days -> cold

# Warm tier (example vector IDs)
for vec_id in vec_30day_001 vec_30day_002 vec_30day_003; do
  compress_warm_data "$vec_id"
done

# Cold tier (example vector IDs)
for vec_id in vec_90day_001 vec_90day_002; do
  compress_cold_data "$vec_id"
done

# Calculate total savings
echo "Total storage savings: ~40% for warm data, ~50% for cold data"
```

### Example 7: Selective Compression Based on Entropy

Compress only vectors that are good candidates.

```bash
#!/bin/bash

# Function to check and compress if suitable
compress_if_suitable() {
  local vector_id=$1

  # Measure entropy
  local entropy=$(ainative-code zerodb quantum measure \
    --vector-id "$vector_id" \
    --json | jq -r '.entropy')

  # Only compress if entropy is low (< 0.7)
  if (( $(echo "$entropy < 0.7" | bc -l) )); then
    echo "Vector $vector_id has low entropy ($entropy) - compressing"

    ainative-code zerodb quantum compress \
      --vector-id "$vector_id" \
      --compression-ratio 0.5
  else
    echo "Vector $vector_id has high entropy ($entropy) - skipping compression"
  fi
}

# Process a batch of vectors
for vec_id in vec_001 vec_002 vec_003 vec_004; do
  compress_if_suitable "$vec_id"
done
```

## Enhanced Search Workflows

### Example 8: Progressive Search Enhancement

Start with basic search and progressively enhance if needed.

```bash
#!/bin/bash

QUERY_VECTOR='[0.1, 0.2, 0.3, 0.4, 0.5]'

echo "Step 1: Basic quantum search"
ainative-code zerodb quantum search \
  --query-vector "$QUERY_VECTOR" \
  --limit 5

echo -e "\nStep 2: Enable quantum boost for better ranking"
ainative-code zerodb quantum search \
  --query-vector "$QUERY_VECTOR" \
  --limit 5 \
  --use-quantum-boost

echo -e "\nStep 3: Include entangled vectors for more context"
ainative-code zerodb quantum search \
  --query-vector "$QUERY_VECTOR" \
  --limit 10 \
  --use-quantum-boost \
  --include-entangled

echo -e "\nStep 4: Export results for analysis"
ainative-code zerodb quantum search \
  --query-vector "$QUERY_VECTOR" \
  --limit 10 \
  --use-quantum-boost \
  --include-entangled \
  --json > search_results.json
```

### Example 9: A/B Testing Search Quality

Compare standard vs quantum-enhanced search.

```bash
#!/bin/bash

QUERY='[0.25, 0.35, 0.45, 0.55, 0.65]'

# Standard search
echo "=== Standard Search ==="
ainative-code zerodb vector search \
  --collection documents \
  --query-vector "$QUERY" \
  --limit 10 \
  --json > standard_results.json

# Quantum-enhanced search
echo -e "\n=== Quantum Search ==="
ainative-code zerodb quantum search \
  --query-vector "$QUERY" \
  --limit 10 \
  --use-quantum-boost \
  --json > quantum_results.json

# Compare results
echo -e "\n=== Comparison ==="
echo "Standard results: $(jq 'length' standard_results.json) vectors"
echo "Quantum results: $(jq 'length' quantum_results.json) vectors"

# Analyze similarity scores
echo -e "\nTop similarity scores:"
echo "Standard: $(jq '.[0].similarity' standard_results.json)"
echo "Quantum: $(jq '.[0].similarity' quantum_results.json)"
```

## Production Use Cases

### Example 10: Document Search with Context

Build a document search system with automatic context expansion.

```bash
#!/bin/bash

# Step 1: Index documents with embeddings
index_document() {
  local title=$1
  local vector=$2
  local category=$3

  ainative-code zerodb vector insert \
    --collection documents \
    --vector "$vector" \
    --metadata "{\"title\":\"$title\",\"category\":\"$category\"}"
}

# Index documents
index_document "Introduction to AI" "[0.1,0.2,0.3,...]" "tutorial"
index_document "AI Best Practices" "[0.11,0.21,0.31,...]" "guide"
index_document "Advanced AI Techniques" "[0.12,0.22,0.32,...]" "advanced"

# Step 2: Entangle related documents
ainative-code zerodb quantum entangle \
  --vector-id-1 vec_intro_ai \
  --vector-id-2 vec_best_practices

ainative-code zerodb quantum entangle \
  --vector-id-1 vec_best_practices \
  --vector-id-2 vec_advanced_techniques

# Step 3: Search with context expansion
search_with_context() {
  local query=$1

  ainative-code zerodb quantum search \
    --query-vector "$query" \
    --include-entangled \
    --use-quantum-boost \
    --limit 20 \
    --json
}

# Execute search
search_with_context "[0.1,0.2,0.3,...]" | \
  jq '.[] | {title: .vector.metadata.title, similarity: .similarity}'
```

### Example 11: Multi-Language Support with Compression

Store multiple language embeddings efficiently.

```bash
#!/bin/bash

# Insert and compress non-English documents
index_and_compress() {
  local title=$1
  local vector=$2
  local lang=$3

  # Insert vector
  local vec_id=$(ainative-code zerodb vector insert \
    --collection documents \
    --vector "$vector" \
    --metadata "{\"title\":\"$title\",\"lang\":\"$lang\"}" \
    --json | jq -r '.id')

  # Compress non-English documents more aggressively
  if [ "$lang" != "en" ]; then
    ainative-code zerodb quantum compress \
      --vector-id "$vec_id" \
      --compression-ratio 0.5
  fi

  echo "$vec_id"
}

# Index documents in multiple languages
VEC_EN=$(index_and_compress "Getting Started" "[0.1,0.2,...]" "en")
VEC_ES=$(index_and_compress "Empezando" "[0.11,0.21,...]" "es")
VEC_FR=$(index_and_compress "Commencer" "[0.12,0.22,...]" "fr")

# Entangle translated versions
ainative-code zerodb quantum entangle --vector-id-1 $VEC_EN --vector-id-2 $VEC_ES
ainative-code zerodb quantum entangle --vector-id-1 $VEC_EN --vector-id-2 $VEC_FR

# Search can now find translations automatically
ainative-code zerodb quantum search \
  --query-vector "[0.1,0.2,...]" \
  --include-entangled \
  --limit 10
```

### Example 12: Monitoring and Maintenance Script

Regular maintenance tasks for quantum features.

```bash
#!/bin/bash

# Monitor and maintain quantum features
echo "=== ZeroDB Quantum Maintenance ==="
echo "Date: $(date)"

# Function to check vector health
check_vector_health() {
  local vector_id=$1

  local result=$(ainative-code zerodb quantum measure \
    --vector-id "$vector_id" \
    --json)

  local entropy=$(echo "$result" | jq -r '.entropy')
  local coherence=$(echo "$result" | jq -r '.coherence')
  local entangled=$(echo "$result" | jq -r '.vector.is_entangled')
  local compressed=$(echo "$result" | jq -r '.vector.compression_ratio')

  echo "Vector: $vector_id"
  echo "  Entropy: $entropy"
  echo "  Coherence: $coherence"
  echo "  Entangled: $entangled"
  echo "  Compressed: $compressed"
  echo ""
}

# Check all vectors (example)
for vec_id in vec_001 vec_002 vec_003; do
  check_vector_health "$vec_id"
done

# Identify compression candidates
echo "=== Compression Candidates ==="
echo "Vectors with entropy < 0.7 and not yet compressed:"

# (In production, query all vectors and filter)
# This is a simplified example

# Recommend optimizations
echo -e "\n=== Recommendations ==="
echo "1. Compress 3 vectors with low entropy (save ~1.2 GB)"
echo "2. Review 5 entanglement relationships for accuracy"
echo "3. Consider decompressing 1 critical vector"
```

## Advanced Patterns

### Example 13: Hierarchical Knowledge Structure

Create multi-level knowledge hierarchies.

```bash
#!/bin/bash

# Level 1: Categories
cat_ai=$(ainative-code zerodb vector insert \
  --collection knowledge \
  --vector "[0.1,0.2,...]" \
  --metadata '{"type":"category","name":"AI"}' \
  --json | jq -r '.id')

cat_ml=$(ainative-code zerodb vector insert \
  --collection knowledge \
  --vector "[0.11,0.21,...]" \
  --metadata '{"type":"category","name":"Machine Learning"}' \
  --json | jq -r '.id')

# Level 2: Topics
topic_nlp=$(ainative-code zerodb vector insert \
  --collection knowledge \
  --vector "[0.12,0.22,...]" \
  --metadata '{"type":"topic","name":"NLP"}' \
  --json | jq -r '.id')

topic_cv=$(ainative-code zerodb vector insert \
  --collection knowledge \
  --vector "[0.13,0.23,...]" \
  --metadata '{"type":"topic","name":"Computer Vision"}' \
  --json | jq -r '.id')

# Entangle hierarchies
# AI -> ML (parent-child)
ainative-code zerodb quantum entangle \
  --vector-id-1 "$cat_ai" \
  --vector-id-2 "$cat_ml"

# ML -> NLP (parent-child)
ainative-code zerodb quantum entangle \
  --vector-id-1 "$cat_ml" \
  --vector-id-2 "$topic_nlp"

# ML -> CV (parent-child)
ainative-code zerodb quantum entangle \
  --vector-id-1 "$cat_ml" \
  --vector-id-2 "$topic_cv"

# NLP <-> CV (sibling relationship)
ainative-code zerodb quantum entangle \
  --vector-id-1 "$topic_nlp" \
  --vector-id-2 "$topic_cv"
```

## Testing and Validation

### Example 14: Compression Impact Testing

Test the impact of different compression ratios.

```bash
#!/bin/bash

VECTOR_ID="vec_test_001"
QUERY="[0.1, 0.2, 0.3, 0.4, 0.5]"

# Baseline: measure original vector
echo "=== Baseline (No Compression) ==="
ainative-code zerodb quantum search \
  --query-vector "$QUERY" \
  --limit 5 \
  --json > baseline.json

# Test different compression ratios
for ratio in 0.7 0.5 0.3; do
  echo -e "\n=== Testing Compression Ratio: $ratio ==="

  # Compress
  ainative-code zerodb quantum compress \
    --vector-id "$VECTOR_ID" \
    --compression-ratio "$ratio"

  # Test search
  ainative-code zerodb quantum search \
    --query-vector "$QUERY" \
    --limit 5 \
    --json > "compressed_${ratio}.json"

  # Compare similarity scores
  baseline_sim=$(jq '.[0].similarity' baseline.json)
  compressed_sim=$(jq '.[0].similarity' "compressed_${ratio}.json")

  echo "Baseline similarity: $baseline_sim"
  echo "Compressed similarity: $compressed_sim"

  # Decompress for next test
  ainative-code zerodb quantum decompress --vector-id "$VECTOR_ID"
done

echo -e "\n=== Test Complete ==="
echo "Review compressed_*.json files for detailed results"
```

## Summary

These examples demonstrate:

1. **Basic Operations**: Simple entanglement, measurement, compression
2. **Knowledge Graphs**: Building interconnected vector relationships
3. **Storage Optimization**: Tiered storage and selective compression
4. **Enhanced Search**: Progressive search enhancement and A/B testing
5. **Production Use Cases**: Real-world document search and multi-language support
6. **Maintenance**: Monitoring and optimization scripts
7. **Advanced Patterns**: Hierarchical structures and testing

For more information, see:
- Full documentation: `docs/zerodb/quantum-features.md`
- Quick reference: `docs/zerodb/quantum-quick-reference.md`
