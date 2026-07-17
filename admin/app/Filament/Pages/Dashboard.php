<?php

namespace App\Filament\Pages;

use App\Models\User;
use Filament\Forms\Components\DatePicker;
use Filament\Forms\Components\Select;
use Filament\Pages\Dashboard as BaseDashboard;
use Filament\Pages\Dashboard\Concerns\HasFiltersForm;
use Filament\Schemas\Schema;

class Dashboard extends BaseDashboard
{
    use HasFiltersForm;

    public function filtersForm(Schema $schema): Schema
    {
        return $schema->components([
            Select::make('site_id')
                ->label('Site')
                ->options(function (): array {
                    $user = auth()->user();

                    if (! $user instanceof User) {
                        return [];
                    }

                    return $user->sites()
                        ->orderBy('name')
                        ->pluck('name', 'id')
                        ->all();
                })
                ->default(function (): ?string {
                    $user = auth()->user();

                    return $user instanceof User
                        ? $user->sites()->orderBy('name')->value('id')
                        : null;
                })
                ->searchable()
                ->required(),
            DatePicker::make('from')
                ->default(now()->subDays(29)->toDateString())
                ->required(),
            DatePicker::make('to')
                ->default(now()->toDateString())
                ->required(),
        ]);
    }
}
