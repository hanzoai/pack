---
title: Configuration File
description: Learn about the railpack.json configuration file format and options
---

Railpack will look for a `railpack.json` file in the root of the directory being
built. If found, that configuration will be used to change how the plan is
built.

A config file looks something like this:

```json
{
  "steps": {
    "install": {
      "commands": ["npm install"]
    },
    "build": {
      "dependsOn": ["install"],
      "commands": ["npm run build"],
      "usesSecrets": true,
      "outputs": ["/app/dist"]
    }
  },

  "start": {
    "cmd": "node dist/index.js"
  }
}
```

## Reference

### Steps

### Commands

Commands in steps can be a few different types

#### Exec command

| Field        | Description                                             |
| :----------- | :------------------------------------------------------ |
| `cmd`        | The shell command to execute                            |
| `caches`     | List of cache IDs to use when this command is executing |
| `customName` | The name to display when this command is running        |

If the command is a string, it is assumed to be an exec command in the format
`sh -c '<cmd>'`.

#### Variable command

#### Copy command

#### Path command

#### File command

### Start
