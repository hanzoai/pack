---
title: High Level Overview
description: Understanding Railpack's architecture and components
---

Railpack is split up into three main components:

- Core
  - The main logic that analyzes the app and generates the build plan
- Buildkit
  - Takes the build plan and generates [BuildKit
    LLB](https://github.com/moby/buildkit?tab=readme-ov-file#exploring-llb)
  - Starts a custom frontend or creates a BuildKit client to execute the build
    plan and generate an image
- CLI
  - The main entry point for Railpack

The core can be thought of as a _compiler_. The build plan that is generated is
independent from Docker, BuildKit, or any other tool that can be used to
generate an image. At the moment, BuildKit is the only _backend_, but more could
be added in the future.

## Build Plan

Thie build plan is a JSON object that contains all the information necessary to
generate an image. Things that it includes are:

- Base image
  - File system to start from
- Steps
  - Group of commands to run
- Start information
  - Command, variables, path, etc. Things needed when starting a container from
    the image
- Caches
  - Caches that are referenced by the commands in the steps
- Secrets
  - Build secrets that are referenced by the commands in the steps (just the
    names, not the values)

### Build Step

A step is a group of commands that is executed sequentially in the build. Steps
**can depend** on other steps, which means that they run with the filesystem of
all the outputs of the dependent steps. The build graph is constructed in such a
way that BuildKit will execute non-dependent steps in parallel.

Steps contain:

- Depends on
  - List of other steps that this step must run after
- Commands
  - List of commands to run in the build
    - Exec command: run a shell command
    - Copy command: copy files from the local context (user app) or another
      image into the current FS
    - Variable command: Set an environment variable
    - Path command: Prefix the global PATH with another directory
    - File command: Create a new file in the current FS referencing the step
      assets
- Assets
  - Mapping of name to file contents that is referenced in a file command
- Use secrets
  - Whether or not this step uses build secrets. [Docs](/architecture/secrets)
- Starting image
  - Instead of starting from the output of a previous step, it can start from a
    completely different image. This is typically used for root steps
- Outputs
  - List of file system paths that is the "result" of running this step. Parts
    of the FS that are not included will not appear in the final image. If not
    explicitly defined, the entire FS is assumed to be the output.

## Providers

Language suppport is provided through... providers. Providers are typically
associated a single language (e.g. node, python, php, etc.). A provider will

- Detect
  - Analyze the app and determine if it matches. (e.g. the node provider will
    check for the precense of a `package.json` file).
- Build
  - Modifies the build context with all the steps, commands, and everything that
    is needed to build for that language/framework.
