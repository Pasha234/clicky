<?php

namespace App\Filament\Widgets;

use App\Filament\Widgets\Concerns\ResolvesAnalyticsFilter;
use App\Services\Analytics\ClickHouseAnalytics;
use Filament\Widgets\Concerns\InteractsWithPageFilters;
use Filament\Widgets\StatsOverviewWidget;
use Filament\Widgets\StatsOverviewWidget\Stat;
use Illuminate\Http\Client\RequestException;
use RuntimeException;

class AnalyticsOverview extends StatsOverviewWidget
{
    use InteractsWithPageFilters;
    use ResolvesAnalyticsFilter;

    protected int|string|array $columnSpan = 'full';

    protected function getStats(): array
    {
        $filter = $this->analyticsFilter();

        if (! $filter) {
            return [
                Stat::make('Events', '—'),
                Stat::make('Clicks', '—'),
                Stat::make('Unique visitors', '—'),
            ];
        }

        try {
            $summary = app(ClickHouseAnalytics::class)->summary($filter);
        } catch (RequestException|RuntimeException $e) {
            dd($e);
            return [
                Stat::make('Analytics unavailable', '—')
                    ->description('ClickHouse could not be reached.'),
            ];
        }

        return [
            Stat::make('Events', number_format($summary['events'])),
            Stat::make('Clicks', number_format($summary['clicks'])),
            Stat::make('Unique visitors', number_format($summary['unique_visitors'])),
        ];
    }

    public function updatedPageFilters(): void
    {
        $this->cachedStats = null;
    }
}
