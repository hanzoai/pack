---
title: PHP
description: Building PHP applications with Railpack
---

Railpack can automatically build and deploy PHP applications with Nginx and
PHP-FPM.

## Detection

Your project will be detected as a PHP application if any of these
conditions are met:

- An `index.php` file exists in the root directory
- A `composer.json` file exists in the root directory

## Versions

The PHP version is determined in the following order:

- Read from the `composer.json` file
- Defaults to `8.4.3`

## Configuration

Railpack will configure Nginx and PHP-FPM for your application.
For Laravel applications, the document root is set to the `public`
directory.

### Environment Variables

| Variable                | Description                          | Example       |
| ----------------------- | ------------------------------------ | ------------- |
| `RAILPACK_PHP_ROOT_DIR` | Override the document root for Nginx | `/app/public` |

### Custom Configuration

You can provide custom configuration files in your project root:

- `nginx.conf` or `nginx.template.conf` - Custom Nginx configuration
- `php-fpm.conf` or `php-fpm.template.conf` - Custom PHP-FPM configuration

## Laravel Support

Laravel applications are detected by the presence of an `artisan`
file. When detected:

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

You can see the [node docs](/languages/node) for information on how to
configure node.
