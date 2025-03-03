---
title: Python
description: Building Python applications with Railpack
---

Railpack builds and deploys Python applications with support for various package managers and dependency management tools.

## Detection

Your project will be detected as a Python application if any of these conditions are met:

- A `main.py` file exists in the root directory
- A `requirements.txt` file exists
- A `pyproject.toml` file exists
- A `Pipfile` exists

## Versions

The Python version is determined in the following order:

- Set via the `RAILPACK_PYTHON_VERSION` environment variable
- Read from the `.python-version` file
- Read from the `runtime.txt` file
- Read from the `Pipfile` if present
- Defaults to `3.13.2`

## Runtime Variables

These variables are available at runtime:

```
PYTHONFAULTHANDLER=1
PYTHONUNBUFFERED=1
PYTHONHASHSEED=random
PYTHONDONTWRITEBYTECODE=1
PIP_DISABLE_PIP_VERSION_CHECK=1
PIP_DEFAULT_TIMEOUT=100
```

## Configuration

Railpack builds your Python application based on your project structure. The build process:

- Installs Python and required system dependencies
- Installs project dependencies using your preferred package manager
- Configures the Python environment for production

The start command is the `main.py` file in the root directory.

### Package Managers

Railpack supports multiple Python package managers:

- **pip** - Uses `requirements.txt` for dependencies
- **poetry** - Uses `pyproject.toml` and `poetry.lock`
- **pdm** - Uses `pyproject.toml` and `pdm.lock`
- **uv** - Uses `pyproject.toml` and `uv.lock`
- **pipenv** - Uses `Pipfile`

### Config Variables

| Variable                  | Description                 | Example |
| ------------------------- | --------------------------- | ------- |
| `RAILPACK_PYTHON_VERSION` | Override the Python version | `3.11`  |

### System Dependencies

Railpack installs system dependencies for common Python packages:

- **pdf2image**: Installs `poppler-utils`
- **pydub**: Installs `ffmpeg`
- **pymovie**: Installs `ffmpeg`, `qt5-qmake`, and related Qt packages
