# ZeroDB NoSQL CLI Usage Examples

This guide demonstrates common use cases for the ZeroDB NoSQL table operations CLI.

## Table of Contents
- [Configuration](#configuration)
- [Basic Operations](#basic-operations)
- [Advanced Queries](#advanced-queries)
- [Real-World Examples](#real-world-examples)

## Configuration

Before using ZeroDB commands, configure your project:

### Option 1: Configuration File
```bash
# Create or edit ~/.ainative-code.yaml
cat > ~/.ainative-code.yaml <<EOF
zerodb:
  base_url: "https://api.ainative.studio"
  project_id: "your-project-id"
EOF
```

### Option 2: Environment Variables
```bash
export AINATIVE_CODE_ZERODB_BASE_URL="https://api.ainative.studio"
export AINATIVE_CODE_ZERODB_PROJECT_ID="your-project-id"
```

## Basic Operations

### Create a Table

```bash
# Simple user table
ainative-code zerodb table create \
  --name users \
  --schema '{
    "type": "object",
    "properties": {
      "name": {"type": "string"},
      "email": {"type": "string"},
      "age": {"type": "number"}
    },
    "required": ["name", "email"]
  }'

# Product catalog table
ainative-code zerodb table create \
  --name products \
  --schema '{
    "type": "object",
    "properties": {
      "name": {"type": "string"},
      "price": {"type": "number"},
      "category": {"type": "string"},
      "inStock": {"type": "boolean"},
      "tags": {
        "type": "array",
        "items": {"type": "string"}
      }
    },
    "required": ["name", "price"]
  }'
```

### Insert Documents

```bash
# Insert a user
ainative-code zerodb table insert \
  --table users \
  --data '{"name":"Alice Johnson","email":"alice@example.com","age":28}'

# Insert a product
ainative-code zerodb table insert \
  --table products \
  --data '{
    "name": "Wireless Mouse",
    "price": 29.99,
    "category": "Electronics",
    "inStock": true,
    "tags": ["peripherals", "wireless", "office"]
  }'

# Insert with nested data
ainative-code zerodb table insert \
  --table users \
  --data '{
    "name": "Bob Smith",
    "email": "bob@example.com",
    "age": 35,
    "address": {
      "street": "123 Main St",
      "city": "San Francisco",
      "state": "CA",
      "zip": "94102"
    }
  }'
```

### Query Documents

```bash
# Get all documents
ainative-code zerodb table query --table users

# Get with JSON output
ainative-code zerodb table query --table users --json

# Paginated results
ainative-code zerodb table query \
  --table users \
  --limit 10 \
  --offset 0
```

### Update Documents

```bash
# Update user's age
ainative-code zerodb table update \
  --table users \
  --id user_abc123 \
  --data '{"age": 29}'

# Update multiple fields
ainative-code zerodb table update \
  --table users \
  --id user_abc123 \
  --data '{
    "age": 29,
    "email": "alice.new@example.com"
  }'
```

### Delete Documents

```bash
# Delete a user
ainative-code zerodb table delete \
  --table users \
  --id user_abc123
```

### List Tables

```bash
# Human-readable output
ainative-code zerodb table list

# JSON output for scripting
ainative-code zerodb table list --json
```

## Advanced Queries

### Equality Filters

```bash
# Find users by name
ainative-code zerodb table query \
  --table users \
  --filter '{"name": "Alice Johnson"}'

# Find products by category
ainative-code zerodb table query \
  --table products \
  --filter '{"category": "Electronics"}'
```

### Comparison Operators

```bash
# Users older than 18
ainative-code zerodb table query \
  --table users \
  --filter '{"age": {"$gt": 18}}'

# Users between 18 and 65
ainative-code zerodb table query \
  --table users \
  --filter '{"age": {"$gte": 18, "$lte": 65}}'

# Products cheaper than $50
ainative-code zerodb table query \
  --table products \
  --filter '{"price": {"$lt": 50}}'

# Products not priced at $29.99
ainative-code zerodb table query \
  --table products \
  --filter '{"price": {"$ne": 29.99}}'
```

### Logical Operators

```bash
# Users who are adults AND active
ainative-code zerodb table query \
  --table users \
  --filter '{
    "$and": [
      {"age": {"$gte": 18}},
      {"status": "active"}
    ]
  }'

# Products that are electronics OR office supplies
ainative-code zerodb table query \
  --table products \
  --filter '{
    "$or": [
      {"category": "Electronics"},
      {"category": "Office"}
    ]
  }'

# Users who are NOT inactive
ainative-code zerodb table query \
  --table users \
  --filter '{
    "$not": {
      "status": "inactive"
    }
  }'
```

### Array Operators

```bash
# Products with specific tags
ainative-code zerodb table query \
  --table products \
  --filter '{
    "tags": {
      "$in": ["wireless", "bluetooth", "portable"]
    }
  }'

# Products without specific tags
ainative-code zerodb table query \
  --table products \
  --filter '{
    "tags": {
      "$nin": ["discontinued", "refurbished"]
    }
  }'
```

### Existence Checks

```bash
# Users with email addresses
ainative-code zerodb table query \
  --table users \
  --filter '{"email": {"$exists": true}}'

# Users without phone numbers
ainative-code zerodb table query \
  --table users \
  --filter '{"phone": {"$exists": false}}'
```

### Sorting

```bash
# Sort users by age (descending)
ainative-code zerodb table query \
  --table users \
  --sort "age:desc"

# Sort by multiple fields
ainative-code zerodb table query \
  --table users \
  --sort "age:desc,name:asc"

# Sort products by price (ascending)
ainative-code zerodb table query \
  --table products \
  --sort "price:asc"
```

### Combined Filters and Sorting

```bash
# Active users, sorted by name
ainative-code zerodb table query \
  --table users \
  --filter '{"status": "active"}' \
  --sort "name:asc"

# Affordable electronics, sorted by price
ainative-code zerodb table query \
  --table products \
  --filter '{
    "$and": [
      {"category": "Electronics"},
      {"price": {"$lte": 100}}
    ]
  }' \
  --sort "price:asc" \
  --limit 20
```

## Real-World Examples

### E-Commerce Product Search

```bash
# Create products table
ainative-code zerodb table create \
  --name products \
  --schema '{
    "type": "object",
    "properties": {
      "sku": {"type": "string"},
      "name": {"type": "string"},
      "description": {"type": "string"},
      "price": {"type": "number"},
      "originalPrice": {"type": "number"},
      "category": {"type": "string"},
      "brand": {"type": "string"},
      "inStock": {"type": "boolean"},
      "rating": {"type": "number"},
      "reviewCount": {"type": "number"},
      "tags": {"type": "array", "items": {"type": "string"}}
    }
  }'

# Insert products
ainative-code zerodb table insert \
  --table products \
  --data '{
    "sku": "WM-001",
    "name": "Ergonomic Wireless Mouse",
    "description": "Comfortable wireless mouse with precision tracking",
    "price": 29.99,
    "originalPrice": 49.99,
    "category": "Electronics",
    "brand": "TechCo",
    "inStock": true,
    "rating": 4.5,
    "reviewCount": 128,
    "tags": ["wireless", "ergonomic", "office", "sale"]
  }'

# Find discounted electronics
ainative-code zerodb table query \
  --table products \
  --filter '{
    "$and": [
      {"category": "Electronics"},
      {"price": {"$lt": {"$field": "originalPrice"}}},
      {"inStock": true}
    ]
  }' \
  --sort "rating:desc,reviewCount:desc" \
  --limit 10

# Find highly-rated products
ainative-code zerodb table query \
  --table products \
  --filter '{
    "$and": [
      {"rating": {"$gte": 4.0}},
      {"reviewCount": {"$gte": 50}}
    ]
  }' \
  --sort "rating:desc"
```

### User Management System

```bash
# Create users table
ainative-code zerodb table create \
  --name users \
  --schema '{
    "type": "object",
    "properties": {
      "email": {"type": "string"},
      "name": {"type": "string"},
      "role": {"type": "string"},
      "status": {"type": "string"},
      "lastLogin": {"type": "string"},
      "createdAt": {"type": "string"},
      "permissions": {"type": "array", "items": {"type": "string"}},
      "metadata": {"type": "object"}
    }
  }'

# Insert admin user
ainative-code zerodb table insert \
  --table users \
  --data '{
    "email": "admin@company.com",
    "name": "Admin User",
    "role": "admin",
    "status": "active",
    "lastLogin": "2026-01-03T10:30:00Z",
    "createdAt": "2025-01-01T00:00:00Z",
    "permissions": ["read", "write", "delete", "admin"],
    "metadata": {
      "department": "IT",
      "location": "HQ"
    }
  }'

# Find active admins
ainative-code zerodb table query \
  --table users \
  --filter '{
    "$and": [
      {"role": "admin"},
      {"status": "active"}
    ]
  }'

# Find users with specific permissions
ainative-code zerodb table query \
  --table users \
  --filter '{
    "permissions": {
      "$in": ["admin", "delete"]
    }
  }'

# Update user status
ainative-code zerodb table update \
  --table users \
  --id user_xyz789 \
  --data '{"status": "inactive"}'
```

### Analytics Event Tracking

```bash
# Create events table
ainative-code zerodb table create \
  --name events \
  --schema '{
    "type": "object",
    "properties": {
      "eventType": {"type": "string"},
      "userId": {"type": "string"},
      "timestamp": {"type": "string"},
      "properties": {"type": "object"},
      "sessionId": {"type": "string"},
      "source": {"type": "string"}
    }
  }'

# Insert event
ainative-code zerodb table insert \
  --table events \
  --data '{
    "eventType": "page_view",
    "userId": "user_123",
    "timestamp": "2026-01-03T15:30:00Z",
    "properties": {
      "page": "/products",
      "referrer": "google.com",
      "duration": 45
    },
    "sessionId": "session_abc",
    "source": "web"
  }'

# Query recent events for a user
ainative-code zerodb table query \
  --table events \
  --filter '{
    "$and": [
      {"userId": "user_123"},
      {"timestamp": {"$gte": "2026-01-01T00:00:00Z"}}
    ]
  }' \
  --sort "timestamp:desc" \
  --limit 100

# Query conversion events
ainative-code zerodb table query \
  --table events \
  --filter '{
    "eventType": {
      "$in": ["purchase", "signup", "subscription"]
    }
  }' \
  --sort "timestamp:desc"
```

## Tips and Best Practices

### 1. Use JSON Output for Scripting

```bash
# Process query results with jq
ainative-code zerodb table query \
  --table users \
  --filter '{"status": "active"}' \
  --json | jq '.[] | .data.email'

# Count results
ainative-code zerodb table list --json | jq 'length'
```

### 2. Pagination for Large Datasets

```bash
# Fetch first page
ainative-code zerodb table query \
  --table users \
  --limit 100 \
  --offset 0

# Fetch second page
ainative-code zerodb table query \
  --table users \
  --limit 100 \
  --offset 100
```

### 3. Complex Filters in Scripts

```bash
# Save filter to file
cat > filter.json <<EOF
{
  "$and": [
    {"age": {"$gte": 18, "$lte": 65}},
    {"status": "active"},
    {"$or": [
      {"role": "admin"},
      {"permissions": {"$in": ["write", "admin"]}}
    ]}
  ]
}
EOF

# Use filter from file
ainative-code zerodb table query \
  --table users \
  --filter "$(cat filter.json)"
```

### 4. Backup and Restore

```bash
# Export all documents from a table
ainative-code zerodb table query \
  --table users \
  --json > users_backup.json

# Re-insert from backup (requires scripting)
jq -c '.[]' users_backup.json | while read doc; do
  ainative-code zerodb table insert \
    --table users_restored \
    --data "$doc"
done
```

## Troubleshooting

### Common Errors

**Error: "zerodb.project_id not configured"**
```bash
# Solution: Set project ID in config or environment
export AINATIVE_CODE_ZERODB_PROJECT_ID="your-project-id"
```

**Error: "invalid filter JSON"**
```bash
# Solution: Validate JSON with jq first
echo '{"age": {"$gte": 18}}' | jq .
```

**Error: "authentication failed"**
```bash
# Solution: Login first
ainative-code login
```

### Debug Mode

```bash
# Enable verbose logging
ainative-code zerodb table query \
  --table users \
  --verbose
```

---

For more information, see:
- [ZeroDB NoSQL API Documentation](../api/zerodb.md)
- [Configuration Guide](../user-guide/configuration.md)
- [Authentication Guide](../user-guide/authentication.md)
