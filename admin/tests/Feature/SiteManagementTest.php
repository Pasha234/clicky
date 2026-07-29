<?php

use App\Models\Site;
use App\Models\User;

test('creating a site creates an active tracking token', function () {
    $user = User::factory()->create();

    $site = Site::query()->create([
        'user_id' => $user->id,
        'name' => 'Example site',
        'domain' => 'example.test',
        'enabled' => true,
    ]);

    $token = $site->activeToken()->first();

    expect($token)->not->toBeNull()
        ->and($token->site_id)->toBe($site->id)
        ->and($token->token)->toMatch('/^clk_[a-f0-9]{64}$/')
        ->and($token->revoked_at)->toBeNull();
});

test('rotating a token revokes the previous token and creates a new active token', function () {
    $user = User::factory()->create();
    $site = Site::query()->create([
        'user_id' => $user->id,
        'name' => 'Example site',
        'domain' => 'example.test',
        'enabled' => true,
    ]);
    $oldToken = $site->activeToken()->firstOrFail();

    $newToken = $site->rotateToken();

    expect($newToken->id)->not->toBe($oldToken->id)
        ->and($newToken->token)->not->toBe($oldToken->token)
        ->and($newToken->revoked_at)->toBeNull();

    expect($oldToken->refresh()->revoked_at)->not->toBeNull()
        ->and($site->fresh()->activeToken?->id)->toBe($newToken->id);
});
