<?php

namespace App\Filament\Resources\Sites;

use App\Filament\Resources\Sites\Pages\CreateSite;
use App\Filament\Resources\Sites\Pages\EditSite;
use App\Filament\Resources\Sites\Pages\ListSites;
use App\Filament\Resources\Sites\Pages\SiteAnalytics;
use App\Filament\Resources\Sites\Pages\SiteSnippet;
use App\Filament\Resources\Sites\Pages\ViewSite;
use App\Models\Site;
use BackedEnum;
use Filament\Actions\Action;
use Filament\Actions\DeleteAction;
use Filament\Actions\EditAction;
use Filament\Actions\ViewAction;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Notifications\Notification;
use Filament\Resources\Resource;
use Filament\Schemas\Schema;
use Filament\Tables\Columns\IconColumn;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;

class SiteResource extends Resource
{
    protected static ?string $model = Site::class;

    protected static string|BackedEnum|null $navigationIcon = 'heroicon-o-globe-alt';

    protected static ?string $recordTitleAttribute = 'name';

    public static function form(Schema $schema): Schema
    {
        return $schema
            ->components([
                TextInput::make('name')
                    ->required()
                    ->maxLength(255),
                TextInput::make('domain')
                    ->label('Site domain')
                    ->placeholder('example.com')
                    ->maxLength(255),
                Toggle::make('enabled')
                    ->default(true)
                    ->required(),
            ]);
    }

    public static function table(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('name')
                    ->searchable()
                    ->sortable(),
                TextColumn::make('domain')
                    ->searchable(),
                TextColumn::make('activeToken.prefix')
                    ->label('Active token')
                    ->placeholder('No active token'),
                IconColumn::make('enabled')
                    ->boolean()
                    ->sortable(),
                TextColumn::make('created_at')
                    ->dateTime()
                    ->sortable()
                    ->toggleable(isToggledHiddenByDefault: true),
            ])
            ->defaultSort('created_at', 'desc')
            ->recordActions([
                Action::make('rotateToken')
                    ->label('Rotate token')
                    ->requiresConfirmation()
                    ->action(function (Site $record): void {
                        $token = $record->rotateToken();

                        Notification::make()
                            ->success()
                            ->title('Tracking token rotated')
                            ->body("Copy the new token now: {$token->token}")
                            ->persistent()
                            ->send();
                    }),
                EditAction::make(),
                Action::make('snippet')
                    ->icon('heroicon-o-code-bracket')
                    ->url(fn (Site $record): string => static::getUrl('snippet', ['record' => $record])),
                Action::make('analytics')
                    ->icon('heroicon-o-chart-bar')
                    ->url(fn (Site $record): string => static::getUrl('analytics', ['record' => $record])),
                ViewAction::make(),
                DeleteAction::make(),
            ]);
    }

    /**
     * @return Builder<Site>
     */
    public static function getEloquentQuery(): Builder
    {
        return parent::getEloquentQuery()
            ->where('user_id', auth()->id());
    }

    public static function getPages(): array
    {
        return [
            'index' => ListSites::route('/'),
            'create' => CreateSite::route('/create'),
            'view' => ViewSite::route('/{record}'),
            'settings' => EditSite::route('/{record}/settings'),
            'snippet' => SiteSnippet::route('/{record}/snippet'),
            'analytics' => SiteAnalytics::route('/{record}/analytics'),
        ];
    }
}
