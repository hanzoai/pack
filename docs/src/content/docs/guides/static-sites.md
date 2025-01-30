---
title: Deploying Static Sites
description: Learn how to deploy static websites with Railpack
---

# Deploying Static Sites

This guide will show you how to deploy static websites using Railpack's staticfile provider.

## Basic Usage

The simplest way to deploy a static site is to have an `index.html` file in your project root:

```
my-static-site/
├── index.html
├── styles.css
└── script.js
```

Railpack will automatically detect this as a static site and configure it appropriately.

## Using a Custom Root Directory

### Option 1: Using Staticfile

Create a `Staticfile` in your project root:

```yaml
root: dist # Replace with your build output directory
```

### Option 2: Using Environment Variables

Set the `RAILPACK_STATIC_FILE_ROOT` environment variable:

```bash
export RAILPACK_STATIC_FILE_ROOT=dist
```

## Example Projects

### Basic Static Site

```
my-static-site/
├── index.html
├── styles.css
└── script.js
```

### Built Static Site

```
my-built-site/
├── Staticfile      # Contains: root: dist
├── src/
│   └── index.html
└── dist/           # Built files
    ├── index.html
    ├── styles.css
    └── script.js
```

### Public Directory Convention

```
my-site/
└── public/         # Automatically detected
    ├── index.html
    ├── styles.css
    └── script.js
```

## What's Next?

- Learn about [configuring Caddy](/docs/reference/config) for advanced use cases
- Explore other [providers](/docs/reference/providers) for different types of applications
