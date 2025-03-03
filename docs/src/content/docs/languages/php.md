---
title: PHP
description: Building PHP applications with Railpack
---

Railpack can automatically build and deploy PHP applications with Nginx and
PHP-FPM.

## Detection

Your project will be detected as a PHP application if any of these conditions
are met:

- An `index.php` file exists in the root directory
- A `composer.json` file exists in the root directory

## Versions

The PHP version is determined in the following order:

- Read from the `composer.json` file
- Defaults to `8.4.3`

## Configuration

Railpack will configure Nginx and PHP-FPM for your application. For Laravel
applications, the document root is set to the `public` directory.

### Config Variables

| Variable                | Description                          | Example       |
| ----------------------- | ------------------------------------ | ------------- |
| `RAILPACK_PHP_ROOT_DIR` | Override the document root for Nginx | `/app/public` |

### Custom Configuration

Railpack uses a custom [Nginx
config](https://github.com/railwayapp/railpack/blob/main/core/providers/php/nginx.template.conf)
file and [PHP-FPM
config](https://github.com/railwayapp/railpack/blob/main/core/providers/php/php-fpm.template.conf)
file. You can overwrite these files with your own configuration files in your
project root.

## Laravel Support

Laravel applications are detected by the presence of an `artisan` file. When
detected:

- The document root is set to the `/app/public` directory
- Storage directory permissions are set to be writable
- Composer dependencies are installed

## Node.js Integration

If a `package.json` file is detected in your PHP project:

- Node.js will be installed
- NPM dependencies will be installed
- Build scripts defined in `package.json` will be executed
- Development dependencies will be pruned in the final image

This is particularly useful for Laravel applications that use frontend
frameworks like Vue.js or React.

You can see the [node docs](/languages/node) for information on how to configure
node.
