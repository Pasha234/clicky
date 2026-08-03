# Clicky Go

Clicky Go is a small self-hosted click analytics platform. A browser tracking
snippet sends page views and clicks to a Go collector. Events are buffered in
RabbitMQ, written to ClickHouse by a worker, and displayed in a Laravel +
Filament dashboard.

```text
Browser snippet -> Collector -> RabbitMQ -> Worker -> ClickHouse -> Admin dashboard
```

## Tech stack

- Go and Fiber — public event collector and asynchronous worker.
- RabbitMQ — durable event buffer and dead-letter queue.
- ClickHouse — analytics event store and query engine.
- PostgreSQL through PgBouncer — users, sites, and tracking tokens.
- Redis — tracking-token cache and collector rate limiting.
- Laravel 12 and Filament — registration, site management, and analytics UI.
- Prometheus and Grafana — metrics collection and visualisation in Docker Compose.
- Docker Compose — local development stack.
- Kubernetes manifests and Kustomize — deployment starting point for Minikube or a cluster.

## Services

| Service | Local address | What it does |
| --- | --- | --- |
| Admin | http://localhost:8000 | Laravel registration/login, site management, snippets, and analytics dashboard. |
| Collector | http://localhost:3000 | Public Go HTTP API which validates tokens, rate-limits requests, and publishes events. |
| Worker | http://localhost:3001 | Consumes events, enriches them, batches ClickHouse inserts, and exposes health/metrics. |
| RabbitMQ | http://localhost:15672 | Event queue and dead-letter queue; local credentials are `clicky` / `clicky_local_password`. |
| ClickHouse | http://localhost:8123 | Analytics database HTTP endpoint. |
| PostgreSQL | `localhost:5432` | Application database; exposed for local development. |
| PgBouncer | `localhost:6432` | PostgreSQL connection-pool endpoint used by application services. |
| Redis | `localhost:6379` | Token cache and rate-limit counters. |
| Prometheus | http://localhost:9090 | Scrapes collector and worker metrics. |
| Grafana | http://localhost:3002 | Metrics UI; default credentials are `admin` / `admin`. |

## Run locally with Docker Compose

### Requirements

- Docker Engine with Docker Compose v2.

### Start

The helper builds changed images and starts the complete stack:

```bash
./docker-build.sh up
```

For a complete no-cache rebuild:

```bash
./docker-build.sh rebuild
```

The equivalent direct Compose command is:

```bash
docker compose up -d --build
docker compose ps
```

On its first run, `clickhouse-migrator` applies ClickHouse migrations and
exits. The admin container applies Laravel migrations and synchronises active
tracking tokens to Redis before serving requests.

Open http://localhost:8000/register, create an account, then create a site in
the admin panel. The site’s **Snippet** page contains the tracking code and an
active token.

### Stop or reset

```bash
docker compose down       # stops services; keeps local volumes
docker compose down -v    # also removes all local data
```

## Docker Compose configuration

Set these variables in a root `.env` file or export them before running
`docker-build.sh` or `docker compose`. Defaults are for local development.

| Variable | Default | Description |
| --- | --- | --- |
| `POSTGRES_PASSWORD` | `clicky_local_password` | Password shared by PostgreSQL, PgBouncer, and the admin application. |
| `ADMIN_APP_KEY` | local development key | Laravel encryption key. Generate a distinct secret for any non-local environment. |
| `GRAFANA_ADMIN_USER` | `admin` | Grafana administrator username. |
| `GRAFANA_ADMIN_PASSWORD` | `admin` | Grafana administrator password. |
| `RATE_LIMIT_PER_MINUTE` | `120` | Per-token and per-IP collector request limit. Use a high temporary value only for load tests. |
| `TOKEN_CACHE_TTL` | `5m` | Redis token-cache lifetime. Format is a positive whole `Ns`, `Nm`, or `Nh`, such as `30s`, `5m`, or `1h`. |
| `COLLECTOR_CORS_ORIGINS` | `*` | Comma-separated browser origins allowed to call the collector. Restrict this in production. |
| `COLLECTOR_TRUSTED_PROXIES` | empty | Comma-separated trusted proxy CIDRs/IPs for forwarded client IP handling. Keep empty locally; never trust all addresses in production. |
| `GEOIP_DATABASE_HOST_PATH` | `./docker/geoip` | Host directory mounted read-only into the worker at `/geoip`. |
| `GEOIP_DATABASE_PATH` | empty, or auto-detected by `docker-build.sh` | Path inside the worker container to the GeoLite2 City database. |

The remaining service connection values are deliberately internal Compose
addresses in `docker-compose.yml`. Override them only if you are adapting the
stack rather than using it locally.

### Optional GeoIP database

The worker always derives browser, device, and operating system from the user
agent. Country and city require a local MaxMind GeoLite2 City database; no
event data is sent to a third-party service.

1. Download `GeoLite2-City.mmdb` from MaxMind after creating a GeoLite account.
2. Place it at `docker/geoip/GeoLite2-City.mmdb`.
3. Start the stack with the helper, which detects that location automatically:

```bash
./docker-build.sh up
```

Or configure another directory explicitly:

```bash
GEOIP_DATABASE_HOST_PATH=/absolute/path/to/geoip \
GEOIP_DATABASE_PATH=/geoip/GeoLite2-City.mmdb \
./docker-build.sh up
```

The `.mmdb` file is ignored by Git. Without it, the worker remains operational
and stores empty country/city values.

## Endpoints

### Admin and analytics API

The web pages use a session-based Laravel login.

| Method | Path | Description |
| --- | --- | --- |
| `GET`, `POST` | `/register` | Registration page and account creation. |
| `GET`, `POST` | `/login` | Login page and session creation. |
| `POST` | `/logout` | Ends the authenticated session. |
| `GET` | `/admin` | Filament dashboard. |
| `GET` | `/admin/sites` | Site list and create flow. |
| `GET` | `/admin/sites/{site}` | Site overview. |
| `GET` | `/admin/sites/{site}/settings` | Site settings and token controls. |
| `GET` | `/admin/sites/{site}/snippet` | Tracking snippet with copy button. |
| `GET` | `/admin/sites/{site}/analytics` | Per-site analytics UI. |
| `GET` | `/api/sites/{site}/analytics/summary` | Event, click, and unique-visitor totals. |
| `GET` | `/api/sites/{site}/analytics/timeline` | Daily event counts. |
| `GET` | `/api/sites/{site}/analytics/top-pages` | Most visited pages. |
| `GET` | `/api/sites/{site}/analytics/referrers` | Top non-empty referrers. |

Analytics API routes require authentication and only return sites owned by the
current user. They accept optional `from` and `to` query parameters in
`YYYY-MM-DD` format; the default range is the last 30 days.

### Collector

Every collection request requires an active tracking token.

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/collect` | Collects an event from query parameters. |
| `POST` | `/collect` | Collects an event from JSON. |
| `GET` | `/healthz` | Process liveness check. |
| `GET` | `/readyz` | Dependency readiness: RabbitMQ, token validation, and Redis. |
| `GET` | `/metrics` | Prometheus metrics. |

Example JSON event:

```bash
curl http://localhost:3000/collect \
  -H 'Content-Type: application/json' \
  -d '{
    "token": "YOUR_TRACKING_TOKEN",
    "event": "click",
    "url": "https://example.test/pricing",
    "referrer": "https://google.com",
    "x": 120,
    "y": 450,
    "meta": {"button": "buy"}
  }'
```

A valid event returns `202 Accepted`.

### Worker

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `http://localhost:3001/healthz` | Worker process liveness check. |
| `GET` | `http://localhost:3001/metrics` | Prometheus metrics: consumed/inserted/failed events, batch timings, retries, and queue lag. |

## Benchmark

Local k6 benchmark, 2026-07-31:

| Metric | Result |
| --- | --- |
| Accepted events | 1,050,014 |
| HTTP failures | 0% |
| Collector latency p95 / p99 | 6.2 ms / 12.23 ms |
| Collector and worker health failures | 0% |
| Planned peak request rate | 15,000 requests/second |

The collector accepted all requests received during the benchmark. Collector
and worker health checks stayed successful throughout the run.

## Tests

### PHP

The admin test suite uses Laravel/Pest. Start the local stack, then run:

```bash
docker compose exec admin php artisan test
```

### Go unit tests

These tests do not require Docker services:

```bash
(cd collector && go test ./...)
(cd worker && go test ./...)
```

### Go integration tests

Integration tests use Testcontainers to create isolated PostgreSQL, PgBouncer,
RabbitMQ, and ClickHouse containers. Docker must be running. They do not use
or modify the local Compose database.

```bash
(cd collector && RUN_INTEGRATION_TESTS=1 go test -tags=integration ./...)
(cd worker && RUN_INTEGRATION_TESTS=1 go test -tags=integration ./...)
```

## Kubernetes with Minikube

The `k8s/` directory contains Kustomize manifests for PostgreSQL, PgBouncer,
Redis, RabbitMQ, ClickHouse, migration jobs, collector, worker, admin, and an
Ingress. Prometheus and Grafana are currently provided by Docker Compose only.

### 1. Start Minikube and build images

```bash
minikube start --driver=docker

# Build the three application images locally.
docker compose build admin collector worker

# Make those images available to Minikube.
minikube image load clicky-go-admin:latest
minikube image load clicky-go-collector:latest
minikube image load clicky-go-worker:latest
```

### 2. Configure secrets and host settings

Before applying anything, replace every `change-me` value in
`k8s/secrets.yaml`, including the passwords embedded in the connection URLs.
Generate a unique Laravel key, for example:

```bash
printf 'base64:'
openssl rand -base64 32
```

Set the generated value as `ADMIN_APP_KEY`. Also edit `k8s/configmaps.yaml`:

- set `APP_URL` and the Ingress host to your intended hostname;
- restrict `COLLECTOR_CORS_ORIGINS` for a real deployment;
- replace `COLLECTOR_TRUSTED_PROXIES` with the actual ingress-controller
  network ranges.

### 3. Deploy and inspect

```bash
minikube addons enable ingress
kubectl apply -k k8s/

kubectl -n clicky get pods
kubectl -n clicky get jobs
kubectl -n clicky wait --for=condition=complete job/clickhouse-migrations --timeout=5m
kubectl -n clicky wait --for=condition=complete job/admin-migrations --timeout=5m
```

For the simplest local access, forward ports in separate terminals:

```bash
kubectl -n clicky port-forward service/admin 8000:8000
kubectl -n clicky port-forward service/collector 3000:3000
```

Then use http://localhost:8000 and http://localhost:3000. To use the Ingress,
point its configured hostname at `minikube ip` (and run `minikube tunnel` when
your driver/network requires it).
