<?php

namespace App\Console\Commands;

use App\Models\Site;
use Illuminate\Console\Command;

final class SyncTrackingTokenCache extends Command
{
    protected $signature = 'tracking-tokens:sync-cache';

    protected $description = 'Publish active tracking tokens to Redis and remove disabled-site tokens.';

    public function handle(): int
    {
        $sites = Site::query()->get();

        foreach ($sites as $site) {
            $site->syncTrackingTokenCache();
        }

        $this->components->info("Synced {$sites->count()} site token cache entries.");

        return self::SUCCESS;
    }
}
