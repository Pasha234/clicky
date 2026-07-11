# Clicky Go Specification

## 1. Project Goal

Build a production-like click analytics system:

- PHP admin application for authentication, site management, tracking tokens, and analytics dashboards.
- Go collector service for high-throughput event ingestion.
- RabbitMQ or Kafka as a buffer between ingestion and storage.
- Go worker for batch processing and ClickHouse inserts.
- ClickHouse as the analytical database.
- Docker Compose for local development.
- Kubernetes manifests as an optional deployment layer.

The result should feel like a small but real analytics platform, not a simple CRUD demo.

## 2. Architecture

Main services:

- `admin`: PHP application, preferably Laravel or Symfony.
- `collector`: Go HTTP service that receives tracking events.
- `worker`: Go queue consumer that writes events to ClickHouse.
- `queue`: RabbitMQ for MVP, Kafka as an optional future alternative.
- `clickhouse`: analytics storage.
- `postgres`: relational storage for users, passwords, sites, tokens, and settings.
- `pgbouncer`: connection pool between application services and PostgreSQL.
- `redis`: token cache, rate limiting, and optional session/cache storage.

Recommended MVP data flow:

```text
Browser tracking script
  -> Go collector
  -> RabbitMQ
  -> Go worker
  -> ClickHouse
  -> PHP admin analytics dashboard
```

## 3. PHP Admin Application

Recommended stack:

- Laravel 11/12 or Symfony 7.
- PostgreSQL for users, passwords, sites, API tokens, and settings.
- PgBouncer as the only database endpoint for application traffic.
- ClickHouse client for analytics queries.
- Blade, Inertia, Livewire, Vue, or React for UI.
- Chart.js, ECharts, or ApexCharts for charts.

Responsibilities:

- User registration and login.
- Site/project creation.
- Tracking token generation and rotation.
- Tracking snippet display.
- Analytics dashboard rendering.
- ClickHouse aggregate queries.
- Access control and ownership checks.
- Token publishing to Redis for collector validation.

Core entities:

```text
User
Site
ApiToken
DashboardFilter
```

Optional entities:

```text
EventSchema
Team
TeamMember
AlertRule
```

Required pages:

```text
/login
/register
/dashboard
/sites
/sites/create
/sites/{id}
/sites/{id}/settings
/sites/{id}/snippet
/sites/{id}/analytics
```

Dashboard widgets:

- Total events.
- Events today.
- Timeline chart.
- Top pages.
- Top referrers.
- Event type breakdown.
- Browser/device/country breakdown.
- Optional click heatmap.

UI requirements:

- Clear navigation.
- Tables with pagination.
- Date range picker.
- Site and event filters.
- Loading states.
- Empty states.
- Error states.
- Responsive layout.

## 4. Go Collector Service

Purpose: receive tracking events quickly and publish them to the queue.

Required endpoints:

```http
GET  /collect
POST /collect
GET  /healthz
GET  /readyz
GET  /metrics
```

Example GET request:

```http
/collect?t=SITE_TOKEN&url=https://example.com/page&event=click&x=120&y=450
```

Example POST request:

```json
{
  "token": "site_token",
  "event": "click",
  "url": "https://example.com/pricing",
  "referrer": "https://google.com",
  "user_agent": "...",
  "ip": "auto/from request",
  "x": 120,
  "y": 450,
  "meta": {
    "button": "buy",
    "plan": "pro"
  },
  "timestamp": "2026-07-09T10:00:00Z"
}
```

Functional requirements:

- Accept GET and POST events.
- Validate required fields.
- Derive IP and user agent from request headers when missing.
- Validate the site token with a fast PostgreSQL query through PgBouncer.
- Apply request size limits.
- Apply rate limiting by token and IP.
- Publish valid events to RabbitMQ.
- Return fast success responses.
- Expose Prometheus metrics.
- Support graceful shutdown.

Recommended Go libraries:

```text
net/http
chi, echo, or fiber
zap or zerolog
prometheus/client_golang
rabbitmq/amqp091-go
redis/go-redis
jackc/pgx/v5
```

Collector metrics:

- Requests total.
- Requests by status.
- Request duration.
- Queue publish duration.
- Queue publish failures.
- Invalid events.
- Rate-limited requests.

## 5. Queue Layer

Use RabbitMQ for the MVP because it is simpler to operate locally.

Required queues:

```text
click_events
click_events_dead_letter
```

RabbitMQ requirements:

- Durable queue.
- Persistent messages.
- Publisher confirms.
- Dead-letter exchange.
- Retry strategy.
- Optional message TTL.
- Optional idempotency key.

Kafka can be added later behind a queue interface if the project needs partitioning and replay semantics.

## 6. Go Worker

Purpose: consume events from the queue in batches and insert them into ClickHouse.

Functional requirements:

- Consume messages from RabbitMQ.
- Decode and validate event messages.
- Batch events before inserting.
- Flush by batch size or flush interval.
- Retry ClickHouse insert failures.
- Nack or dead-letter invalid messages.
- Ack only after successful insert.
- Flush remaining events during graceful shutdown.
- Expose Prometheus metrics.

Example configuration:

```env
BATCH_SIZE=1000
FLUSH_INTERVAL=2s
WORKER_CONCURRENCY=4
CLICKHOUSE_DSN=tcp://clickhouse:9000
```

Worker metrics:

- Events consumed.
- Events inserted.
- Events failed.
- Batch size.
- Batch insert latency.
- Queue lag if available.
- ClickHouse errors.

## 7. ClickHouse Schema

Main events table:

```sql
CREATE TABLE events
(
    site_id UUID,
    token String,
    event_type LowCardinality(String),
    url String,
    referrer String,
    user_agent String,
    ip IPv4,
    country LowCardinality(String),
    city String,
    device LowCardinality(String),
    browser LowCardinality(String),
    os LowCardinality(String),
    x Nullable(UInt16),
    y Nullable(UInt16),
    meta String,
    created_at DateTime64(3)
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(created_at)
ORDER BY (site_id, event_type, created_at);
```

Optional materialized views:

```text
events_by_day
events_by_url
events_by_event_type
events_by_country
```

Analytics queries to support:

- Events by day/hour.
- Top pages.
- Top referrers.
- Top countries.
- Events by browser/device/OS.
- Click coordinates for heatmaps.
- Unique visitors.
- Event conversion funnel.

## 8. Tracking Script

The admin application must generate a JavaScript snippet for every site.

MVP snippet:

```html
<script>
(function () {
  const token = "SITE_TOKEN";

  document.addEventListener("click", function (event) {
    fetch("https://collector.example.com/collect", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      keepalive: true,
      body: JSON.stringify({
        token: token,
        event: "click",
        url: location.href,
        referrer: document.referrer,
        x: event.clientX,
        y: event.clientY,
        timestamp: new Date().toISOString()
      })
    });
  });
})();
</script>
```

Future tracking features:

- Page views.
- Custom events.
- Scroll depth.
- Session ID.
- Visitor ID.
- UTM parsing.
- SPA route change tracking.
- Error tracking.

## 9. API Contracts

PHP admin API:

```http
POST /api/sites
GET  /api/sites
GET  /api/sites/{id}
POST /api/sites/{id}/rotate-token
GET  /api/sites/{id}/analytics/summary
GET  /api/sites/{id}/analytics/timeline
GET  /api/sites/{id}/analytics/top-pages
GET  /api/sites/{id}/analytics/referrers
```

Go collector API:

```http
POST /collect
GET  /collect
GET  /healthz
GET  /readyz
GET  /metrics
```

Token validation options:

1. Go collector reads PostgreSQL through PgBouncer.
2. Go collector reads from Redis cache backed by PostgreSQL.
3. Go collector calls PHP API.
4. Worker resolves token later.

Recommended MVP approach:

```text
PHP creates, rotates, and revokes tokens in PostgreSQL.
Laravel connects to PgBouncer on port `6432`.
Go collector validates token existence and status with a parameterized SELECT through PgBouncer on port `6432`.
Redis may cache active-token lookups after the PostgreSQL path is working.
```

### PostgreSQL and PgBouncer

PostgreSQL is the source of truth for relational application data. ClickHouse stores only analytical events and aggregations; it must not be used for users, authentication, sites, API tokens, or application settings.

Connection flow:

```text
Laravel admin ---\
                 -> PgBouncer:6432 -> PostgreSQL:5432
Go collector ---/
```

Requirements:

- Do not expose PostgreSQL publicly in production; only PgBouncer is reachable by `admin` and `collector`.
- Use a dedicated application database user with the least required privileges.
- Store PostgreSQL and PgBouncer credentials in environment variables locally and Kubernetes Secrets in deployed environments.
- Configure PgBouncer in `transaction` pool mode for high concurrency.
- Application queries must not depend on session state, temporary tables, `SET` commands, or persistent prepared statements because connections can change between transactions.
- Laravel uses the PgBouncer host and port in `DB_HOST` and `DB_PORT`; Go uses the same endpoint in its `DATABASE_URL`.
- The collector checks that the token exists, is active, and belongs to an enabled site before publishing an event to RabbitMQ.
- Set database client pool limits in Laravel and Go below PgBouncer's pool capacity; PgBouncer limits must remain below PostgreSQL's `max_connections`.
- Expose PgBouncer health and pool metrics, including active, waiting, and idle client/server connections.

## 10. Infrastructure

Required Docker Compose services:

```yaml
services:
  admin:
  postgres:
  pgbouncer:
  redis:
  collector:
  worker:
  rabbitmq:
  clickhouse:
```

Optional observability services:

```yaml
services:
  prometheus:
  grafana:
```

Suggested Kubernetes layout:

```text
k8s/
  namespace.yaml
  admin-deployment.yaml
  admin-service.yaml
  collector-deployment.yaml
  collector-service.yaml
  worker-deployment.yaml
  rabbitmq-statefulset.yaml
  clickhouse-statefulset.yaml
  postgres-statefulset.yaml
  pgbouncer-deployment.yaml
  pgbouncer-service.yaml
  redis-deployment.yaml
  configmaps.yaml
  secrets.yaml
  ingress.yaml
```

## 11. Observability

Go services:

- `/metrics` endpoint.
- Structured JSON logs.
- Request latency metrics.
- Queue publish metrics.
- Batch insert metrics.
- Error counters.

PHP application:

- Request logs.
- Authentication logs.
- Failed job logs.
- ClickHouse query duration logs.
- PgBouncer connection and wait-time metrics.
- Analytics API latency.

Grafana dashboards:

- Collector RPS.
- Collector p95 latency.
- Queue depth.
- Worker insert rate.
- Worker failure rate.
- ClickHouse insert latency.
- ClickHouse query latency.
- PgBouncer client wait count and pool utilization.

## 12. Security

Requirements:

- User authentication.
- Per-user site ownership checks.
- Token rotation.
- CORS configuration.
- Rate limiting by token and IP.
- Max request body size.
- Event payload validation.
- Secrets only through environment variables or Kubernetes secrets.
- No raw SQL string interpolation.
- Internal services isolated in Docker/Kubernetes networks.
- Admin analytics endpoints protected by authorization.

## 13. Development Milestones

### Milestone 1: Local Skeleton

- Docker Compose file.
- PHP app starts.
- PostgreSQL and PgBouncer start; Laravel connects through PgBouncer.
- Go collector starts.
- RabbitMQ starts.
- ClickHouse starts.
- Basic health checks.

### Milestone 2: Admin MVP

- User registration and login.
- Site creation.
- Token generation.
- Token rotation.
- Tracking snippet page.
- Sites and tokens stored in SQL database.

### Milestone 3: Collector MVP

- `/collect` endpoint.
- GET and POST event support.
- Request validation.
- Queue publishing.
- Basic logs.
- Basic metrics.

### Milestone 4: Worker MVP

- RabbitMQ consumer.
- Event decoder.
- Batch insert into ClickHouse.
- Retry failed inserts.
- Graceful shutdown.

### Milestone 5: Analytics MVP

- PHP ClickHouse connection.
- Summary endpoint.
- Timeline endpoint.
- Top pages endpoint.
- Referrers endpoint.
- Dashboard charts.
- Date and site filters.

### Milestone 6: Production Features

- Redis token cache.
- Rate limiting.
- Dead-letter queue.
- Prometheus.
- Grafana.
- Load testing.
- Kubernetes manifests.

## 14. Testing Requirements

PHP feature tests:

- Registration and login.
- Site creation.
- Token rotation.
- Analytics endpoint authorization.

PHP unit tests:

- Token generator.
- ClickHouse query service.
- Dashboard filter parsing.

Go collector unit tests:

- Request parsing.
- Event validation.
- Token validation.
- PostgreSQL token lookup through PgBouncer.
- Queue publisher mock.

Go collector integration tests:

- Publish event to RabbitMQ.
- Validate active and revoked tokens through PgBouncer.
- Reject invalid token.
- Reject oversized payload.

Go worker unit tests:

- Batcher.
- Message decoder.
- Retry behavior.
- Invalid message handling.

Go worker integration tests:

- Consume from RabbitMQ.
- Insert into ClickHouse.
- Ack after successful insert.

Load testing tools:

```text
k6
vegeta
wrk
```

Suggested performance targets:

- 1,000 RPS stable for MVP.
- 10,000 RPS as stretch goal.
- Collector p95 latency under 50 ms without queue backpressure.
- Zero event loss during graceful shutdown.

## 15. Repository Structure

Recommended structure:

```text
clicky-go/
  admin/
  collector/
  worker/
  shared/
  docker/
  k8s/
  docs/
  docker-compose.yml
  Makefile
  README.md
  SPEC.md
```

## 16. Definition of Done

The MVP is complete when:

- A user can register and log in.
- A user can create a site.
- The system generates a tracking token.
- Laravel and Go collector connect to PostgreSQL through PgBouncer, not directly.
- The system displays a tracking snippet.
- A browser can send an event to the collector.
- The Go collector publishes the event to RabbitMQ.
- The Go worker writes the event to ClickHouse.
- The PHP admin dashboard displays charts from ClickHouse.
- The project starts with `docker compose up`.
- Basic PHP and Go tests exist.
- Health checks exist for services.
- Basic Prometheus metrics exist for Go services.
- README contains setup and usage instructions.
