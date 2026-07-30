<?php

namespace App\Filament\Resources\Sites\Pages;

use App\Filament\Resources\Sites\SiteResource;
use Filament\Actions\Action;
use Filament\Resources\Pages\ViewRecord;

class ViewSite extends ViewRecord
{
    protected static string $resource = SiteResource::class;

    protected function getHeaderActions(): array
    {
        return [
            Action::make('settings')
                ->url(fn (): string => SiteResource::getUrl('settings', ['record' => $this->getRecord()])),
            Action::make('snippet')
                ->url(fn (): string => SiteResource::getUrl('snippet', ['record' => $this->getRecord()])),
            Action::make('analytics')
                ->url(fn (): string => SiteResource::getUrl('analytics', ['record' => $this->getRecord()])),
        ];
    }
}
