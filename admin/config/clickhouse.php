<?php

return [
    'url' => rtrim((string) env('CLICKHOUSE_URL', 'http://127.0.0.1:8123'), '/'),
    'database' => env('CLICKHOUSE_DATABASE', 'clicky'),
    'username' => env('CLICKHOUSE_USERNAME', 'clicky'),
    'password' => env('CLICKHOUSE_PASSWORD', 'clicky_local_password'),
    'timeout' => (int) env('CLICKHOUSE_TIMEOUT', 5),
];
