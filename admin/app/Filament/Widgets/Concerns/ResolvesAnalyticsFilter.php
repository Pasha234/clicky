<?php

namespace App\Filament\Widgets\Concerns;

use App\Models\User;
use App\Services\Analytics\AnalyticsFilter;
use App\Services\Analytics\ClickHouseAnalytics;

trait ResolvesAnalyticsFilter
{
    protected function analyticsFilter(): ?AnalyticsFilter
    {
        $user = auth()->user();

        if (! $user instanceof User) {
            return null;
        }

        return app(ClickHouseAnalytics::class)->filterFor($user, $this->pageFilters ?? []);
    }
}
