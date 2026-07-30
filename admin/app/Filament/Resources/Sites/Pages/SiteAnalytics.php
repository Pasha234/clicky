<?php

namespace App\Filament\Resources\Sites\Pages;

use App\Filament\Resources\Sites\SiteResource;
use App\Models\Site;
use App\Services\Analytics\AnalyticsFilter;
use App\Services\Analytics\ClickHouseAnalytics;
use Carbon\CarbonImmutable;
use Filament\Actions\Action;
use Filament\Resources\Pages\Concerns\InteractsWithRecord;
use Filament\Resources\Pages\Page;
use Illuminate\Http\Client\RequestException;
use RuntimeException;

class SiteAnalytics extends Page
{
    use InteractsWithRecord;

    protected static string $resource = SiteResource::class;

    protected string $view = 'filament.resources.sites.pages.site-analytics';

    public string $from;

    public string $to;

    public function mount(int|string $record): void
    {
        $this->record = $this->resolveRecord($record);
        $this->from = now()->subDays(29)->toDateString();
        $this->to = now()->toDateString();
    }

    public function getTitle(): string
    {
        return 'Site analytics';
    }

    protected function getHeaderActions(): array
    {
        return [
            Action::make('overview')
                ->url(fn (): string => SiteResource::getUrl('view', ['record' => $this->getRecord()])),
            Action::make('settings')
                ->url(fn (): string => SiteResource::getUrl('settings', ['record' => $this->getRecord()])),
            Action::make('snippet')
                ->url(fn (): string => SiteResource::getUrl('snippet', ['record' => $this->getRecord()])),
        ];
    }

    /** @return array<string, mixed> */
    protected function getViewData(): array
    {
        /** @var Site $site */
        $site = $this->getRecord();

        try {
            $filter = new AnalyticsFilter(
                $site,
                CarbonImmutable::parse($this->from)->startOfDay(),
                CarbonImmutable::parse($this->to)->startOfDay()->addDay(),
            );
            $analytics = app(ClickHouseAnalytics::class);

            return [
                'site' => $site,
                'summary' => $analytics->summary($filter),
                'timeline' => $analytics->timeline($filter),
                'topPages' => $analytics->topPages($filter),
                'referrers' => $analytics->referrers($filter),
                'unavailable' => false,
            ];
        } catch (RequestException|RuntimeException) {
            return [
                'site' => $site,
                'summary' => ['events' => 0, 'clicks' => 0, 'unique_visitors' => 0],
                'timeline' => [],
                'topPages' => [],
                'referrers' => [],
                'unavailable' => true,
            ];
        }
    }
}
