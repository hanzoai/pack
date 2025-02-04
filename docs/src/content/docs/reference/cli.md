---
title: CLI Reference
description: Complete reference for the Railpack CLI commands
---

Complete reference documentation for all Railpack CLI commands.

## Common Options

The following options are available across multiple commands:

| Flag                  | Description                                                                                                                |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `--env`               | Environment variables to set. Format: `KEY=VALUE`                                                                          |
| `--previous-versions` | Versions of packages used for previous builds. These versions will be used instead of the defaults. Format: `NAME@VERSION` |
| `--build-cmd`         | Build command to use                                                                                                       |
| `--start-cmd`         | Start command to use                                                                                                       |

## Commands

### build

Build a project using Railpack. This command takes a directory as input and
builds a container image using BuildKit.

**Usage:**

```bash
railpack build [options] DIRECTORY
```

**Options:**

| Flag          | Description                                      | Default |
| ------------- | ------------------------------------------------ | ------- |
| `--name`      | Name of the image to build                       |         |
| `--output`    | Output the final filesystem to a local directory |         |
| `--progress`  | BuildKit progress output mode (auto, plain, tty) | `auto`  |
| `--show-plan` | Show the build plan before building              | `false` |

### plan

Generate and view build plans for a project. This command analyzes a directory
and outputs the build plan that would be used.

**Usage:**

```bash
railpack plan [options] DIRECTORY
```

**Options:**

| Flag          | Description                   |
| ------------- | ----------------------------- |
| `--out`, `-o` | Output file name for the plan |

### info

View detailed information about a project. This command analyzes a directory and
provides information about the detected configuration, dependencies, and build
requirements.

**Usage:**

```bash
railpack info [options] DIRECTORY
```

**Options:**

| Flag       | Description                  | Default  |
| ---------- | ---------------------------- | -------- |
| `--format` | Output format (pretty, json) | `pretty` |
| `--out`    | Output file name             |          |

### schema

Output the JSON schema for the Railpack configuration file. This command outputs
the schema that defines the structure of valid Railpack configuration files. The
schema can be used by IDEs and other tools for providing autocompletion and
validation.

**Usage:**

```bash
railpack schema
```

### frontend

Start the BuildKit GRPC frontend server. This command is typically used
internally by the build system.

**Usage:**

```bash
railpack frontend
```

## Global Options

These options can be used with any command:

| Flag              | Description              |
| ----------------- | ------------------------ |
| `--help`, `-h`    | Show help information    |
| `--version`, `-v` | Show version information |
