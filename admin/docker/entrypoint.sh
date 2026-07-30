#!/bin/sh
set -eu

# Laravel (and Laravel Boost) expects a physical .env file. Compose is the
# source of truth, so recreate that file from the same container environment
# on every start instead of baking a separate, drifting .env into the image.
umask 077

write_env() {
    name="$1"
    value="$(printenv "$name" 2>/dev/null || true)"
    printf '%s=%s\n' "$name" "$value"
}

{
    for name in \
        APP_ENV APP_DEBUG APP_URL APP_KEY \
        DB_CONNECTION DB_HOST DB_PORT DB_DATABASE DB_USERNAME DB_PASSWORD DB_SSLMODE \
        REDIS_CLIENT REDIS_HOST REDIS_PORT \
        SESSION_DRIVER CACHE_STORE \
        COLLECTOR_URL \
        CLICKHOUSE_URL CLICKHOUSE_DATABASE CLICKHOUSE_USERNAME CLICKHOUSE_PASSWORD
    do
        write_env "$name"
    done
} > .env

exec "$@"
