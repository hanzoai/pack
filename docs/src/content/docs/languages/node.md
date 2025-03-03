---
title: Node.js
description: Building Node.js applications with Railpack
---

Railpack builds and deploys Node.js applications with support for various
package managers and frameworks.

## Detection

Your project will be detected as a Node.js application if a `package.json` file
exists in the root directory.

## Versions

The Node.js version is determined in the following order:

- Set via the `RAILPACK_NODE_VERSION` environment variable
- Read from the `engines` field in `package.json`
- Read from the `.nvmrc` file
- Defaults to `22`

## Runtime Variables

These variables are available at runtime:

```
NODE_ENV=production
NPM_CONFIG_PRODUCTION=false
NPM_CONFIG_UPDATE_NOTIFIER=false
NPM_CONFIG_FUND=false
YARN_PRODUCTION=false
CI=true
```

## Configuration

Railpack builds your Node.js application based on your project structure. The
build process:

- Installs dependencies using your preferred package manager (npm, yarn, pnpm,
  or bun)
- Executes the build script if defined in `package.json`
- Sets up the start command based on your project configuration

Railpack determines the start command in the following order:

1. The `start` script in `package.json`
2. The `main` field in `package.json`
3. An `index.js` or `index.ts` file in the root directory

### Package Managers

Railpack detects your package manager based on lock files:

- `pnpm-lock.yaml` for pnpm
- `bun.lockb` or `bun.lock` for Bun
- `.yarnrc.yml` or `.yarnrc.yaml` for Yarn 2
- `yarn.lock` for Yarn 1
- Defaults to npm if no lock file is found

### Config Variables

| Variable                  | Description                             | Example |
| ------------------------- | --------------------------------------- | ------- |
| `RAILPACK_NODE_VERSION`   | Override the Node.js version            | `20`    |
| `RAILPACK_BUN_VERSION`    | Override the Bun version                | `1.0.0` |
| `RAILPACK_SPA_OUTPUT_DIR` | Directory containing built static files | `dist`  |
| `PRUNE_DEPS`              | Remove development dependencies         | `true`  |

## Static Sites

Railpack supports building static sites with Vite and Astro:

- **Vite**: Detects Vite projects by the presence of `vite.config.js` or
  `vite.config.ts` or a `vite build` in the `package.json` build script
- **Astro**: Detects Astro projects by the presence of `astro.config.js`

For both frameworks, Railpack will try to detect the output directory and will
default to `dist`. Set the `RAILPACK_SPA_OUTPUT_DIR` environment variable to
specify a custom output directory.

Static sites are served using the [Caddy](https://caddyserver.com/) web server
and a [default
Caddyfile](https://github.com/railwayapp/railpack/blob/main/core/providers/node/Caddyfile.template).
You can overwrite this file with your own Caddyfile at the root of your project.

## Framework Support

Railpack detects and configures caches and commands for popular frameworks.
Including:

- Next.js
- Vite
- Astro
