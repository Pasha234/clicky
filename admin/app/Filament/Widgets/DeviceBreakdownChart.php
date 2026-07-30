<?php

namespace App\Filament\Widgets;

use App\Filament\Widgets\Concerns\InteractsWithBreakdownChart;
use App\Filament\Widgets\Concerns\ResolvesAnalyticsFilter;
use Filament\Widgets\ChartWidget;
use Filament\Widgets\Concerns\InteractsWithPageFilters;

class DeviceBreakdownChart extends ChartWidget
{
    use InteractsWithPageFilters;
    use ResolvesAnalyticsFilter;
    use InteractsWithBreakdownChart;

    protected ?string $heading = 'Devices';

    protected string $dimension = 'device';
}
