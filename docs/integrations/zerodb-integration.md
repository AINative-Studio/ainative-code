# ZeroDB Integration Guide

## Overview

ZeroDB is AINative's comprehensive cloud database platform that provides vector storage, NoSQL tables, PostgreSQL instances, file storage, and agent memory systems. This guide shows you how to integrate and use ZeroDB within the AINative Code CLI.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Vector Operations](#vector-operations)
3. [NoSQL Tables](#nosql-tables)
4. [PostgreSQL Instances](#postgresql-instances)
5. [File Storage](#file-storage)
6. [Agent Memory](#agent-memory)
7. [Quantum Search](#quantum-search)
8. [Event Streaming](#event-streaming)
9. [Best Practices](#best-practices)
10. [Code Examples](#code-examples)
11. [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites

1. **Authentication**: Login to AINative platform
```bash
ainative-code auth login
```

2. **Environment Setup**: Configure ZeroDB credentials
```bash
export ZERODB_PROJECT_ID="your-project-id"
export ZERODB_API_KEY="your-api-key"
```

3. **Verify Connection**: Test ZeroDB connectivity
```bash
/zerodb-project-info
```

### Your First ZeroDB Operation

```bash
# Store a vector embedding
/zerodb-vector-upsert

# Search for similar vectors
/zerodb-vector-search

# View project statistics
/zerodb-project-stats
```

## Vector Operations

Vector operations allow you to store and search high-dimensional embeddings for semantic search, RAG (Retrieval-Augmented Generation), and similarity matching.

### Storing Vectors

**Basic Vector Upsert:**

```javascript
// Use the MCP tool directly
mcp__ainative-zerodb__zerodb_upsert_vector({
  "vector_id": "doc-001",
  "vector": [0.1, 0.2, 0.3, ...], // 1536 dimensions
  "metadata": {
    "title": "Getting Started Guide",
    "type": "documentation",
    "source": "docs/getting-started.md",
    "created_at": "2024-01-15"
  },
  "namespace": "documentation"
})
```

**Batch Vector Upload:**

```javascript
mcp__ainative-zerodb__zerodb_batch_upsert_vectors({
  "vectors": [
    {
      "vector_id": "doc-001",
      "vector": [...],
      "metadata": {"title": "Guide 1"}
    },
    {
      "vector_id": "doc-002",
      "vector": [...],
      "metadata": {"title": "Guide 2"}
    }
  ],
  "namespace": "documentation"
})
```

### Searching Vectors

**Semantic Search:**

```javascript
mcp__ainative-zerodb__zerodb_search_vectors({
  "query_vector": [0.1, 0.2, 0.3, ...], // Your query embedding
  "top_k": 5,
  "similarity_threshold": 0.7,
  "namespace": "documentation",
  "metadata_filter": {
    "type": "documentation"
  }
})
```

**Response Format:**

```json
{
  "results": [
    {
      "vector_id": "doc-001",
      "similarity_score": 0.95,
      "metadata": {
        "title": "Getting Started Guide",
        "type": "documentation"
      }
    },
    {
      "vector_id": "doc-003",
      "similarity_score": 0.87,
      "metadata": {
        "title": "Advanced Features",
        "type": "documentation"
      }
    }
  ]
}
```

### Managing Vectors

**List Vectors:**

```javascript
mcp__ainative-zerodb__zerodb_list_vectors({
  "namespace": "documentation",
  "limit": 100,
  "offset": 0
})
```

**Get Vector Details:**

```javascript
mcp__ainative-zerodb__zerodb_get_vector({
  "vector_id": "doc-001",
  "namespace": "documentation"
})
```

**Delete Vector:**

```javascript
mcp__ainative-zerodb__zerodb_delete_vector({
  "vector_id": "doc-001",
  "namespace": "documentation"
})
```

### Vector Statistics

```javascript
mcp__ainative-zerodb__zerodb_vector_stats({
  "namespace": "documentation" // Optional
})
```

**Response:**

```json
{
  "total_vectors": 1234,
  "namespaces": ["documentation", "code", "chat-history"],
  "dimension": 1536,
  "storage_used_mb": 45.2,
  "by_namespace": {
    "documentation": 856,
    "code": 234,
    "chat-history": 144
  }
}
```

## NoSQL Tables

NoSQL tables provide flexible document storage with MongoDB-style queries.

### Creating Tables

```javascript
mcp__ainative-zerodb__zerodb_create_table({
  "table_name": "users",
  "schema": {
    "fields": [
      {"name": "id", "type": "string", "required": true},
      {"name": "email", "type": "string", "required": true},
      {"name": "name", "type": "string"},
      {"name": "metadata", "type": "json"},
      {"name": "created_at", "type": "timestamp"}
    ],
    "indexes": [
      {"field": "email", "unique": true},
      {"field": "created_at"}
    ]
  }
})
```

### Inserting Data

**Single Insert:**

```javascript
mcp__ainative-zerodb__zerodb_insert_rows({
  "table_name": "users",
  "rows": [{
    "id": "user-001",
    "email": "john@example.com",
    "name": "John Doe",
    "metadata": {
      "role": "admin",
      "team": "engineering"
    },
    "created_at": "2024-01-15T10:30:00Z"
  }]
})
```

**Bulk Insert:**

```javascript
mcp__ainative-zerodb__zerodb_insert_rows({
  "table_name": "users",
  "rows": [
    {"id": "user-001", "email": "john@example.com", "name": "John Doe"},
    {"id": "user-002", "email": "jane@example.com", "name": "Jane Smith"},
    {"id": "user-003", "email": "bob@example.com", "name": "Bob Johnson"}
  ]
})
```

### Querying Data

**Basic Query:**

```javascript
mcp__ainative-zerodb__zerodb_query_rows({
  "table_name": "users",
  "filter": {
    "metadata.role": "admin"
  },
  "limit": 10,
  "offset": 0
})
```

**Advanced Query with Operators:**

```javascript
mcp__ainative-zerodb__zerodb_query_rows({
  "table_name": "users",
  "filter": {
    "$and": [
      {"metadata.role": {"$in": ["admin", "moderator"]}},
      {"created_at": {"$gte": "2024-01-01T00:00:00Z"}}
    ]
  },
  "sort": {"created_at": -1},
  "limit": 50,
  "projection": ["id", "email", "name", "metadata.role"]
})
```

**Supported Query Operators:**

- `$eq`, `$ne`: Equality and inequality
- `$gt`, `$gte`, `$lt`, `$lte`: Comparison
- `$in`, `$nin`: Array membership
- `$and`, `$or`, `$not`: Logical operators
- `$regex`: Regular expression matching
- `$exists`: Field existence check

### Updating Data

```javascript
mcp__ainative-zerodb__zerodb_update_rows({
  "table_name": "users",
  "filter": {"id": "user-001"},
  "update": {
    "$set": {
      "metadata.role": "user",
      "metadata.last_updated": "2024-01-15T14:30:00Z"
    },
    "$inc": {
      "metadata.login_count": 1
    }
  }
})
```

**Update Operators:**

- `$set`: Set field value
- `$unset`: Remove field
- `$inc`: Increment numeric field
- `$push`: Add to array
- `$pull`: Remove from array
- `$addToSet`: Add unique value to array

### Deleting Data

```javascript
mcp__ainative-zerodb__zerodb_delete_rows({
  "table_name": "users",
  "filter": {
    "created_at": {"$lt": "2023-01-01T00:00:00Z"}
  }
})
```

## PostgreSQL Instances

Provision and manage dedicated PostgreSQL instances with direct SQL access.

### Provisioning an Instance

```javascript
mcp__ainative-zerodb__zerodb_postgres_provision({
  "project_id": "your-project-id",
  "instance_size": "standard-2", // micro-1, standard-2, performance-8, etc.
  "postgres_version": "15",
  "backup_enabled": true,
  "backup_retention_days": 7
})
```

**Instance Sizes:**

| Size | vCPU | RAM | Storage | Price/mo |
|------|------|-----|---------|----------|
| micro-1 | 1 | 1GB | 10GB | $5 |
| standard-2 | 2 | 4GB | 20GB | $25 |
| standard-4 | 4 | 8GB | 40GB | $50 |
| performance-8 | 8 | 16GB | 80GB | $100 |

**Provisioning Time:** Approximately 2-3 minutes

### Checking Instance Status

```javascript
mcp__ainative-zerodb__zerodb_postgres_status({
  "project_id": "your-project-id"
})
```

**Response:**

```json
{
  "status": "active", // provisioning, active, maintenance, error
  "instance_size": "standard-2",
  "postgres_version": "15",
  "health": {
    "cpu_usage_percent": 12.5,
    "memory_usage_percent": 34.2,
    "storage_used_gb": 8.5,
    "active_connections": 12,
    "max_connections": 100
  },
  "uptime_hours": 168
}
```

### Getting Connection Details

```javascript
mcp__ainative-zerodb__zerodb_postgres_connection({
  "project_id": "your-project-id",
  "credential_type": "primary" // primary, readonly, admin
})
```

**Response:**

```json
{
  "connection_string": "postgresql://user:password@host:5432/database",
  "host": "pg-instance-abc123.railway.app",
  "port": 5432,
  "database": "railway",
  "username": "postgres",
  "password": "SecurePassword123",
  "ssl_required": true
}
```

### Connecting to PostgreSQL

**Using psql:**

```bash
psql postgresql://user:password@host:5432/database
```

**Using Python (psycopg2):**

```python
import psycopg2

conn = psycopg2.connect(
    host="pg-instance-abc123.railway.app",
    port=5432,
    database="railway",
    user="postgres",
    password="SecurePassword123",
    sslmode="require"
)

cursor = conn.cursor()
cursor.execute("SELECT version();")
print(cursor.fetchone())
conn.close()
```

**Using Node.js (pg):**

```javascript
const { Client } = require('pg');

const client = new Client({
  connectionString: 'postgresql://user:password@host:5432/database',
  ssl: { rejectUnauthorized: false }
});

await client.connect();
const result = await client.query('SELECT NOW()');
console.log(result.rows);
await client.end();
```

### Viewing Query Logs

```javascript
mcp__ainative-zerodb__zerodb_postgres_logs({
  "project_id": "your-project-id",
  "limit": 50,
  "query_type": "SELECT", // Optional: SELECT, INSERT, UPDATE, DELETE
  "min_execution_time_ms": 100 // Optional: Filter slow queries
})
```

**Response:**

```json
{
  "logs": [
    {
      "timestamp": "2024-01-15T14:30:15Z",
      "query": "SELECT * FROM users WHERE role = $1",
      "execution_time_ms": 245,
      "rows_returned": 156,
      "query_type": "SELECT",
      "credits_used": 0.5
    }
  ]
}
```

### Monitoring Usage

```javascript
mcp__ainative-zerodb__zerodb_postgres_usage({
  "project_id": "your-project-id",
  "hours": 24 // Last 24 hours
})
```

**Response:**

```json
{
  "period_hours": 24,
  "total_queries": 15678,
  "total_execution_time_ms": 234567,
  "average_execution_time_ms": 14.9,
  "credits_used": 156.78,
  "breakdown": {
    "SELECT": 12340,
    "INSERT": 2134,
    "UPDATE": 987,
    "DELETE": 217
  }
}
```

## File Storage

Store and manage files with metadata and presigned URLs.

### Uploading Files

```javascript
mcp__ainative-zerodb__zerodb_upload_file({
  "file_name": "report.pdf",
  "content": "base64-encoded-content-here",
  "content_type": "application/pdf",
  "folder": "reports/2024",
  "metadata": {
    "author": "John Doe",
    "department": "Engineering",
    "generated_at": "2024-01-15"
  }
})
```

### Listing Files

```javascript
mcp__ainative-zerodb__zerodb_list_files({
  "folder": "reports/2024",
  "content_type": "application/pdf",
  "limit": 50,
  "offset": 0
})
```

### Downloading Files

```javascript
mcp__ainative-zerodb__zerodb_download_file({
  "file_id": "file-abc123"
})
```

**Response:**

```json
{
  "file_name": "report.pdf",
  "content": "base64-encoded-content",
  "content_type": "application/pdf",
  "size_bytes": 245678,
  "metadata": {...}
}
```

### Generating Presigned URLs

```javascript
mcp__ainative-zerodb__zerodb_generate_presigned_url({
  "file_id": "file-abc123",
  "operation": "download", // download or upload
  "expires_in_seconds": 3600 // 1 hour
})
```

**Response:**

```json
{
  "url": "https://storage.ainative.studio/files/abc123?signature=...",
  "expires_at": "2024-01-15T15:30:00Z"
}
```

## Agent Memory

Store and retrieve conversation context for long-term agent memory.

### Storing Memories

```javascript
mcp__ainative-zerodb__zerodb_store_memory({
  "session_id": "sess-abc123",
  "content": "User prefers TypeScript over JavaScript for backend development",
  "role": "system", // user, assistant, system
  "metadata": {
    "type": "preference",
    "category": "programming",
    "confidence": 0.95
  }
})
```

### Searching Memories

```javascript
mcp__ainative-zerodb__zerodb_search_memory({
  "session_id": "sess-abc123",
  "query": "programming language preferences",
  "top_k": 5,
  "similarity_threshold": 0.7
})
```

**Response:**

```json
{
  "results": [
    {
      "content": "User prefers TypeScript over JavaScript",
      "similarity_score": 0.92,
      "metadata": {
        "type": "preference",
        "timestamp": "2024-01-15T14:00:00Z"
      }
    }
  ]
}
```

### Getting Context Window

```javascript
mcp__ainative-zerodb__zerodb_get_context({
  "session_id": "sess-abc123",
  "max_tokens": 4000,
  "include_system": true
})
```

**Response:**

```json
{
  "context": [
    {
      "role": "user",
      "content": "How do I set up authentication?",
      "timestamp": "2024-01-15T14:15:00Z"
    },
    {
      "role": "assistant",
      "content": "To set up authentication...",
      "timestamp": "2024-01-15T14:15:10Z"
    }
  ],
  "total_tokens": 3456
}
```

## Quantum Search

Quantum-enhanced vector search for improved semantic matching.

### Hybrid Quantum-Classical Search

```javascript
mcp__ainative-zerodb__zerodb_quantum_hybrid_search({
  "query_vector": [0.1, 0.2, 0.3, ...],
  "top_k": 10,
  "quantum_weight": 0.5, // 0-1, balance between quantum and classical
  "classical_weight": 0.5,
  "namespace": "documentation"
})
```

**Benefits:**

- Better semantic understanding
- Improved recall for complex queries
- Enhanced context awareness
- Finds results traditional search might miss

## Event Streaming

Publish and subscribe to events for workflow automation.

### Creating Events

```javascript
mcp__ainative-zerodb__zerodb_create_event({
  "event_type": "user.signup",
  "source": "web-app",
  "data": {
    "user_id": "user-123",
    "email": "john@example.com",
    "timestamp": "2024-01-15T14:30:00Z"
  },
  "metadata": {
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0..."
  }
})
```

### Listing Events

```javascript
mcp__ainative-zerodb__zerodb_list_events({
  "event_type": "user.signup",
  "source": "web-app",
  "from_timestamp": "2024-01-15T00:00:00Z",
  "to_timestamp": "2024-01-15T23:59:59Z",
  "limit": 100
})
```

### Event Statistics

```javascript
mcp__ainative-zerodb__zerodb_event_stats({
  "event_type": "user.signup",
  "hours": 24
})
```

## Best Practices

### 1. Namespace Organization

```javascript
// Good: Organize vectors by purpose
namespaces = [
  "documentation",
  "code-snippets",
  "chat-history",
  "user-preferences"
]

// Bad: Generic namespace
namespace = "data"
```

### 2. Metadata Design

```javascript
// Good: Rich, queryable metadata
metadata = {
  "title": "Getting Started Guide",
  "type": "documentation",
  "category": "tutorials",
  "tags": ["beginner", "setup", "installation"],
  "version": "1.0",
  "created_at": "2024-01-15",
  "author": "docs-team"
}

// Bad: Minimal metadata
metadata = {
  "name": "doc1"
}
```

### 3. Batch Operations

```javascript
// Good: Batch upload
await batchUpsertVectors({
  vectors: arrayOf100Vectors
});

// Bad: Individual uploads
for (let vector of vectors) {
  await upsertVector(vector); // Too many API calls
}
```

### 4. Error Handling

```javascript
try {
  const result = await searchVectors(query);
  return result;
} catch (error) {
  if (error.code === 'RATE_LIMIT_EXCEEDED') {
    await delay(1000);
    return searchVectors(query); // Retry
  }
  throw error;
}
```

### 5. Connection Pooling (PostgreSQL)

```javascript
// Good: Use connection pool
const pool = new Pool({
  connectionString: dbUrl,
  max: 20,
  idleTimeoutMillis: 30000
});

// Bad: New connection per query
const client = new Client(dbUrl);
await client.connect();
// ... query ...
await client.end();
```

## Code Examples

### RAG (Retrieval-Augmented Generation) System

```javascript
async function ragQuery(userQuestion) {
  // 1. Generate query embedding
  const queryEmbedding = await generateEmbedding(userQuestion);

  // 2. Search for relevant documents
  const searchResults = await mcp__ainative_zerodb__zerodb_search_vectors({
    query_vector: queryEmbedding,
    top_k: 5,
    namespace: "documentation",
    similarity_threshold: 0.7
  });

  // 3. Build context from results
  const context = searchResults.results
    .map(r => r.metadata.content)
    .join('\n\n');

  // 4. Generate answer with LLM
  const answer = await llm.chat([
    {
      role: "system",
      content: "Answer based on this context:\n\n" + context
    },
    {
      role: "user",
      content: userQuestion
    }
  ]);

  return answer;
}
```

### Document Indexing Pipeline

```javascript
async function indexDocument(filePath) {
  // 1. Read and chunk document
  const content = await readFile(filePath);
  const chunks = splitIntoChunks(content, 500);

  // 2. Generate embeddings
  const vectors = await Promise.all(
    chunks.map(async (chunk, i) => ({
      vector_id: `${filePath}-chunk-${i}`,
      vector: await generateEmbedding(chunk.text),
      metadata: {
        source: filePath,
        chunk_index: i,
        content: chunk.text,
        created_at: new Date().toISOString()
      }
    }))
  );

  // 3. Batch upload to ZeroDB
  await mcp__ainative_zerodb__zerodb_batch_upsert_vectors({
    vectors: vectors,
    namespace: "documentation"
  });

  console.log(`Indexed ${vectors.length} chunks from ${filePath}`);
}
```

### Session Context Manager

```javascript
class SessionContextManager {
  constructor(sessionId) {
    this.sessionId = sessionId;
  }

  async addMessage(role, content) {
    await mcp__ainative_zerodb__zerodb_store_memory({
      session_id: this.sessionId,
      role: role,
      content: content,
      metadata: {
        timestamp: new Date().toISOString()
      }
    });
  }

  async getRelevantContext(query, maxTokens = 4000) {
    // Search for relevant past messages
    const memories = await mcp__ainative_zerodb__zerodb_search_memory({
      session_id: this.sessionId,
      query: query,
      top_k: 10
    });

    // Get recent context window
    const context = await mcp__ainative_zerodb__zerodb_get_context({
      session_id: this.sessionId,
      max_tokens: maxTokens
    });

    return {
      recent: context.context,
      relevant: memories.results
    };
  }
}
```

## Troubleshooting

### Connection Issues

**Problem:** Cannot connect to ZeroDB

**Solution:**

```bash
# Check authentication
ainative-code auth whoami

# Verify environment variables
echo $ZERODB_PROJECT_ID
echo $ZERODB_API_KEY

# Test connection
/zerodb-project-info
```

### Vector Dimension Mismatch

**Problem:** `Vector dimension mismatch` error

**Solution:**

All vectors must be exactly 1536 dimensions (OpenAI ada-002 compatible).

```javascript
// Verify vector length
if (vector.length !== 1536) {
  throw new Error(`Invalid vector dimension: ${vector.length}`);
}
```

### Query Performance

**Problem:** Slow queries

**Solutions:**

1. Add indexes to frequently queried fields
2. Use pagination (limit/offset)
3. Add metadata filters to narrow search
4. Use namespaces to partition data

### Storage Limits

**Problem:** Storage quota exceeded

**Solutions:**

```javascript
// Check usage
const stats = await mcp__ainative_zerodb__zerodb_project_stats();
console.log(stats.storage_usage);

// Delete old data
await mcp__ainative_zerodb__zerodb_delete_rows({
  table_name: "logs",
  filter: {
    created_at: { $lt: "2024-01-01T00:00:00Z" }
  }
});
```

## Next Steps

- [Design Token Integration](design-token-integration.md)
- [Strapi CMS Integration](strapi-integration.md)
- [RLHF Feedback System](rlhf-integration.md)
- [Authentication Setup](authentication-setup.md)
- [API Reference](/docs/api-reference/README.md)

## Support

- Documentation: https://docs.ainative.studio
- GitHub Issues: https://github.com/AINative-studio/ainative-code/issues
- Community: https://community.ainative.studio
