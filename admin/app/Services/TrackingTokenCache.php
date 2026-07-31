<?php

namespace App\Services;

use App\Models\ApiToken;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Redis;
use Throwable;

final class TrackingTokenCache
{
    public function publish(ApiToken $token): void
    {
        if ($token->revoked_at !== null) {
            $this->forget($token->token);

            return;
        }

        try {
            Redis::connection()->setex(
                $this->key($token->token),
                $this->ttlSeconds(),
                (string) $token->site_id,
            );
        } catch (Throwable $exception) {
            Log::warning('Could not publish tracking token to Redis.', [
                'token_prefix' => $token->prefix,
                'exception' => $exception->getMessage(),
            ]);
        }
    }

    public function forget(string $token): void
    {
        // Callers use this before a PostgreSQL revocation/disable/delete.
        // Let Redis failures abort that state change instead of leaving a
        // positive cache hit that can outlive the database revocation.
        Redis::connection()->del($this->key($token));
    }

    private function key(string $token): string
    {
        return 'collector:token:'.hash('sha256', $token);
    }

    private function ttlSeconds(): int
    {
        $value = (string) config('services.tracking_tokens.cache_ttl');

        if (! preg_match('/^([1-9]\d*)(s|m|h)$/', $value, $matches)) {
            return 300;
        }

        $multiplier = match ($matches[2]) {
            's' => 1,
            'm' => 60,
            'h' => 3600,
        };

        return max(1, (int) $matches[1] * $multiplier);
    }
}
