# ZeroDB Documentation

Welcome to the ZeroDB documentation! This directory contains comprehensive guides for all ZeroDB features.

## Available Features

### 1. Quantum Features (TASK-054)

Advanced vector manipulation and search capabilities using quantum-inspired algorithms.

**Documentation:**
- **[Quantum Features Guide](quantum-features.md)** - Comprehensive 530-line guide covering all aspects
- **[Quick Reference](quantum-quick-reference.md)** - 226-line quick reference for rapid lookup
- **[Practical Examples](quantum-examples.md)** - 628 lines of real-world, runnable examples

**Commands:**
- `ainative-code zerodb quantum entangle` - Create quantum correlations between vectors
- `ainative-code zerodb quantum measure` - Analyze quantum state and properties
- `ainative-code zerodb quantum compress` - Reduce vector dimensions efficiently
- `ainative-code zerodb quantum decompress` - Restore compressed vectors
- `ainative-code zerodb quantum search` - Perform quantum-enhanced similarity search

**Key Capabilities:**
- Vector entanglement for knowledge graphs
- Quantum state measurement and analysis
- Intelligent vector compression (30-70% savings)
- Quantum-boosted similarity search
- Entanglement-aware search results

### 2. Vector Operations (TASK-051)

Standard vector database operations for embeddings and similarity search.

**Commands:**
- `ainative-code zerodb vector create-collection` - Create vector collections
- `ainative-code zerodb vector insert` - Insert vector embeddings
- `ainative-code zerodb vector search` - Similarity search
- `ainative-code zerodb vector delete` - Delete vectors
- `ainative-code zerodb vector list-collections` - List all collections

### 3. NoSQL Tables (TASK-052)

MongoDB-style document database with flexible schemas.

**Commands:**
- `ainative-code zerodb table create` - Create NoSQL tables
- `ainative-code zerodb table insert` - Insert documents
- `ainative-code zerodb table query` - Query with filters
- `ainative-code zerodb table update` - Update documents
- `ainative-code zerodb table delete` - Delete documents
- `ainative-code zerodb table list` - List all tables

### 4. Agent Memory (TASK-053)

Long-term memory storage for AI agents with semantic search.

**Commands:**
- `ainative-code zerodb memory store` - Store agent memories
- `ainative-code zerodb memory retrieve` - Semantic memory search
- `ainative-code zerodb memory list` - List all memories
- `ainative-code zerodb memory clear` - Clear memories

## Getting Started

### Quick Start - Quantum Features

```bash
# 1. Entangle two related vectors
ainative-code zerodb quantum entangle \
  --vector-id-1 vec_doc_1 \
  --vector-id-2 vec_doc_2

# 2. Measure a vector's quantum state
ainative-code zerodb quantum measure --vector-id vec_doc_1

# 3. Compress a vector to save storage
ainative-code zerodb quantum compress \
  --vector-id vec_large \
  --compression-ratio 0.5

# 4. Perform quantum-enhanced search
ainative-code zerodb quantum search \
  --query-vector '[0.1,0.2,0.3,0.4,0.5]' \
  --use-quantum-boost \
  --limit 10
```

### Configuration

Set your ZeroDB project ID:

```bash
# Via environment variable
export AINATIVE_CODE_ZERODB_PROJECT_ID=your-project-id

# Or via config file (~/.ainative-code.yaml)
zerodb:
  project_id: your-project-id
  base_url: https://api.ainative.studio
```

## Documentation Structure

```
docs/zerodb/
├── README.md                      # This file - overview and navigation
├── quantum-features.md            # Comprehensive quantum features guide
├── quantum-quick-reference.md     # Quick reference for quantum commands
└── quantum-examples.md            # Practical, runnable examples
```

## Common Use Cases

### Use Case 1: Building a Knowledge Graph

Use quantum entanglement to create semantic relationships:

```bash
# Entangle related concepts
ainative-code zerodb quantum entangle \
  --vector-id-1 vec_ml \
  --vector-id-2 vec_neural_networks

# Search with entanglement awareness
ainative-code zerodb quantum search \
  --query-vector '[...]' \
  --include-entangled
```

**See:** [quantum-examples.md](quantum-examples.md) Example 4-5

### Use Case 2: Optimizing Storage Costs

Use quantum compression to reduce storage:

```bash
# Measure entropy first
ainative-code zerodb quantum measure --vector-id vec_old

# Compress if entropy is low
ainative-code zerodb quantum compress \
  --vector-id vec_old \
  --compression-ratio 0.5
```

**See:** [quantum-examples.md](quantum-examples.md) Example 6-7

### Use Case 3: Enhanced Search Quality

Use quantum boost for better results:

```bash
# Standard search
ainative-code zerodb quantum search \
  --query-vector '[...]'

# With quantum boost
ainative-code zerodb quantum search \
  --query-vector '[...]' \
  --use-quantum-boost
```

**See:** [quantum-examples.md](quantum-examples.md) Example 8-9

## Feature Comparison

| Feature | Vector Ops | NoSQL Tables | Agent Memory | Quantum |
|---------|-----------|--------------|--------------|---------|
| Similarity Search | ✓ | - | ✓ | ✓ (Enhanced) |
| Document Storage | - | ✓ | - | - |
| Compression | - | - | - | ✓ |
| Entanglement | - | - | - | ✓ |
| State Analysis | - | - | - | ✓ |
| Metadata Filtering | ✓ | ✓ | ✓ | ✓ |
| JSON Output | ✓ | ✓ | ✓ | ✓ |

## Performance Tips

1. **Use Compression for Old Data**: Compress vectors older than 30-90 days
2. **Entangle Strategically**: Don't entangle everything, focus on meaningful relationships
3. **Enable Quantum Boost Selectively**: Use for important queries where quality matters
4. **Measure Before Compressing**: Check entropy to ensure good compression
5. **Cache Search Results**: Cache frequently used quantum search results

## Best Practices

### Quantum Entanglement
- Be selective: Don't entangle everything
- Document relationships: Track why vectors are entangled
- Monitor graph size: Large graphs can impact performance
- Use hierarchies: Different entanglement levels for different relationships

### Quantum Compression
- Measure first: Always check entropy before compressing
- Test impact: Compress a sample and test search quality
- Use tiers: Compress based on access patterns (hot/warm/cold)
- Monitor quality: Track information loss and search accuracy

### Quantum Search
- Start simple: Begin with basic search, add features as needed
- A/B test: Compare quantum boost vs standard search
- Use filters: Combine quantum features with metadata filtering
- Cache results: Cache frequently used results

## Troubleshooting

### Common Issues

**Poor search quality after compression**
- Check information loss percentage (should be <5%)
- Try higher compression ratio (e.g., 0.6 instead of 0.4)
- Measure entropy - high entropy vectors compress poorly

**Entanglement not affecting search**
- Verify vectors are entangled (use measure command)
- Check `--include-entangled` flag is set
- Ensure entangled vectors have sufficient similarity

**Quantum boost not improving results**
- Measure vector coherence (low coherence = limited benefits)
- Compare classical and quantum similarity scores
- Quantum boost is query-dependent

See [quantum-features.md](quantum-features.md) for detailed troubleshooting guide.

## Resources

### Documentation
- **Comprehensive Guide**: [quantum-features.md](quantum-features.md)
- **Quick Reference**: [quantum-quick-reference.md](quantum-quick-reference.md)
- **Examples**: [quantum-examples.md](quantum-examples.md)

### External Links
- ZeroDB API Documentation: https://api.ainative.studio/docs
- ZeroDB Community: https://community.ainative.studio
- Support: support@ainative.studio

## Version Information

- **Quantum Features**: v1.0.0 (Released: 2026-01-03)
- **Vector Operations**: v1.0.0
- **NoSQL Tables**: v1.0.0
- **Agent Memory**: v1.0.0

## Contributing

For questions, issues, or feature requests related to ZeroDB:
1. Check the documentation first
2. Search existing issues
3. Contact support@ainative.studio
4. Join the community at https://community.ainative.studio

## License

Copyright (c) 2026 AINative Studio. All rights reserved.
