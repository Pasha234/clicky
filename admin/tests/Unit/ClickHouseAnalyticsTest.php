<?php

use App\Models\Site;
use App\Services\Analytics\AnalyticsFilter;
use App\Services\Analytics\ClickHouseAnalytics;
use Carbon\CarbonImmutable;
use Illuminate\Support\Facades\Http;

function analyticsFilter(): AnalyticsFilter
{
    return new AnalyticsFilter(
        new Site(['id' => '11111111-1111-1111-1111-111111111111']),
        CarbonImmutable::parse('2026-07-01')->startOfDay(),
        CarbonImmutable::parse('2026-07-04')->startOfDay(),
    );
}

test('summary maps ClickHouse numeric strings to integers', function () {
    Http::fake(['*' => Http::response(['data' => [[
        'events' => '12', 'clicks' => '10', 'unique_visitors' => '5',
    ]]])]);

    expect(app(ClickHouseAnalytics::class)->summary(analyticsFilter()))->toBe([
        'events' => 12, 'clicks' => 10, 'unique_visitors' => 5,
    ]);
    Http::assertSentCount(1);
});

test('timeline fills dates for which ClickHouse has no events', function () {
    Http::fake(['*' => Http::response(['data' => [
        ['date' => '2026-07-01', 'events' => '4'],
        ['date' => '2026-07-03', 'events' => '2'],
    ]])]);

    expect(app(ClickHouseAnalytics::class)->timeline(analyticsFilter()))->toBe([
        ['date' => '2026-07-01', 'events' => 4],
        ['date' => '2026-07-02', 'events' => 0],
        ['date' => '2026-07-03', 'events' => 2],
    ]);
});

test('top pages and referrers expose typed API data', function () {
    Http::fakeSequence()
        ->push(['data' => [['url' => 'https://example.test/', 'events' => '7']]])
        ->push(['data' => [['referrer' => 'https://search.test/', 'events' => '3']]]);

    $analytics = app(ClickHouseAnalytics::class);

    expect($analytics->topPages(analyticsFilter()))->toBe([
        ['url' => 'https://example.test/', 'events' => 7],
    ])->and($analytics->referrers(analyticsFilter()))->toBe([
        ['referrer' => 'https://search.test/', 'events' => 3],
    ]);
    Http::assertSentCount(2);
});

test('breakdowns map labels and numeric ClickHouse values', function () {
    Http::fake(['*' => Http::response(['data' => [
        ['label' => 'Chrome', 'events' => '9'],
        ['label' => 'Unknown', 'events' => '2'],
    ]])]);

    expect(app(ClickHouseAnalytics::class)->breakdown(analyticsFilter(), 'browser'))->toBe([
        ['label' => 'Chrome', 'events' => 9],
        ['label' => 'Unknown', 'events' => 2],
    ]);

    Http::assertSent(function ($request): bool {
        return str_contains($request->body(), "if(browser = '', 'Unknown', browser)");
    });
});
