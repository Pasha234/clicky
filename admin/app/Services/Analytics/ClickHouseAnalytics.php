<?php

namespace App\Services\Analytics;

use App\Models\Site;
use App\Models\User;
use Carbon\CarbonImmutable;
use Illuminate\Http\Client\PendingRequest;
use Illuminate\Support\Facades\Http;
use RuntimeException;

final class ClickHouseAnalytics
{
    public function filterFor(User $user, array $input): ?AnalyticsFilter
    {
        $siteId = $input['site_id'] ?? null;

        if (! is_string($siteId) || $siteId === '') {
            return null;
        }

        $site = $user->sites()->find($siteId);

        if (! $site instanceof Site) {
            return null;
        }

        $from = CarbonImmutable::parse($input['from'] ?? now()->subDays(29)->toDateString())
            ->startOfDay();
        $toExclusive = CarbonImmutable::parse($input['to'] ?? now()->toDateString())
            ->startOfDay()
            ->addDay();

        if ($toExclusive->lessThanOrEqualTo($from)) {
            return null;
        }

        return new AnalyticsFilter($site, $from, $toExclusive);
    }

    /** @return array{events: int, clicks: int, unique_visitors: int} */
    public function summary(AnalyticsFilter $filter): array
    {
        $row = $this->selectOne(<<<'SQL'
            SELECT
                count() AS events,
                countIf(event_type = 'click') AS clicks,
                uniqCombined64(ip) AS unique_visitors
            FROM events
            WHERE site_id = {site_id:UUID}
              AND created_at >= parseDateTimeBestEffort({from:String})
              AND created_at < parseDateTimeBestEffort({to:String})
            FORMAT JSON
            SQL, $filter);

        return [
            'events' => (int) ($row['events'] ?? 0),
            'clicks' => (int) ($row['clicks'] ?? 0),
            'unique_visitors' => (int) ($row['unique_visitors'] ?? 0),
        ];
    }

    /** @return list<array{date: string, events: int}> */
    public function timeline(AnalyticsFilter $filter): array
    {
        $rows = $this->select(<<<'SQL'
            SELECT toDate(created_at) AS date, count() AS events
            FROM events
            WHERE site_id = {site_id:UUID}
              AND created_at >= parseDateTimeBestEffort({from:String})
              AND created_at < parseDateTimeBestEffort({to:String})
            GROUP BY date
            ORDER BY date
            FORMAT JSON
            SQL, $filter);

        $eventsByDate = collect($rows)
            ->mapWithKeys(fn (array $row): array => [(string) $row['date'] => (int) $row['events']]);

        $timeline = [];
        for ($date = $filter->from; $date->lessThan($filter->toExclusive); $date = $date->addDay()) {
            $day = $date->toDateString();
            $timeline[] = ['date' => $day, 'events' => $eventsByDate->get($day, 0)];
        }

        return $timeline;
    }

    /** @return list<array{url: string, events: int}> */
    public function topPages(AnalyticsFilter $filter): array
    {
        return array_map(fn (array $row): array => [
            'url' => (string) $row['url'],
            'events' => (int) $row['events'],
        ], $this->select(<<<'SQL'
            SELECT url, count() AS events
            FROM events
            WHERE site_id = {site_id:UUID}
              AND created_at >= parseDateTimeBestEffort({from:String})
              AND created_at < parseDateTimeBestEffort({to:String})
            GROUP BY url
            ORDER BY events DESC
            LIMIT 10
            FORMAT JSON
            SQL, $filter));
    }

    /** @return list<array{referrer: string, events: int}> */
    public function referrers(AnalyticsFilter $filter): array
    {
        return array_map(fn (array $row): array => [
            'referrer' => (string) $row['referrer'],
            'events' => (int) $row['events'],
        ], $this->select(<<<'SQL'
            SELECT referrer, count() AS events
            FROM events
            WHERE site_id = {site_id:UUID}
              AND referrer != ''
              AND created_at >= parseDateTimeBestEffort({from:String})
              AND created_at < parseDateTimeBestEffort({to:String})
            GROUP BY referrer
            ORDER BY events DESC
            LIMIT 10
            FORMAT JSON
            SQL, $filter));
    }

    /** @return list<array<string, mixed>> */
    private function select(string $sql, AnalyticsFilter $filter): array
    {
        $response = $this->client()->withOptions([
            'query' => [
                'database' => config('clickhouse.database'),
                'param_site_id' => $filter->site->getKey(),
                'param_from' => $filter->from->toDateTimeString(),
                'param_to' => $filter->toExclusive->toDateTimeString(),
            ],
        ])->withBody($sql, 'text/plain')->post('/');

        $response->throw();
        $decoded = json_decode($response->body(), true);

        if (! is_array($decoded) || ! is_array($decoded['data'] ?? null)) {
            throw new RuntimeException('ClickHouse returned an invalid JSON response.');
        }

        return $decoded['data'];
    }

    /** @return array<string, mixed> */
    private function selectOne(string $sql, AnalyticsFilter $filter): array
    {
        return $this->select($sql, $filter)[0] ?? [];
    }

    private function client(): PendingRequest
    {
        return Http::baseUrl(config('clickhouse.url'))
            ->withBasicAuth(config('clickhouse.username'), config('clickhouse.password'))
            ->timeout(config('clickhouse.timeout'));
    }
}
