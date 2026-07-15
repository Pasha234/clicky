<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Concerns\HasUuids;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\Relations\HasOne;

class Site extends Model
{
    /** @use HasFactory<\Database\Factories\SiteFactory> */
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
        $this->apiTokens()
            ->whereNull('revoked_at')
            ->update(['revoked_at' => now()]);

        $token = 'clk_'.bin2hex(random_bytes(32));

        $apiToken = $this->apiTokens()->create([
            'token' => $token,
            'prefix' => substr($token, 0, 12),
        ]);

        $this->unsetRelation('activeToken');

        return $apiToken;
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
