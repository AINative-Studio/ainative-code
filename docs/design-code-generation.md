# Design Code Generation

## Overview

The `ainative-code design generate` command generates code from design tokens in various output formats. This enables teams to maintain a single source of truth for design tokens and automatically generate platform-specific code.

## Supported Output Formats

### 1. Tailwind CSS Configuration

Generate a Tailwind CSS configuration file with your design tokens:

```bash
ainative-code design generate \
  --tokens design-tokens.json \
  --format tailwind \
  --output tailwind.config.js
```

**Output Example:**
```javascript
/** @type {import('tailwindcss').Config} */
module.exports = {
  theme: {
    extend: {
      colors: {
        "primary-color": "#007bff",
        "secondary-color": "#6c757d",
      },
      spacing: {
        "spacing-base": "16px",
        "spacing-sm": "8px",
      },
      // ...
    },
  },
  plugins: [],
}
```

### 2. CSS Custom Properties

Generate CSS variables (custom properties):

```bash
ainative-code design generate \
  --tokens design-tokens.json \
  --format css \
  --output design-tokens.css
```

**Output Example:**
```css
/**
 * Design Tokens - CSS Custom Properties
 * Generated from design system
 */

:root {
  --primary-color: #007bff;
  --secondary-color: #6c757d;
  --spacing-base: 16px;
  --font-family-base: Helvetica, Arial, sans-serif;
}
```

### 3. SCSS Variables

Generate SCSS variables:

```bash
ainative-code design generate \
  --tokens design-tokens.json \
  --format scss \
  --output _tokens.scss
```

**Output Example:**
```scss
/**
 * Design Tokens - SCSS Variables
 * Generated from design system
 */

$primary-color: #007bff;
$secondary-color: #6c757d;
$spacing-base: 16px;
$font-family-base: Helvetica, Arial, sans-serif;
```

### 4. TypeScript Constants

Generate TypeScript constants with type safety:

```bash
ainative-code design generate \
  --tokens design-tokens.json \
  --format typescript \
  --output tokens.ts
```

**Output Example:**
```typescript
/**
 * Design Tokens - TypeScript Constants
 * Generated from design system
 */

export const DesignTokens = {
  primaryColor: "#007bff",
  secondaryColor: "#6c757d",
  spacingBase: "16px",
  fontFamilyBase: "Helvetica, Arial, sans-serif",
} as const;

export type DesignToken = typeof DesignTokens;
```

### 5. JavaScript Constants

Generate JavaScript constants (ES6):

```bash
ainative-code design generate \
  --tokens design-tokens.json \
  --format javascript \
  --output tokens.js
```

**Output Example:**
```javascript
/**
 * Design Tokens - JavaScript Constants
 * Generated from design system
 */

export const DesignTokens = {
  primaryColor: "#007bff",
  secondaryColor: "#6c757d",
  spacingBase: "16px",
  fontFamilyBase: "Helvetica, Arial, sans-serif",
};
```

### 6. JSON Output

Generate formatted JSON output:

```bash
ainative-code design generate \
  --tokens design-tokens.json \
  --format json \
  --output tokens.json \
  --pretty
```

## Command Options

### Required Flags

- `--tokens, -t`: Path to input tokens file (JSON format) - **required**

### Optional Flags

- `--format, -f`: Output format (default: "json")
  - `tailwind`, `tw`: Tailwind CSS configuration
  - `css`: CSS custom properties
  - `scss`, `sass`: SCSS variables
  - `typescript`, `ts`: TypeScript constants
  - `javascript`, `js`: JavaScript constants
  - `json`: JSON format

- `--output, -o`: Output file path (prints to stdout if not specified)

- `--pretty, -p`: Pretty-print output for JSON format (default: true)

- `--template`: Path to custom template file (advanced usage)

## Input Format

Design tokens should be provided as a JSON file with the following structure:

```json
{
  "tokens": [
    {
      "name": "primary-color",
      "type": "color",
      "value": "#007bff",
      "category": "colors",
      "description": "Primary brand color"
    },
    {
      "name": "spacing-base",
      "type": "spacing",
      "value": "16px",
      "category": "spacing"
    }
  ]
}
```

### Token Structure

Each token must have:
- `name`: Token identifier (kebab-case recommended)
- `type`: Token type (color, spacing, typography, etc.)
- `value`: Token value

Optional fields:
- `category`: Category for grouping tokens
- `description`: Human-readable description
- `metadata`: Additional metadata as key-value pairs

### Supported Token Types

- `color`: Color values (hex, rgb, hsl)
- `spacing`: Spacing values (px, rem, em)
- `typography`: Typography-related values
- `font-family`: Font family stacks
- `font-size`: Font sizes
- `shadow`: Box shadow values
- `border-radius`: Border radius values

## Custom Templates

For advanced use cases, you can provide custom templates:

```bash
ainative-code design generate \
  --tokens design-tokens.json \
  --template my-template.tmpl \
  --output custom-output.txt
```

### Template Syntax

Templates use Go's `text/template` syntax with helpful functions:

**Available Functions:**
- `kebabCase`: Convert to kebab-case
- `camelCase`: Convert to camelCase
- `pascalCase`: Convert to PascalCase
- `snakeCase`: Convert to snake_case
- `upper`: Convert to uppercase
- `lower`: Convert to lowercase
- `quote`: Wrap in double quotes
- `indent N`: Indent by N spaces

**Example Template:**
```
Design Tokens
=============

{{- range .Tokens }}
{{ .Name | kebabCase }}: {{ .Value }}
{{- end }}
```

## Integration with Design Systems

### Workflow Example

1. **Extract tokens from design files:**
   ```bash
   ainative-code design extract \
     --source styles.css \
     --output tokens.json
   ```

2. **Generate platform-specific code:**
   ```bash
   # Generate Tailwind config
   ainative-code design generate \
     --tokens tokens.json \
     --format tailwind \
     --output tailwind.config.js

   # Generate TypeScript constants
   ainative-code design generate \
     --tokens tokens.json \
     --format typescript \
     --output src/tokens.ts
   ```

3. **Use in your application:**
   ```typescript
   import { DesignTokens } from './tokens';

   const buttonStyle = {
     backgroundColor: DesignTokens.primaryColor,
     padding: DesignTokens.spacingBase,
   };
   ```

## CI/CD Integration

Add code generation to your build pipeline:

```yaml
# .github/workflows/design-tokens.yml
name: Generate Design Tokens

on:
  push:
    paths:
      - 'design-tokens.json'

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Generate Tailwind Config
        run: |
          ainative-code design generate \
            --tokens design-tokens.json \
            --format tailwind \
            --output tailwind.config.js

      - name: Commit generated files
        run: |
          git config user.name "Design Token Bot"
          git commit -am "Update generated design token files"
          git push
```

## Best Practices

1. **Single Source of Truth**: Maintain design tokens in a single JSON file
2. **Version Control**: Commit both token definitions and generated code
3. **Automation**: Automate code generation in CI/CD pipelines
4. **Naming Conventions**: Use consistent, descriptive token names
5. **Documentation**: Include descriptions for all tokens
6. **Categories**: Group related tokens using categories

## Troubleshooting

### Common Issues

**Invalid JSON format:**
```
Error: failed to parse tokens JSON
```
Solution: Validate your JSON file structure matches the expected format.

**Missing required fields:**
```
Error: token name is required
```
Solution: Ensure all tokens have name, type, and value fields.

**Unsupported format:**
```
Error: invalid format 'xml'
```
Solution: Use one of the supported formats: tailwind, css, scss, typescript, javascript, json.

## Examples

See the `/examples` directory for complete examples:
- `examples/design-tokens.json`: Sample token definitions
- `examples/generated/`: Generated output in all formats
