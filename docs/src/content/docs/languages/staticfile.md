---
title: Static Sites
description: Deploy static websites with Railpack
---

Railpack can automatically build and setup a server for static sites that
require no build steps. The [Caddy](https://caddyserver.com/) server is used as
the underlying web server.

## Detection

Your project will be automatically detected as a static site if any of these conditions are met:

- A `Staticfile` configuration file exists in the root directory
- An `index.html` file exists in the root directory
- A `public` directory exists
- The `RAILPACK_STATIC_FILE_ROOT` environment variable is set

## Root Directory Resolution

The provider determines the root directory in this order:

1. `RAILPACK_STATIC_FILE_ROOT` environment variable if set
2. `root` directory specified in `Staticfile` if present
3. `public` directory if it exists
4. Current directory (`.`) if `index.html` exists in root

## Configuration

### Staticfile

You can create a `Staticfile` in your project root to configure the provider:

```yaml
root: dist # the directory containing your files to serve
```

### Environment Variables

| Variable                    | Description                 | Example     |
| --------------------------- | --------------------------- | ----------- |
| `RAILPACK_STATIC_FILE_ROOT` | Override the root directory | `/app/dist` |
