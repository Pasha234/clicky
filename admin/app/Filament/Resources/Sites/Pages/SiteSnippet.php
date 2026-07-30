<?php

namespace App\Filament\Resources\Sites\Pages;

use App\Filament\Resources\Sites\SiteResource;
use App\Models\Site;
use Filament\Actions\Action;
use Filament\Resources\Pages\Concerns\InteractsWithRecord;
use Filament\Resources\Pages\Page;

class SiteSnippet extends Page
{
    use InteractsWithRecord;

    protected static string $resource = SiteResource::class;

    protected string $view = 'filament.resources.sites.pages.site-snippet';

    public function mount(int|string $record): void
    {
        $this->record = $this->resolveRecord($record);
    }

    public function getTitle(): string
    {
        return 'Tracking snippet';
    }

    protected function getHeaderActions(): array
    {
        return [
            Action::make('overview')
                ->url(fn (): string => SiteResource::getUrl('view', ['record' => $this->getRecord()])),
            Action::make('settings')
                ->url(fn (): string => SiteResource::getUrl('settings', ['record' => $this->getRecord()])),
            Action::make('analytics')
                ->url(fn (): string => SiteResource::getUrl('analytics', ['record' => $this->getRecord()])),
        ];
    }

    /** @return array{site: Site, token: ?string, snippet: string} */
    protected function getViewData(): array
    {
        /** @var Site $site */
        $site = $this->getRecord();

        return [
            'site' => $site,
            'token' => $site->activeToken?->token,
            'snippet' => $site->trackingSnippet(),
        ];
    }
}
