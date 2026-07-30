<x-filament-panels::page>
    <style>
        .site-analytics { display: grid; gap: 1.5rem; }
        .site-analytics__summary { display: grid; gap: 1rem; }
        .site-analytics__summary .fi-section { height: 100%; }
        .site-analytics__metric { display: flex; align-items: flex-start; justify-content: space-between; gap: .75rem; }
        .site-analytics__metric-value { margin-top: .75rem; }
        .site-analytics__lists { display: grid; gap: 1.5rem; }
        .site-analytics__timeline { margin-top: 0; }
        .site-analytics__filter { margin-top: 1rem; }
        .site-analytics__empty { margin-top: .75rem; }
        .site-analytics__chart { margin-top: 1rem; overflow-x: auto; }
        .site-analytics__chart svg { display: block; min-width: 38rem; width: 100%; }
        .site-analytics__chart-grid { stroke: #dbe1ea; stroke-width: 1; }
        .dark .site-analytics__chart-grid { stroke: rgba(255, 255, 255, .14); }
        .site-analytics__chart-area { fill: rgba(99, 102, 241, .12); }
        .site-analytics__chart-line { fill: none; stroke: #6366f1; stroke-linecap: round; stroke-linejoin: round; stroke-width: 3; }
        .site-analytics__chart-point { fill: #fff; stroke: #6366f1; stroke-width: 2.5; }
        .dark .site-analytics__chart-point { fill: #111827; }
        .site-analytics__chart-label { fill: #6b7280; font-size: 12px; }
        .dark .site-analytics__chart-label { fill: #9ca3af; }
        @media (min-width: 768px) { .site-analytics__summary { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
        @media (min-width: 1024px) { .site-analytics__lists { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
    </style>

    <div class="site-analytics space-y-6">
        <x-filament::section>
            <x-slot name="heading">{{ $site->name }}</x-slot>
            <x-slot name="description">Choose the period you want to analyse.</x-slot>

            <div class="site-analytics__filter rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-white/10 dark:bg-white/5">
                <div class="grid gap-4 md:grid-cols-[1fr_1fr_auto] md:items-end">
                    <label class="grid gap-1.5 text-sm font-medium text-gray-700 dark:text-gray-200">
                        <span>From</span>
                        <input type="date" wire:model.live="from" class="w-full rounded-lg border-gray-300 bg-white px-3 py-2 shadow-sm transition focus:border-primary-500 focus:ring-primary-500 dark:border-gray-700 dark:bg-gray-900">
                    </label>
                    <label class="grid gap-1.5 text-sm font-medium text-gray-700 dark:text-gray-200">
                        <span>To</span>
                        <input type="date" wire:model.live="to" class="w-full rounded-lg border-gray-300 bg-white px-3 py-2 shadow-sm transition focus:border-primary-500 focus:ring-primary-500 dark:border-gray-700 dark:bg-gray-900">
                    </label>
                    <div class="pb-2 text-sm text-gray-500 dark:text-gray-400" wire:loading wire:target="from,to">
                        Updating…
                    </div>
                </div>
            </div>
        </x-filament::section>

        @if ($unavailable)
            <x-filament::section>
                <div class="rounded-lg border border-danger-200 bg-danger-50 p-4 text-sm text-danger-700 dark:border-danger-500/30 dark:bg-danger-500/10 dark:text-danger-300">
                    Analytics are temporarily unavailable. Please try again shortly.
                </div>
            </x-filament::section>
        @else
            <div class="site-analytics__summary grid gap-4 md:grid-cols-3">
                <x-filament::section class="h-full">
                    <div class="site-analytics__metric flex items-start justify-between gap-3">
                        <div>
                            <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Events</p>
                            <p class="site-analytics__metric-value mt-3 text-3xl font-bold tracking-tight text-gray-950 dark:text-white">{{ number_format($summary['events']) }}</p>
                        </div>
                        <span class="rounded-lg bg-primary-50 px-2.5 py-1 text-xs font-semibold text-primary-700 dark:bg-primary-500/10 dark:text-primary-300">Total</span>
                    </div>
                </x-filament::section>

                <x-filament::section class="h-full">
                    <div class="site-analytics__metric flex items-start justify-between gap-3">
                        <div>
                            <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Clicks</p>
                            <p class="site-analytics__metric-value mt-3 text-3xl font-bold tracking-tight text-gray-950 dark:text-white">{{ number_format($summary['clicks']) }}</p>
                        </div>
                        <span class="rounded-lg bg-success-50 px-2.5 py-1 text-xs font-semibold text-success-700 dark:bg-success-500/10 dark:text-success-300">Interactions</span>
                    </div>
                </x-filament::section>

                <x-filament::section class="h-full">
                    <div class="site-analytics__metric flex items-start justify-between gap-3">
                        <div>
                            <p class="text-sm font-medium text-gray-500 dark:text-gray-400">Unique visitors</p>
                            <p class="site-analytics__metric-value mt-3 text-3xl font-bold tracking-tight text-gray-950 dark:text-white">{{ number_format($summary['unique_visitors']) }}</p>
                        </div>
                        <span class="rounded-lg bg-warning-50 px-2.5 py-1 text-xs font-semibold text-warning-700 dark:bg-warning-500/10 dark:text-warning-300">Audience</span>
                    </div>
                </x-filament::section>
            </div>

            <div class="site-analytics__lists grid gap-6 lg:grid-cols-2">
                <x-filament::section>
                    <x-slot name="heading">Top pages</x-slot>
                    <x-slot name="description">Most active URLs in the selected period.</x-slot>

                    <div class="divide-y divide-gray-100 dark:divide-white/10">
                        @forelse ($topPages as $page)
                            <div class="flex items-center justify-between gap-4 py-3 first:pt-0 last:pb-0">
                                <span class="min-w-0 truncate text-sm font-medium text-gray-700 dark:text-gray-200" title="{{ $page['url'] }}">{{ $page['url'] }}</span>
                                <span class="shrink-0 rounded-md bg-gray-100 px-2 py-1 text-xs font-semibold tabular-nums text-gray-700 dark:bg-white/10 dark:text-gray-200">{{ number_format($page['events']) }}</span>
                            </div>
                        @empty
                            <div class="site-analytics__empty rounded-lg bg-gray-50 px-4 py-8 text-center text-sm text-gray-500 dark:bg-white/5 dark:text-gray-400">No events in this range.</div>
                        @endforelse
                    </div>
                </x-filament::section>

                <x-filament::section>
                    <x-slot name="heading">Referrers</x-slot>
                    <x-slot name="description">Where your visitors came from.</x-slot>

                    <div class="divide-y divide-gray-100 dark:divide-white/10">
                        @forelse ($referrers as $referrer)
                            <div class="flex items-center justify-between gap-4 py-3 first:pt-0 last:pb-0">
                                <span class="min-w-0 truncate text-sm font-medium text-gray-700 dark:text-gray-200" title="{{ $referrer['referrer'] }}">{{ $referrer['referrer'] }}</span>
                                <span class="shrink-0 rounded-md bg-gray-100 px-2 py-1 text-xs font-semibold tabular-nums text-gray-700 dark:bg-white/10 dark:text-gray-200">{{ number_format($referrer['events']) }}</span>
                            </div>
                        @empty
                            <div class="site-analytics__empty rounded-lg bg-gray-50 px-4 py-8 text-center text-sm text-gray-500 dark:bg-white/5 dark:text-gray-400">No referrers in this range.</div>
                        @endforelse
                    </div>
                </x-filament::section>
            </div>

            <x-filament::section class="site-analytics__timeline">
                <x-slot name="heading">Daily events</x-slot>
                <x-slot name="description">Event volume for each day in the selected period.</x-slot>

                @if (count($timeline))
                    @php
                        $chartWidth = 720;
                        $chartHeight = 260;
                        $chartLeft = 48;
                        $chartRight = 20;
                        $chartTop = 20;
                        $chartBottom = 42;
                        $plotWidth = $chartWidth - $chartLeft - $chartRight;
                        $plotHeight = $chartHeight - $chartTop - $chartBottom;
                        $maxTimelineEvents = max(1, max(array_column($timeline, 'events')));
                        $chartPoints = [];

                        foreach ($timeline as $index => $day) {
                            $x = $chartLeft + (count($timeline) === 1 ? $plotWidth / 2 : $index * $plotWidth / (count($timeline) - 1));
                            $y = $chartTop + $plotHeight - ((int) $day['events'] / $maxTimelineEvents * $plotHeight);
                            $chartPoints[] = ['x' => $x, 'y' => $y, 'day' => $day];
                        }

                        $linePoints = collect($chartPoints)->map(fn (array $point) => $point['x'].','.$point['y'])->implode(' ');
                        $areaPoints = $chartLeft.','.($chartTop + $plotHeight).' '.$linePoints.' '.($chartLeft + $plotWidth).','.($chartTop + $plotHeight);
                        $labelEvery = max(1, (int) ceil(count($timeline) / 6));
                    @endphp

                    <div class="site-analytics__chart rounded-xl border border-gray-200 bg-white p-3 dark:border-white/10 dark:bg-white/5">
                        <svg viewBox="0 0 {{ $chartWidth }} {{ $chartHeight }}" role="img" aria-label="Daily events line chart">
                            @foreach ([0, 0.5, 1] as $step)
                                @php($y = $chartTop + $plotHeight - ($step * $plotHeight))
                                <line class="site-analytics__chart-grid" x1="{{ $chartLeft }}" x2="{{ $chartWidth - $chartRight }}" y1="{{ $y }}" y2="{{ $y }}" />
                                <text class="site-analytics__chart-label" x="{{ $chartLeft - 10 }}" y="{{ $y + 4 }}" text-anchor="end">{{ number_format((int) round($maxTimelineEvents * $step)) }}</text>
                            @endforeach

                            <polygon class="site-analytics__chart-area" points="{{ $areaPoints }}" />
                            <polyline class="site-analytics__chart-line" points="{{ $linePoints }}" />

                            @foreach ($chartPoints as $index => $point)
                                <circle class="site-analytics__chart-point" cx="{{ $point['x'] }}" cy="{{ $point['y'] }}" r="4">
                                    <title>{{ \Carbon\Carbon::parse($point['day']['date'])->format('M j, Y') }}: {{ number_format($point['day']['events']) }} events</title>
                                </circle>

                                @if ($index % $labelEvery === 0 || $index === count($chartPoints) - 1)
                                    <text class="site-analytics__chart-label" x="{{ $point['x'] }}" y="{{ $chartHeight - 15 }}" text-anchor="middle">{{ \Carbon\Carbon::parse($point['day']['date'])->format('M j') }}</text>
                                @endif
                            @endforeach
                        </svg>
                    </div>
                @else
                    <p class="site-analytics__empty rounded-lg bg-gray-50 px-4 py-8 text-center text-sm text-gray-500 dark:bg-white/5 dark:text-gray-400">No events in this range.</p>
                @endif
            </x-filament::section>
        @endif
    </div>
</x-filament-panels::page>
