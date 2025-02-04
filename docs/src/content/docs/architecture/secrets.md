---
title: Secrets and Environment Variables
description: How Railpack handles secrets and environment variables
---

Build secrets and environment variables are treated separatley. The main differences being

- Environment variables are saved in the final image and should not contain
  sensitive information. Since they are in the final image, providers can add
  variables that will be available to the app at runtime.
- Secrets are never logged or saved in the build logs. They are also only
  available at build time and not saved to the final image.

## Environment Variables

Environment variables are added to the build plan as a step command. This allows
the providers to control exactly when to bust the layer cache (in contrast to
Nixpacks which treats env vars as all or nothing at the start of the build).

```json
{
  "steps": {
    "install": {
      "commands": [
        { "name": "NODE_ENV", "value": "production" }
        // ...
      ]
      // ...
    }
  }
}
```

## Secrets

The names of all secrets that should be used during the build are added to the
top of the build plan. Whether or not a step uses the secrets is determined by a
`useSecrets` key. If this key is present, the secrets will be available to all
exec commands run in the step.

Under the hood, Railpack uses [BuildKit secrets
mounts](https://docs.docker.com/build/building/secrets/) to supply an exec
command with the secret value as an environment variable.

```json
{
  "secrets": ["STRIPE_LIVE_KEY"],

  "steps": {
    "build": {
      "useSecrets": true
      // ...
    }
  }
}
```

### Providing Secrets

You can add secrets when building or generating a build plan with the `--env`
flag. The names of these variables will be added to the build plan as secrets.

#### CLI Build

If building with [the CLI](/guides/building-with-cli), Railpack will check that
all the secrets defined in the build plan have variables.

```bash
railpack build --env STRIPE_LIVE_KEY=sk_live_asdf
```

#### Custom Frontend

If building with a [custom frontend](/guides/building-with-custom-frontends),
you should still provide the secrets when generating the plan with `--env`. This
adds the secrets to the build plan. You then need to pass the secrets to Docker
or BuildKit with the `--secret` flag.

```bash
# Generate a build plan
railpack plan --env STRIPE_LIVE_KEY=sk_live_asdf --out test/railpack-plan.json

# Build with the custom frontend
STRIPE_LIVE_KEY=sk_live_asdf docker build \
  --build-arg BUILDKIT_SYNTAX="ghcr.io/railwayapp/railpack:railpack-frontend" \
  -f test/railpack-plan.json \
  --secret id=STRIPE_LIVE_KEY,env=STRIPE_LIVE_KEY \
  examples/node-bun
```

### Secret hash

By default, BuildKit will not invalidate the a layer if a secret is changed. To
get around this, Railpack uses a hash of the secret values and mounts this as a
file in the layer. This will bust the layer cache if the secret is changed.

If using the CLI to build, this will happen automatically.

If using a custom frontend, you will need to provide the secret hash manually
via the `--opt secrets-hash=<hash>` flag.

```bash
STRIPE_LIVE_KEY=sk_live_asdf docker build \
  --build-arg BUILDKIT_SYNTAX="ghcr.io/railwayapp/railpack:railpack-frontend" \
  -f test/railpack-plan.json \
  --secret id=STRIPE_LIVE_KEY,env=STRIPE_LIVE_KEY \
  --opt secrets-hash=asdf123456789... \
  examples/node-bun
```

This value can be anything that indicates that the secrets have changed (a
simple counter also works). However, we recommend using a non-reversible hash of
the secret values.
