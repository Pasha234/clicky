<?php

use App\Models\Site;
use App\Models\User;
use App\Services\Analytics\ClickHouseAnalytics;
use Carbon\CarbonImmutable;

test('analytics filters parse an inclusive date range into an exclusive end date', function () {
    $user = User::factory()->create();
    $site = Site::query()->create([
        'user_id' => $user->id,
        'name' => 'Example',
        'domain' => 'example.test',
        'enabled' => true,
    ]);

    $filter = app(ClickHouseAnalytics::class)->filterFor($user, [
        'site_id' => $site->id,
        'from' => '2026-07-01',
        'to' => '2026-07-17',
    ]);

    expect($filter)->not->toBeNull()
        ->and($filter->from->toDateTimeString())->toBe('2026-07-01 00:00:00')
        ->and($filter->toExclusive->toDateTimeString())->toBe('2026-07-18 00:00:00');
});

test('analytics filters use the last thirty days when dashboard dates are absent', function () {
    CarbonImmutable::setTestNow('2026-07-17 12:00:00');

    try {
        $user = User::factory()->create();
        $site = Site::query()->create([
            'user_id' => $user->id,
            'name' => 'Example',
            'domain' => 'example.test',
            'enabled' => true,
        ]);

        $filter = app(ClickHouseAnalytics::class)->filterFor($user, ['site_id' => $site->id]);

        expect($filter->from->toDateString())->toBe('2026-06-18')
            ->and($filter->toExclusive->toDateString())->toBe('2026-07-18');
    } finally {
        CarbonImmutable::setTestNow();
    }
});
