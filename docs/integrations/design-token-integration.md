# Design Token Integration Guide

## Overview

Design tokens are design decisions represented as data (colors, typography, spacing, shadows, etc.) that ensure consistency across your application. AINative Code provides comprehensive design token extraction, generation, and synchronization capabilities.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Token Extraction](#token-extraction)
3. [Token Generation](#token-generation)
4. [Token Upload and Sync](#token-upload-and-sync)
5. [Workflow Automation](#workflow-automation)
6. [Integration Examples](#integration-examples)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites

```bash
# Login to AINative platform
ainative-code auth login

# Configure design service (if using Figma)
export FIGMA_TOKEN="your-figma-token"
```

### Basic Workflow

```bash
# 1. Extract tokens from Figma
ainative-code design extract \
  --source figma \
  --file-id "ABC123" \
  --output tokens.json

# 2. Generate CSS
ainative-code design generate \
  --input tokens.json \
  --format css \
  --output styles/tokens.css

# 3. Upload to AINative Design System
ainative-code design upload \
  --tokens tokens.json \
  --project my-app
```

## Token Extraction

### From Figma

**Extract All Design Tokens:**

```bash
ainative-code design extract \
  --source figma \
  --file-id "ABC123XYZ" \
  --token "${FIGMA_TOKEN}" \
  --output tokens.json
```

**Extract Specific Components:**

```bash
ainative-code design extract \
  --source figma \
  --file-id "ABC123XYZ" \
  --token "${FIGMA_TOKEN}" \
  --filter "colors,spacing" \
  --output tokens.json
```

**Extract from Specific Page:**

```bash
ainative-code design extract \
  --source figma \
  --file-id "ABC123XYZ" \
  --token "${FIGMA_TOKEN}" \
  --page "Design System" \
  --output tokens.json
```

### From Sketch

```bash
ainative-code design extract \
  --source sketch \
  --file design-system.sketch \
  --output tokens.json
```

### From Adobe XD

```bash
ainative-code design extract \
  --source xd \
  --file design-system.xd \
  --output tokens.json
```

### Token File Format

```json
{
  "tokens": [
    {
      "name": "primary-color",
      "value": "#6366F1",
      "type": "color",
      "category": "colors",
      "description": "Primary brand color used for CTAs and important UI elements"
    },
    {
      "name": "spacing-base",
      "value": "16px",
      "type": "spacing",
      "category": "spacing",
      "description": "Base spacing unit, used for consistent padding and margins"
    },
    {
      "name": "font-family-heading",
      "value": "Inter, system-ui, -apple-system, sans-serif",
      "type": "font-family",
      "category": "typography",
      "description": "Font stack for headings"
    },
    {
      "name": "shadow-elevated",
      "value": "0 10px 30px rgba(0, 0, 0, 0.1)",
      "type": "shadow",
      "category": "effects",
      "description": "Shadow for elevated UI components"
    },
    {
      "name": "border-radius-medium",
      "value": "8px",
      "type": "border-radius",
      "category": "borders",
      "description": "Medium border radius for cards and containers"
    }
  ]
}
```

## Token Generation

### CSS Variables

**Generate CSS Custom Properties:**

```bash
ainative-code design generate \
  --input tokens.json \
  --format css \
  --output styles/tokens.css
```

**Generated Output:**

```css
:root {
  /* Colors */
  --primary-color: #6366F1;
  --secondary-color: #8B5CF6;
  --success-color: #10B981;
  --error-color: #EF4444;
  --warning-color: #F59E0B;

  /* Spacing */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;

  /* Typography */
  --font-family-base: system-ui, -apple-system, sans-serif;
  --font-family-heading: Inter, system-ui, sans-serif;
  --font-size-xs: 12px;
  --font-size-sm: 14px;
  --font-size-md: 16px;
  --font-size-lg: 18px;
  --font-size-xl: 24px;
  --font-weight-normal: 400;
  --font-weight-medium: 500;
  --font-weight-bold: 700;

  /* Effects */
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 30px rgba(0, 0, 0, 0.15);

  /* Borders */
  --border-radius-sm: 4px;
  --border-radius-md: 8px;
  --border-radius-lg: 16px;
  --border-radius-full: 9999px;
}
```

### SCSS/Sass Variables

```bash
ainative-code design generate \
  --input tokens.json \
  --format scss \
  --output styles/_tokens.scss
```

**Generated Output:**

```scss
// Colors
$primary-color: #6366F1;
$secondary-color: #8B5CF6;
$success-color: #10B981;

// Spacing
$spacing-xs: 4px;
$spacing-sm: 8px;
$spacing-md: 16px;

// Typography
$font-family-base: system-ui, -apple-system, sans-serif;
$font-size-md: 16px;
$font-weight-normal: 400;

// Create spacing map
$spacing: (
  'xs': $spacing-xs,
  'sm': $spacing-sm,
  'md': $spacing-md,
  'lg': $spacing-lg,
  'xl': $spacing-xl
);
```

### JavaScript/TypeScript

```bash
ainative-code design generate \
  --input tokens.json \
  --format typescript \
  --output src/tokens.ts
```

**Generated Output:**

```typescript
export const tokens = {
  colors: {
    primary: '#6366F1',
    secondary: '#8B5CF6',
    success: '#10B981',
    error: '#EF4444',
    warning: '#F59E0B',
  },
  spacing: {
    xs: '4px',
    sm: '8px',
    md: '16px',
    lg: '24px',
    xl: '32px',
  },
  typography: {
    fontFamily: {
      base: 'system-ui, -apple-system, sans-serif',
      heading: 'Inter, system-ui, sans-serif',
    },
    fontSize: {
      xs: '12px',
      sm: '14px',
      md: '16px',
      lg: '18px',
      xl: '24px',
    },
    fontWeight: {
      normal: 400,
      medium: 500,
      bold: 700,
    },
  },
  effects: {
    shadow: {
      sm: '0 1px 2px rgba(0, 0, 0, 0.05)',
      md: '0 4px 6px rgba(0, 0, 0, 0.1)',
      lg: '0 10px 30px rgba(0, 0, 0, 0.15)',
    },
  },
  borders: {
    radius: {
      sm: '4px',
      md: '8px',
      lg: '16px',
      full: '9999px',
    },
  },
} as const;

export type Tokens = typeof tokens;
```

### Tailwind Configuration

```bash
ainative-code design generate \
  --input tokens.json \
  --format tailwind \
  --output tailwind.config.js
```

**Generated Output:**

```javascript
module.exports = {
  theme: {
    extend: {
      colors: {
        primary: '#6366F1',
        secondary: '#8B5CF6',
        success: '#10B981',
        error: '#EF4444',
        warning: '#F59E0B',
      },
      spacing: {
        'xs': '4px',
        'sm': '8px',
        'md': '16px',
        'lg': '24px',
        'xl': '32px',
      },
      fontFamily: {
        sans: ['system-ui', '-apple-system', 'sans-serif'],
        heading: ['Inter', 'system-ui', 'sans-serif'],
      },
      fontSize: {
        'xs': '12px',
        'sm': '14px',
        'md': '16px',
        'lg': '18px',
        'xl': '24px',
      },
      boxShadow: {
        'sm': '0 1px 2px rgba(0, 0, 0, 0.05)',
        'md': '0 4px 6px rgba(0, 0, 0, 0.1)',
        'lg': '0 10px 30px rgba(0, 0, 0, 0.15)',
      },
      borderRadius: {
        'sm': '4px',
        'md': '8px',
        'lg': '16px',
        'full': '9999px',
      },
    },
  },
};
```

### JSON Output

```bash
ainative-code design generate \
  --input tokens.json \
  --format json \
  --output dist/tokens.json
```

## Token Upload and Sync

### Upload to AINative Design System

```bash
ainative-code design upload \
  --tokens tokens.json \
  --project my-design-system \
  --conflict merge
```

**Conflict Resolution Modes:**

- `overwrite`: Replace existing tokens
- `merge`: Merge new with existing (prefer new)
- `skip`: Skip conflicting tokens

### Validate Before Upload

```bash
ainative-code design upload \
  --tokens tokens.json \
  --validate-only
```

**Output:**

```
📦 Loaded 45 tokens from tokens.json
✅ All tokens validated successfully

Token Summary:
  Colors: 12
  Spacing: 8
  Typography: 15
  Effects: 6
  Borders: 4

✨ Validation complete (upload skipped)
```

### Bidirectional Sync

```bash
# Download from AINative Design System
ainative-code design download \
  --project my-design-system \
  --output tokens-remote.json

# Merge local and remote
ainative-code design merge \
  --local tokens.json \
  --remote tokens-remote.json \
  --output tokens-merged.json \
  --strategy prefer-local

# Upload merged tokens
ainative-code design upload \
  --tokens tokens-merged.json \
  --project my-design-system
```

## Workflow Automation

### Watch Mode

**Auto-sync on Design File Changes:**

```bash
ainative-code design watch \
  --source figma \
  --file-id "ABC123" \
  --output tokens.json \
  --auto-generate css \
  --interval 300  # Check every 5 minutes
```

**What it does:**

1. Monitors Figma file for changes
2. Automatically extracts updated tokens
3. Regenerates CSS files
4. Optionally uploads to AINative Design System

### CI/CD Integration

**GitHub Actions Example:**

```yaml
name: Design Token Sync

on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours
  workflow_dispatch:

jobs:
  sync-tokens:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install AINative Code
        run: |
          curl -fsSL https://install.ainative.studio | sh

      - name: Extract Design Tokens
        env:
          FIGMA_TOKEN: ${{ secrets.FIGMA_TOKEN }}
        run: |
          ainative-code design extract \
            --source figma \
            --file-id "${{ secrets.FIGMA_FILE_ID }}" \
            --output tokens.json

      - name: Generate CSS
        run: |
          ainative-code design generate \
            --input tokens.json \
            --format css \
            --output src/styles/tokens.css

      - name: Generate TypeScript
        run: |
          ainative-code design generate \
            --input tokens.json \
            --format typescript \
            --output src/tokens.ts

      - name: Commit Changes
        run: |
          git config user.name "Design Token Bot"
          git config user.email "bot@ainative.studio"
          git add tokens.json src/styles/tokens.css src/tokens.ts
          git diff --quiet && git diff --staged --quiet || \
            git commit -m "Update design tokens from Figma"
          git push
```

### NPM Scripts

**package.json:**

```json
{
  "scripts": {
    "tokens:extract": "ainative-code design extract --source figma --file-id $FIGMA_FILE_ID --output tokens.json",
    "tokens:generate": "npm run tokens:generate:css && npm run tokens:generate:ts",
    "tokens:generate:css": "ainative-code design generate --input tokens.json --format css --output src/styles/tokens.css",
    "tokens:generate:ts": "ainative-code design generate --input tokens.json --format typescript --output src/tokens.ts",
    "tokens:upload": "ainative-code design upload --tokens tokens.json --project my-app",
    "tokens:sync": "npm run tokens:extract && npm run tokens:generate && npm run tokens:upload"
  }
}
```

**Usage:**

```bash
npm run tokens:sync
```

## Integration Examples

### React Component Usage

```typescript
import { tokens } from './tokens';

function Button({ variant = 'primary', children }) {
  return (
    <button
      style={{
        backgroundColor: tokens.colors[variant],
        padding: `${tokens.spacing.sm} ${tokens.spacing.md}`,
        fontSize: tokens.typography.fontSize.md,
        fontWeight: tokens.typography.fontWeight.medium,
        borderRadius: tokens.borders.radius.md,
        boxShadow: tokens.effects.shadow.sm,
        fontFamily: tokens.typography.fontFamily.base,
      }}
    >
      {children}
    </button>
  );
}
```

### CSS-in-JS (Styled Components)

```typescript
import styled from 'styled-components';
import { tokens } from './tokens';

const Button = styled.button`
  background-color: ${tokens.colors.primary};
  padding: ${tokens.spacing.sm} ${tokens.spacing.md};
  font-size: ${tokens.typography.fontSize.md};
  font-weight: ${tokens.typography.fontWeight.medium};
  border-radius: ${tokens.borders.radius.md};
  box-shadow: ${tokens.effects.shadow.sm};
  font-family: ${tokens.typography.fontFamily.base};

  &:hover {
    background-color: ${tokens.colors.primaryDark};
    box-shadow: ${tokens.effects.shadow.md};
  }
`;
```

### Tailwind with Design Tokens

```html
<button class="
  bg-primary
  px-md py-sm
  text-md font-medium
  rounded-md
  shadow-sm
  hover:shadow-md
  font-sans
">
  Click Me
</button>
```

### Vue Component

```vue
<template>
  <button :style="buttonStyles">
    <slot></slot>
  </button>
</template>

<script setup>
import { tokens } from './tokens';

const buttonStyles = {
  backgroundColor: tokens.colors.primary,
  padding: `${tokens.spacing.sm} ${tokens.spacing.md}`,
  fontSize: tokens.typography.fontSize.md,
  borderRadius: tokens.borders.radius.md,
};
</script>
```

## Best Practices

### 1. Semantic Naming

```javascript
// Good: Semantic names
{
  "primary-color": "#6366F1",
  "success-color": "#10B981",
  "spacing-base": "16px"
}

// Bad: Generic names
{
  "blue": "#6366F1",
  "green": "#10B981",
  "size-1": "16px"
}
```

### 2. Consistent Organization

```javascript
// Organize by category
{
  "tokens": [
    // Colors
    {"name": "primary-color", "category": "colors"},
    {"name": "secondary-color", "category": "colors"},

    // Spacing
    {"name": "spacing-xs", "category": "spacing"},
    {"name": "spacing-sm", "category": "spacing"},

    // Typography
    {"name": "font-size-base", "category": "typography"}
  ]
}
```

### 3. Version Control

```bash
# Track token changes
git add tokens.json
git commit -m "Update design tokens: Add new color palette"

# Tag releases
git tag -a v1.2.0 -m "Design tokens v1.2.0"
git push origin v1.2.0
```

### 4. Documentation

```json
{
  "name": "primary-color",
  "value": "#6366F1",
  "type": "color",
  "category": "colors",
  "description": "Primary brand color. Used for CTAs, links, and important UI elements. Ensure WCAG AA contrast ratio (4.5:1) when used with white text."
}
```

### 5. Token Hierarchy

```javascript
// Base tokens
const baseTokens = {
  "color-indigo-500": "#6366F1"
};

// Semantic tokens (reference base)
const semanticTokens = {
  "primary-color": "{color-indigo-500}"
};

// Component tokens (reference semantic)
const componentTokens = {
  "button-background": "{primary-color}"
};
```

### 6. Automated Testing

```javascript
// Test token values
describe('Design Tokens', () => {
  it('should have valid color formats', () => {
    const colorTokens = tokens.tokens.filter(t => t.type === 'color');
    colorTokens.forEach(token => {
      expect(token.value).toMatch(/^#[0-9A-F]{6}$/i);
    });
  });

  it('should have valid spacing units', () => {
    const spacingTokens = tokens.tokens.filter(t => t.type === 'spacing');
    spacingTokens.forEach(token => {
      expect(token.value).toMatch(/^\d+(px|rem|em)$/);
    });
  });
});
```

## Troubleshooting

### Figma Token Extraction Fails

**Problem:** Cannot extract tokens from Figma

**Solutions:**

```bash
# Verify Figma token
echo $FIGMA_TOKEN

# Get token from Figma settings:
# Figma > Settings > Account > Personal Access Tokens

# Test token
curl -H "X-Figma-Token: $FIGMA_TOKEN" \
  https://api.figma.com/v1/me

# Use correct file ID
# URL: https://www.figma.com/file/ABC123XYZ/Design-System
# File ID: ABC123XYZ
```

### Token Validation Errors

**Problem:** Tokens fail validation

**Solutions:**

```bash
# Check token format
ainative-code design upload --tokens tokens.json --validate-only

# Common issues:
# 1. Invalid color format
"value": "blue"  # Bad
"value": "#0000FF"  # Good

# 2. Missing units
"value": "16"  # Bad
"value": "16px"  # Good

# 3. Invalid type
"type": "colour"  # Bad
"type": "color"  # Good
```

### Generated Code Not Working

**Problem:** Generated CSS/JS not working as expected

**Solutions:**

```bash
# Regenerate with specific format
ainative-code design generate \
  --input tokens.json \
  --format css \
  --output styles/tokens.css \
  --prefix custom  # Add prefix to avoid conflicts

# Verify output
cat styles/tokens.css

# Check for syntax errors
npx prettier --check styles/tokens.css
```

### Sync Conflicts

**Problem:** Token conflicts when syncing

**Solutions:**

```bash
# Use merge strategy
ainative-code design upload \
  --tokens tokens.json \
  --conflict merge

# Or download and manually merge
ainative-code design download \
  --project my-app \
  --output tokens-remote.json

# Compare files
diff tokens.json tokens-remote.json
```

## Next Steps

- [Strapi CMS Integration](strapi-integration.md)
- [ZeroDB Integration](zerodb-integration.md)
- [RLHF Feedback System](rlhf-integration.md)
- [Authentication Setup](authentication-setup.md)

## Resources

- [Figma API Documentation](https://www.figma.com/developers/api)
- [Design Tokens W3C Specification](https://design-tokens.github.io/community-group/)
- [Tailwind Configuration](https://tailwindcss.com/docs/configuration)
- [CSS Custom Properties](https://developer.mozilla.org/en-US/docs/Web/CSS/--*)
