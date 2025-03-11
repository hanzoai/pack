#!/bin/bash

set -e

if [ "$IS_LARAVEL" = "true" ]; then
  echo "Running migrations and seeding database ..."
  php artisan migrate --isolated --seed --force || php artisan migrate --seed --force

  php artisan storage:link
  php artisan optimize:clear
  php artisan optimize
fi

# Start the FrankenPHP server
docker-php-entrypoint --config /Caddyfile --adapter caddyfile 2>&1
