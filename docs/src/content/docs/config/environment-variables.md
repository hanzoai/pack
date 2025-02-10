---
title: Environment Variables
description: Understanding environment variables in Railpack
---

Some parts of the build can be configured with environment variables. These are
often prefixed with `RAILPACK_`.

| Name           | Description                                                                                                |
| :------------- | :--------------------------------------------------------------------------------------------------------- |
| `INSTALL_CMD`  | Set the `steps.install.command` to the value                                                               |
| `BUILD_CMD`    | Set the `steps.build.commands` to the value                                                                |
| `START_CMD`    | Set the `start.cmd` to the value                                                                           |
| `PACKAGES`     | Install additional Mise packages. In the format `pkg@version`. The latest version is used if not provided. |
| `APT_PACKAGES` | Install additional Apt packages. e.g. `build-essential`                                                    |

To configure more parts of the build, it is recommended to use a [config file](/config/file).
