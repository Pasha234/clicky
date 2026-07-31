<?php

namespace App\Models;

use App\Services\TokenGenerator;
use App\Services\TrackingTokenCache;
use Database\Factories\SiteFactory;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Concerns\HasUuids;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\Relations\HasOne;
use Illuminate\Support\Facades\DB;

class Site extends Model
{
    /** @use HasFactory<SiteFactory> */
    use HasFactory, HasUuids;

    protected $fillable = [
        'user_id',
        'name',
        'domain',
        'enabled',
    ];

    protected function casts(): array
    {
        return [
            'enabled' => 'boolean',
        ];
    }

    /**
     * PDO's emulated PostgreSQL prepares bind PHP booleans as integers.
     * Use PostgreSQL's accepted boolean literal while keeping the model cast.
     */
    protected function setEnabledAttribute(mixed $value): void
    {
        $this->attributes['enabled'] = filter_var($value, FILTER_VALIDATE_BOOL)
            ? 'true'
            : 'false';
    }

    protected static function booted(): void
    {
        static::created(function (self $site): void {
            $site->rotateToken();
        });

        static::updating(function (self $site): void {
            if ($site->isDirty('enabled') && ! $site->enabled) {
                $site->invalidateTrackingTokenCache();
            }
        });

        static::updated(function (self $site): void {
            if ($site->wasChanged('enabled') && $site->enabled) {
                $site->syncTrackingTokenCache();
            }
        });

        static::deleting(function (self $site): void {
            $site->invalidateTrackingTokenCache();
        });
    }

    /**
     * @return BelongsTo<User, $this>
     */
    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    /**
     * @return HasMany<ApiToken, $this>
     */
    public function apiTokens(): HasMany
    {
        return $this->hasMany(ApiToken::class);
    }

    /**
     * @return HasOne<ApiToken, $this>
     */
    public function activeToken(): HasOne
    {
        return $this->hasOne(ApiToken::class)
            ->whereNull('revoked_at')
            ->latest('created_at');
    }

    public function rotateToken(): ApiToken
    {
        $oldTokens = $this->apiTokens()->active()->pluck('token');
        $oldTokens->each(fn (string $token): mixed => app(TrackingTokenCache::class)->forget($token));

        $token = app(TokenGenerator::class)->generate();
        $apiToken = DB::transaction(function () use ($token): ApiToken {
            $this->apiTokens()
                ->whereNull('revoked_at')
                ->update(['revoked_at' => now()]);

            return $this->apiTokens()->create([
                'token' => $token,
                'prefix' => substr($token, 0, 12),
            ]);
        });

        $this->unsetRelation('activeToken');

        if ($this->enabled) {
            app(TrackingTokenCache::class)->publish($apiToken);
        }

        return $apiToken;
    }

    public function syncTrackingTokenCache(): void
    {
        if (! $this->enabled) {
            return;
        }

        foreach ($this->apiTokens()->active()->get() as $token) {
            app(TrackingTokenCache::class)->publish($token);
        }
    }

    public function invalidateTrackingTokenCache(): void
    {
        $this->apiTokens()->pluck('token')->each(
            fn (string $token): mixed => app(TrackingTokenCache::class)->forget($token),
        );
    }

    public function trackingSnippet(): string
    {
        $collectorUrl = json_encode(
            rtrim((string) config('services.collector.url'), '/').'/collect',
            JSON_UNESCAPED_SLASHES,
        );
        $token = json_encode($this->activeToken?->token, JSON_UNESCAPED_SLASHES);

        return <<<HTML
<script>
(() => {
  const token = {$token};

  document.addEventListener('click', (event) => {
    fetch({$collectorUrl}, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      keepalive: true,
      body: JSON.stringify({
        token,
        event: 'click',
        url: location.href,
        referrer: document.referrer,
        x: event.clientX,
        y: event.clientY,
        timestamp: new Date().toISOString(),
      }),
    });
  });
})();
</script>
HTML;
    }

    /**
     * @param  Builder<self>  $query
     * @return Builder<self>
     */
    public function scopeEnabled(Builder $query): Builder
    {
        return $query->where('enabled', 'true');
    }
}
