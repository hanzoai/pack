---
title: Shell Scripts
description: Deploy applications using shell scripts with Railpack
---

Railpack can deploy applications that use shell scripts as their entry point.

## Detection

Your project will be automatically detected as a shell script application if any
of these conditions are met:

- A `start.sh` script exists in the root directory
- The `RAILPACK_SHELL_SCRIPT` environment variable is set to a valid script file

## Script File

Create a shell script in your project root (e.g., `start.sh`):

```bash
#!/bin/bash

echo "Hello world..."
```

## Config Variables

| Variable                | Description                              | Example     |
| ----------------------- | ---------------------------------------- | ----------- |
| `RAILPACK_SHELL_SCRIPT` | Specify a custom shell script to execute | `deploy.sh` |
