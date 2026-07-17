#!/bin/sh

set -eu

client() {
  clickhouse-client \
    --host "$CLICKHOUSE_HOST" \
    --user "$CLICKHOUSE_USER" \
    --password "$CLICKHOUSE_PASSWORD" \
    --database "$CLICKHOUSE_DATABASE" \
    "$@"
}

client --multiquery <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations
(
    version String,
    applied_at DateTime DEFAULT now()
)
ENGINE = MergeTree
ORDER BY version;
SQL

for file in /migrations/*.sql; do
  [ -f "$file" ] || continue

  version=$(basename "$file" .sql)

  case "$version" in
    ''|*[!A-Za-z0-9_-]*)
      echo "Invalid migration filename: $file" >&2
      exit 1
      ;;
  esac

  applied=$(client --query "SELECT count() FROM schema_migrations WHERE version = '$version'")

  if [ "$applied" -gt 0 ]; then
    echo "Skipping migration $version (already applied)"
    continue
  fi

  echo "Applying migration $version"
  client --multiquery < "$file"
  client --query "INSERT INTO schema_migrations (version) VALUES ('$version')"
done
