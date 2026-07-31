#!/usr/bin/env bash
set -Eeuo pipefail

# Start the full local stack with one command. Every default matches
# docker-compose.yml, while values already exported by the caller win.
project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$project_dir"

export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-clicky_local_password}"
export ADMIN_APP_KEY="${ADMIN_APP_KEY:-base64:AfMVYqY0E5mP2rTJckcM6mEtgm13VblWPMYrt0W1P7A=}"
export GRAFANA_ADMIN_USER="${GRAFANA_ADMIN_USER:-admin}"
export GRAFANA_ADMIN_PASSWORD="${GRAFANA_ADMIN_PASSWORD:-admin}"
export RATE_LIMIT_PER_MINUTE="${RATE_LIMIT_PER_MINUTE:-120}"
export TOKEN_CACHE_TTL="${TOKEN_CACHE_TTL:-5m}"
export COLLECTOR_CORS_ORIGINS="${COLLECTOR_CORS_ORIGINS:-*}"

# Keep forwarded-IP trust disabled by default. Enable it only when testing a
# trusted proxy locally; never use 0.0.0.0/0 in a real deployment.
export COLLECTOR_TRUSTED_PROXIES="${COLLECTOR_TRUSTED_PROXIES:-}"

# GeoIP is optional. If the usual local GeoLite2 file exists, enable it
# automatically; otherwise the worker still enriches browser/device/OS.
export GEOIP_DATABASE_HOST_PATH="${GEOIP_DATABASE_HOST_PATH:-$project_dir/docker/geoip}"
if [[ -z "${GEOIP_DATABASE_PATH:-}" && -f "$GEOIP_DATABASE_HOST_PATH/GeoLite2-City.mmdb" ]]; then
    export GEOIP_DATABASE_PATH="/geoip/GeoLite2-City.mmdb"
else
    export GEOIP_DATABASE_PATH="${GEOIP_DATABASE_PATH:-}"
fi

case "${1:-up}" in
    up)
        docker compose up --build --detach --remove-orphans
        ;;
    rebuild)
        docker compose build --no-cache
        docker compose up --detach --force-recreate --remove-orphans
        ;;
    *)
        echo "Usage: $0 [up|rebuild]" >&2
        echo "  up       Build changed images and start the full stack (default)." >&2
        echo "  rebuild  Rebuild every image without cache and recreate every service." >&2
        exit 2
        ;;
esac

docker compose ps
