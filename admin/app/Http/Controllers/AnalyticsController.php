<?php

namespace App\Http\Controllers;

use App\Services\Analytics\ClickHouseAnalytics;
use Illuminate\Http\Client\RequestException;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use RuntimeException;

class AnalyticsController extends Controller
{
    public function __construct(private readonly ClickHouseAnalytics $analytics) {}

    public function summary(Request $request): JsonResponse
    {
        return $this->respond($request, fn ($filter): array => $this->analytics->summary($filter));
    }

    public function timeline(Request $request): JsonResponse
    {
        return $this->respond($request, fn ($filter): array => ['data' => $this->analytics->timeline($filter)]);
    }

    public function topPages(Request $request): JsonResponse
    {
        return $this->respond($request, fn ($filter): array => ['data' => $this->analytics->topPages($filter)]);
    }

    public function referrers(Request $request): JsonResponse
    {
        return $this->respond($request, fn ($filter): array => ['data' => $this->analytics->referrers($filter)]);
    }

    private function respond(Request $request, callable $callback): JsonResponse
    {
        $input = $request->validate([
            'site_id' => ['required', 'uuid'],
            'from' => ['nullable', 'date_format:Y-m-d'],
            'to' => ['nullable', 'date_format:Y-m-d', 'after_or_equal:from'],
        ]);

        $filter = $this->analytics->filterFor($request->user(), $input);
        abort_unless($filter, 404);

        try {
            return response()->json($callback($filter));
        } catch (RequestException|RuntimeException) {
            return response()->json(['message' => 'Analytics are temporarily unavailable.'], 503);
        }
    }
}
