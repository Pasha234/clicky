<?php

namespace App\Filament\Widgets\Concerns;

use App\Services\Analytics\ClickHouseAnalytics;
use Illuminate\Http\Client\RequestException;
use RuntimeException;

trait InteractsWithBreakdownChart
{
    protected function getData(): array
    {
        $filter = $this->analyticsFilter();

        if (! $filter) {
            return ['datasets' => [], 'labels' => []];
        }

        try {
            $breakdown = app(ClickHouseAnalytics::class)->breakdown($filter, $this->dimension);
        } catch (RequestException|RuntimeException) {
            return ['datasets' => [], 'labels' => []];
        }

        return [
            'datasets' => [[
                'label' => 'Events',
                'data' => array_column($breakdown, 'events'),
            ]],
            'labels' => array_column($breakdown, 'label'),
        ];
    }

    protected function getType(): string
    {
        return 'doughnut';
    }

    public function updatedPageFilters(): void
    {
        $this->cachedData = null;
    }
}
