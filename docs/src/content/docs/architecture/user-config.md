---
title: User Config
description: How Railpack handles user config
---

Users can configure Railpack in a few different ways:

- CLI flags
- Environment variables
- Config file

These configs are merged together and then applied to the generate context.

Everything that affects a part of the build plan _should_ be configurable.
Config affects the generate context rather than the plan itself as it allows
Railpack to perform optimizations after the config is applied. It also allows
the user config format to be abstracted at a higher level compared to the
relatively low level build plan schema.
