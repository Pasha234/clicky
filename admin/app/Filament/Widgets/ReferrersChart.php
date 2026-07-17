<?php

namespace App\Filament\Widgets;

use App\Filament\Widgets\Concerns\ResolvesAnalyticsFilter;
use App\Services\Analytics\ClickHouseAnalytics;
use Filament\Widgets\ChartWidget;
use Filament\Widgets\Concerns\InteractsWithPageFilters;
use Illuminate\Http\Client\RequestException;
use RuntimeException;

class ReferrersChart extends ChartWidget
{
    use InteractsWithPageFilters;
    use ResolvesAnalyticsFilter;

    protected ?string $heading = 'Top referrers';

    protected function getData(): array
    {
        $filter = $this->analyticsFilter();

        if (! $filter) {
            return ['datasets' => [], 'labels' => []];
        }

        try {
            $referrers = app(ClickHouseAnalytics::class)->referrers($filter);
        } catch (RequestException|RuntimeException) {
            return ['datasets' => [], 'labels' => []];
        }

        return [
            'datasets' => [[
                'label' => 'Events',
                'data' => array_column($referrers, 'events'),
            ]],
            'labels' => array_column($referrers, 'referrer'),
        ];
    }

    protected function getType(): string
    {
        return 'bar';
    }

    public function updatedPageFilters(): void
    {
        $this->cachedData = null;
    }
}
