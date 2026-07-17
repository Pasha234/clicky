<?php

namespace App\Filament\Widgets;

use App\Filament\Widgets\Concerns\ResolvesAnalyticsFilter;
use App\Services\Analytics\ClickHouseAnalytics;
use Filament\Widgets\ChartWidget;
use Filament\Widgets\Concerns\InteractsWithPageFilters;
use Illuminate\Http\Client\RequestException;
use RuntimeException;

class TopPagesChart extends ChartWidget
{
    use InteractsWithPageFilters;
    use ResolvesAnalyticsFilter;

    protected ?string $heading = 'Top pages';

    protected function getData(): array
    {
        $filter = $this->analyticsFilter();

        if (! $filter) {
            return ['datasets' => [], 'labels' => []];
        }

        try {
            $pages = app(ClickHouseAnalytics::class)->topPages($filter);
        } catch (RequestException|RuntimeException) {
            return ['datasets' => [], 'labels' => []];
        }

        return [
            'datasets' => [[
                'label' => 'Events',
                'data' => array_column($pages, 'events'),
            ]],
            'labels' => array_column($pages, 'url'),
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
