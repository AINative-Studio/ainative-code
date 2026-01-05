# ZeroDB Quantum Features - Quick Reference

## Command Overview

| Command | Purpose | Example |
|---------|---------|---------|
| `quantum entangle` | Link two vectors | `ainative-code zerodb quantum entangle --vector-id-1 vec_1 --vector-id-2 vec_2` |
| `quantum measure` | Analyze vector state | `ainative-code zerodb quantum measure --vector-id vec_123` |
| `quantum compress` | Reduce dimensions | `ainative-code zerodb quantum compress --vector-id vec_123 --compression-ratio 0.5` |
| `quantum decompress` | Restore dimensions | `ainative-code zerodb quantum decompress --vector-id vec_123` |
| `quantum search` | Enhanced search | `ainative-code zerodb quantum search --query-vector '[0.1,0.2,0.3]' --limit 10` |

## Quick Examples

### Entangle Related Vectors

```bash
# Link two related product vectors
ainative-code zerodb quantum entangle \
  --vector-id-1 vec_laptop \
  --vector-id-2 vec_charger
```

### Check Vector State

```bash
# Measure quantum properties
ainative-code zerodb quantum measure --vector-id vec_123

# Output shows: entropy, coherence, entanglement status
```

### Compress for Storage Savings

```bash
# Compress to 50% of original size
ainative-code zerodb quantum compress \
  --vector-id vec_large \
  --compression-ratio 0.5

# Results in 50% storage reduction
```

### Restore Full Dimensions

```bash
# Decompress when full resolution needed
ainative-code zerodb quantum decompress --vector-id vec_compressed
```

### Enhanced Search

```bash
# Basic quantum search
ainative-code zerodb quantum search \
  --query-vector '[0.1,0.2,0.3,0.4,0.5]' \
  --limit 10

# With quantum boost
ainative-code zerodb quantum search \
  --query-vector '[0.1,0.2,0.3]' \
  --use-quantum-boost

# Include entangled vectors
ainative-code zerodb quantum search \
  --query-vector '[0.1,0.2,0.3]' \
  --include-entangled

# All features
ainative-code zerodb quantum search \
  --query-vector '[0.1,0.2,0.3]' \
  --use-quantum-boost \
  --include-entangled \
  --limit 20
```

## Compression Ratio Guide

| Ratio | Use Case | Information Loss | Recommendation |
|-------|----------|------------------|----------------|
| 0.7 | Critical data | Minimal (<2%) | Production vectors |
| 0.5 | General purpose | Low (2-5%) | Most use cases |
| 0.3 | Archival | Moderate (5-10%) | Cold storage |

## When to Use Each Feature

### Entanglement
- Building knowledge graphs
- Creating semantic relationships
- Linking related content
- Enhancing search context

### Measurement
- Assessing vector quality
- Deciding on compression
- Debugging search issues
- Monitoring vector health

### Compression
- Reducing storage costs
- Improving search speed
- Edge deployment
- Archiving old data

### Quantum Search
- Improved search accuracy
- Finding related content
- Research and exploration
- Context-aware results

## Common Patterns

### Pattern 1: Compress Old Data

```bash
# 1. Check if suitable for compression
ainative-code zerodb quantum measure --vector-id vec_old

# 2. If entropy < 0.7, compress
ainative-code zerodb quantum compress \
  --vector-id vec_old \
  --compression-ratio 0.5
```

### Pattern 2: Build Knowledge Graph

```bash
# 1. Entangle related concepts
ainative-code zerodb quantum entangle \
  --vector-id-1 vec_concept_a \
  --vector-id-2 vec_concept_b

# 2. Search with awareness
ainative-code zerodb quantum search \
  --query-vector '[...]' \
  --include-entangled
```

### Pattern 3: Optimize Search

```bash
# 1. Standard search
ainative-code zerodb quantum search \
  --query-vector '[...]' \
  --limit 10

# 2. If results need improvement, enable boost
ainative-code zerodb quantum search \
  --query-vector '[...]' \
  --use-quantum-boost \
  --limit 10
```

## Flags Reference

### Global Flags

| Flag | Type | Description |
|------|------|-------------|
| `--json` | boolean | Output as JSON |

### Entangle Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--vector-id-1` | string | Yes | First vector ID |
| `--vector-id-2` | string | Yes | Second vector ID |

### Measure Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--vector-id` | string | Yes | Vector ID to measure |

### Compress Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--vector-id` | string | Yes | Vector ID to compress |
| `--compression-ratio` | float | Yes | Ratio (0-1, exclusive) |

### Decompress Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--vector-id` | string | Yes | Vector ID to decompress |

### Search Flags

| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `--query-vector` | string | Yes | - | Query vector as JSON array |
| `--limit` | int | No | 10 | Maximum results |
| `--use-quantum-boost` | boolean | No | false | Enable quantum boost |
| `--include-entangled` | boolean | No | false | Include entangled vectors |

## Troubleshooting Quick Guide

| Problem | Solution |
|---------|----------|
| Poor compression quality | Check entropy first; use higher ratio |
| Search not finding results | Try quantum boost; check entanglement |
| High information loss | Use higher compression ratio (e.g., 0.6) |
| Entanglement not working | Verify with measure command |
| Slow search | Compress vectors to reduce dimensions |

## Performance Tips

1. **Compress rarely-accessed vectors** (storage optimization)
2. **Entangle sparingly** (maintain graph performance)
3. **Use quantum boost selectively** (when quality matters)
4. **Measure before compressing** (avoid poor compression)
5. **Cache search results** (reduce query overhead)

## Next Steps

- Read full documentation: `docs/zerodb/quantum-features.md`
- Try example workflows in the main documentation
- Experiment with different compression ratios
- Build your first knowledge graph with entanglement

## Support

- Documentation: https://docs.ainative.studio/zerodb/quantum
- Support: support@ainative.studio
- Community: https://community.ainative.studio
