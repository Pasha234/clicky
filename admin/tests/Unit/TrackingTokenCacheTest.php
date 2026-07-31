<?php

use App\Models\ApiToken;
use App\Services\TrackingTokenCache;
use Illuminate\Support\Facades\Redis;

function cachedToken(): ApiToken
{
    $token = new ApiToken;
    $token->token = 'clk_'.str_repeat('a', 64);
    $token->prefix = substr($token->token, 0, 12);
    $token->site_id = '11111111-1111-1111-1111-111111111111';

    return $token;
}

test('it publishes an active token as a hashed Redis key', function () {
    config()->set('services.tracking_tokens.cache_ttl', '5m');

    $token = cachedToken();
    $connection = Mockery::mock();
    Redis::shouldReceive('connection')->once()->andReturn($connection);
    $connection->shouldReceive('setex')->once()->with(
        'collector:token:'.hash('sha256', $token->token),
        300,
        $token->site_id,
    );

    app(TrackingTokenCache::class)->publish($token);
});

test('it removes a token mapping when invalidated', function () {
    $token = cachedToken();
    $connection = Mockery::mock();
    Redis::shouldReceive('connection')->once()->andReturn($connection);
    $connection->shouldReceive('del')->once()->with(
        'collector:token:'.hash('sha256', $token->token),
    );

    app(TrackingTokenCache::class)->forget($token->token);
});

test('it surfaces a Redis failure during invalidation', function () {
    $token = cachedToken();
    $connection = Mockery::mock();
    Redis::shouldReceive('connection')->once()->andReturn($connection);
    $connection->shouldReceive('del')->once()->andThrow(new RuntimeException('Redis is down'));

    expect(fn () => app(TrackingTokenCache::class)->forget($token->token))
        ->toThrow(RuntimeException::class, 'Redis is down');
});

test('it uses the shared fallback for an unsupported token cache duration', function () {
    config()->set('services.tracking_tokens.cache_ttl', '500ms');

    $token = cachedToken();
    $connection = Mockery::mock();
    Redis::shouldReceive('connection')->once()->andReturn($connection);
    $connection->shouldReceive('setex')->once()->with(
        'collector:token:'.hash('sha256', $token->token),
        300,
        $token->site_id,
    );

    app(TrackingTokenCache::class)->publish($token);
});
