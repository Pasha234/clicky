<?php

use App\Models\Site;
use App\Models\User;
use Illuminate\Support\Facades\Http;

function createAnalyticsSite(User $user): Site
{
    return Site::query()->create([
        'user_id' => $user->id,
        'name' => 'Analytics site',
        'domain' => 'analytics.test',
        'enabled' => true,
    ]);
}

test('an owner can request an analytics summary', function () {
    $user = User::factory()->create();
    $site = createAnalyticsSite($user);

    Http::fake([
        '*' => Http::response([
            'data' => [[
                'events' => '12',
                'clicks' => '10',
                'unique_visitors' => '5',
            ]],
        ]),
    ]);

    $response = $this->actingAs($user)->getJson(
        "/api/sites/{$site->id}/analytics/summary?from=2026-07-01&to=2026-07-17",
    );

    $response->assertOk()->assertExactJson([
        'events' => 12,
        'clicks' => 10,
        'unique_visitors' => 5,
    ]);

    Http::assertSentCount(1);
});

test('a user cannot request analytics for a site they do not own', function () {
    $owner = User::factory()->create();
    $otherUser = User::factory()->create();
    $site = createAnalyticsSite($owner);

    Http::fake();

    $this->actingAs($otherUser)
        ->getJson("/api/sites/{$site->id}/analytics/summary")
        ->assertNotFound();

    Http::assertNothingSent();
});

test('analytics endpoints validate date filters before querying ClickHouse', function () {
    $user = User::factory()->create();
    $site = createAnalyticsSite($user);

    Http::fake();

    $this->actingAs($user)
        ->getJson("/api/sites/{$site->id}/analytics/timeline?from=not-a-date")
        ->assertUnprocessable()
        ->assertJsonValidationErrors('from');

    Http::assertNothingSent();
});
