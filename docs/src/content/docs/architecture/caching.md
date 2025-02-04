---
title: Caching
description: Understanding Railpack's caching mechanisms
---

Railpack takes advantage of BuildKit layer and mount caches to speed up
successive builds.

## Layer Cache

Railpack takes advantage of BuildKit's layer cache and avoids busting the cache
when possible. Cache busting events are defined in a granular way as part of the
[steps commands list](/architecture/overview/#build-step). These include

- Copying files from the local context to the build context
- Changing environment variables
- Adding new generated files to the build context
- Executing shell commands in the build context

## Mount Cache

The [BuildKit mount
cache](https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/reference.md#run---mounttypecache)
is used to save the contents of a directory from the build context between
builds. This is useful for speeding up commands that download or compile assets
(e.g. npm install). The directory **does not** appear in the final image.

Caches are defined on the build plan and can be referenced via execution commands.

```json
{
  "caches": {
    "npm-install": {
      "directory": "/root/.npm",
      "type": "shared"
    }
  },

  "steps": {
    "install": {
      "commands": [
        {
          "cmd": "npm install",
          "caches": ["npm-install"]
        }
        // ...
      ]
      // ...
    }
  }
}
```

Caches are shared across all steps. This is useful for common caches such as the
apt-cache or apt-lists.
