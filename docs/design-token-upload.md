# Design Token Upload Guide

This guide explains how to upload design tokens to the AINative Design system using the `ainative-code` CLI.

## Overview

Design tokens are design decisions represented as data. They include colors, typography, spacing, shadows, and other visual properties that ensure consistency across your application.

The `ainative-code design upload` command allows you to:
- Upload design tokens from JSON or YAML files
- Validate tokens before uploading
- Handle conflicts with existing tokens
- Track upload progress for large token sets

## Quick Start

### Basic Upload

```bash
ainative-code design upload \
  --tokens tokens.json \
  --project my-project
```

### Upload with Conflict Resolution

```bash
ainative-code design upload \
  --tokens tokens.yaml \
  --project my-project \
  --conflict merge
```

### Validate Only (No Upload)

```bash
ainative-code design upload \
  --tokens tokens.json \
  --validate-only
```

## Token File Format

Design tokens can be provided in JSON or YAML format.

### JSON Format

```json
{
  "tokens": [
    {
      "name": "primary-color",
      "value": "#007bff",
      "type": "color",
      "category": "colors",
      "description": "Primary brand color"
    },
    {
      "name": "spacing-base",
      "value": "16px",
      "type": "spacing",
      "category": "spacing",
      "description": "Base spacing unit"
    },
    {
      "name": "font-family-base",
      "value": "system-ui, -apple-system, sans-serif",
      "type": "font-family",
      "category": "typography"
    }
  ]
}
```

### YAML Format

```yaml
tokens:
  - name: primary-color
    value: "#007bff"
    type: color
    category: colors
    description: Primary brand color

  - name: spacing-base
    value: 16px
    type: spacing
    category: spacing
    description: Base spacing unit

  - name: font-family-base
    value: system-ui, -apple-system, sans-serif
    type: font-family
    category: typography
```

## Token Types

The following token types are supported:

### Color
- **Type**: `color`
- **Valid Formats**:
  - Hex: `#fff`, `#ffffff`, `#ffffffff`
  - RGB: `rgb(255, 255, 255)`
  - RGBA: `rgba(255, 255, 255, 0.5)`
  - HSL: `hsl(180, 50%, 50%)`
  - HSLA: `hsla(180, 50%, 50%, 0.5)`
  - Named: `white`, `black`, `red`, `transparent`

```json
{
  "name": "primary-color",
  "value": "#007bff",
  "type": "color",
  "category": "colors"
}
```

### Typography
- **Type**: `font-family`, `font-size`, `font-weight`, `line-height`, `letter-spacing`
- **Valid Formats**:
  - Font family: Any string (e.g., `"system-ui, sans-serif"`)
  - Font size: Size with unit (e.g., `16px`, `1.5rem`)
  - Font weight: `100-900`, `normal`, `bold`, `lighter`, `bolder`
  - Line height: Unitless number or size (e.g., `1.5`, `24px`)
  - Letter spacing: Size with unit (e.g., `0.5px`, `0.05em`)

```json
{
  "name": "font-size-base",
  "value": "16px",
  "type": "font-size",
  "category": "typography"
}
```

### Spacing
- **Type**: `spacing`
- **Valid Formats**: Number with unit (`px`, `rem`, `em`, `%`, `vh`, `vw`)

```json
{
  "name": "spacing-md",
  "value": "24px",
  "type": "spacing",
  "category": "spacing"
}
```

### Shadow
- **Type**: `shadow`
- **Valid Formats**: CSS box-shadow values

```json
{
  "name": "shadow-large",
  "value": "0 10px 30px rgba(0, 0, 0, 0.2)",
  "type": "shadow",
  "category": "effects"
}
```

### Border Radius
- **Type**: `border-radius`
- **Valid Formats**: Size with unit (e.g., `4px`, `0.5rem`)

```json
{
  "name": "border-radius-base",
  "value": "4px",
  "type": "border-radius",
  "category": "borders"
}
```

### Other Types
- `opacity`: `0` to `1`
- `z-index`: Integer or `auto`
- `duration`: Time with `ms` or `s` (e.g., `200ms`, `0.3s`)
- `easing`: Any string (e.g., `ease-in-out`, `cubic-bezier(0.4, 0, 0.2, 1)`)

## Conflict Resolution Modes

When uploading tokens that already exist in the Design system, you can specify how conflicts should be handled:

### Overwrite (Default)
Replaces existing tokens with new values.

```bash
ainative-code design upload --tokens tokens.json --conflict overwrite
```

### Merge
Merges new tokens with existing ones, preferring new values for conflicts.

```bash
ainative-code design upload --tokens tokens.json --conflict merge
```

### Skip
Skips conflicting tokens and keeps existing values.

```bash
ainative-code design upload --tokens tokens.json --conflict skip
```

## Token Validation

All tokens are validated before upload to ensure they conform to the design token specification.

### Validation Rules

1. **Required Fields**:
   - `name`: Token name (lowercase alphanumeric with dashes or dots)
   - `value`: Token value
   - `type`: Token type (must be a valid type)

2. **Name Format**:
   - Lowercase letters and numbers only
   - Separators: dash (`-`) or dot (`.`)
   - Examples: `primary-color`, `colors.primary.base`, `spacing-md`

3. **Value Format**:
   - Must match the expected format for the token type
   - Color values must be valid CSS colors
   - Sizing values must include units

4. **No Duplicates**:
   - Token names must be unique within a batch

### Validation Errors

If validation fails, you'll see detailed error messages:

```
❌ Token validation failed:
  - token 'primary-color': value - invalid color format: not-a-color
  - token 'spacing-base': value - invalid sizing format: 16 (expected number with unit)
  - token '': name - token name is required
```

## Progress Tracking

For large token sets (> 100 tokens), you can enable progress tracking:

```bash
ainative-code design upload \
  --tokens large-token-set.json \
  --project my-project \
  --progress
```

This will show real-time progress:

```
⬆️  Uploading: 250/500 tokens (50.0%)
```

## Upload Summary

After upload, you'll see a summary of the results:

```
📊 Upload Summary:
  ✅ Uploaded: 150 tokens
  🔄 Updated: 25 tokens
  ⏭️  Skipped: 10 tokens
```

## Complete Example

Here's a complete example workflow:

### 1. Create a token file (`tokens.json`)

```json
{
  "tokens": [
    {
      "name": "primary-color",
      "value": "#007bff",
      "type": "color",
      "category": "colors",
      "description": "Primary brand color"
    },
    {
      "name": "secondary-color",
      "value": "#6c757d",
      "type": "color",
      "category": "colors",
      "description": "Secondary brand color"
    },
    {
      "name": "spacing-xs",
      "value": "8px",
      "type": "spacing",
      "category": "spacing"
    },
    {
      "name": "spacing-sm",
      "value": "12px",
      "type": "spacing",
      "category": "spacing"
    },
    {
      "name": "spacing-md",
      "value": "16px",
      "type": "spacing",
      "category": "spacing"
    },
    {
      "name": "spacing-lg",
      "value": "24px",
      "type": "spacing",
      "category": "spacing"
    },
    {
      "name": "font-family-base",
      "value": "system-ui, -apple-system, sans-serif",
      "type": "font-family",
      "category": "typography"
    },
    {
      "name": "font-size-base",
      "value": "16px",
      "type": "font-size",
      "category": "typography"
    },
    {
      "name": "font-weight-normal",
      "value": "400",
      "type": "font-weight",
      "category": "typography"
    },
    {
      "name": "font-weight-bold",
      "value": "700",
      "type": "font-weight",
      "category": "typography"
    }
  ]
}
```

### 2. Validate the tokens

```bash
ainative-code design upload --tokens tokens.json --validate-only
```

Output:
```
📦 Loaded 10 tokens from tokens.json
✅ All tokens validated successfully
✨ Validation complete (upload skipped)
```

### 3. Upload to your project

```bash
ainative-code design upload \
  --tokens tokens.json \
  --project my-design-system \
  --conflict merge \
  --progress
```

Output:
```
📦 Loaded 10 tokens from tokens.json
✅ All tokens validated successfully

🚀 Uploading tokens to project 'my-design-system' (conflict mode: merge)...

📊 Upload Summary:
  ✅ Uploaded: 10 tokens

Successfully processed 10 tokens
```

## Best Practices

1. **Always validate first**: Use `--validate-only` to check your tokens before uploading
2. **Use version control**: Keep your token files in Git to track changes
3. **Organize by category**: Group related tokens using the `category` field
4. **Add descriptions**: Include meaningful descriptions for better documentation
5. **Use semantic names**: Choose clear, descriptive names (e.g., `primary-color` instead of `blue`)
6. **Consistent naming**: Stick to one naming convention (kebab-case or dot notation)
7. **Test locally**: Validate token files locally before uploading to production

## Troubleshooting

### Authentication Errors

If you see authentication errors, make sure you're logged in:

```bash
ainative-code login
```

### Project Not Found

Verify your project ID is correct:

```bash
ainative-code config get design.project_id
```

### Validation Errors

Check that your token values match the expected format for their type. See the "Token Types" section above for valid formats.

### Network Errors

If upload fails due to network issues, the command will automatically retry with exponential backoff (up to 3 attempts).

## API Integration

The design token upload functionality integrates with the AINative Design API:

- **Endpoint**: `POST /api/v1/design/tokens/upload`
- **Authentication**: JWT bearer token (obtained via `ainative-code login`)
- **Batch Size**: Tokens are uploaded in batches of 100 for optimal performance

## Related Commands

- `ainative-code design list` - List all design tokens
- `ainative-code design show <token-name>` - Show token details
- `ainative-code design export` - Export tokens to a file
- `ainative-code design sync` - Sync tokens with remote

## Further Reading

- [Design Token Extraction Guide](./design-token-extraction.md)
- [Design Token Best Practices](./design-token-best-practices.md)
- [AINative Design API Reference](./api-reference.md)
