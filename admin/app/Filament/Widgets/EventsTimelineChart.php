<?php

namespace App\Filament\Widgets;

use App\Filament\Widgets\Concerns\ResolvesAnalyticsFilter;
use App\Services\Analytics\ClickHouseAnalytics;
use Filament\Widgets\ChartWidget;
use Filament\Widgets\Concerns\InteractsWithPageFilters;
use Illuminate\Http\Client\RequestException;
use RuntimeException;

class EventsTimelineChart extends ChartWidget
{
    use InteractsWithPageFilters;
    use ResolvesAnalyticsFilter;

    protected ?string $heading = 'Events over time';

    protected int|string|array $columnSpan = 'full';

    protected function getData(): array
    {
        $filter = $this->analyticsFilter();

        if (! $filter) {
            return ['datasets' => [], 'labels' => []];
        }

        try {
            $timeline = app(ClickHouseAnalytics::class)->timeline($filter);
        } catch (RequestException|RuntimeException) {
            return ['datasets' => [], 'labels' => []];
        }

        return [
            'datasets' => [[
                'label' => 'Events',
                'data' => array_column($timeline, 'events'),
            ]],
            'labels' => array_column($timeline, 'date'),
        ];
    }

    protected function getType(): string
    {
        return 'line';
    }

    protected function getOptions(): ?array
    {
        return [
            'scales' => [
                'y' => [
                    'beginAtZero' => true,
                    'min' => 0,
                    'ticks' => [
                        'precision' => 0,
                    ],
                ],
            ],
        ];
    }

    public function updatedPageFilters(): void
    {
        $this->cachedData = null;
    }
}
