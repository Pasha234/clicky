<?php

namespace App\Services\Analytics;

use App\Models\Site;
use Carbon\CarbonImmutable;

final readonly class AnalyticsFilter
{
    public function __construct(
        public Site $site,
        public CarbonImmutable $from,
        public CarbonImmutable $toExclusive,
    ) {}
}
