<?php

use App\Http\Controllers\AnalyticsController;
use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    return view('welcome');
});

Route::middleware('auth')->prefix('analytics')->group(function (): void {
    Route::get('summary', [AnalyticsController::class, 'summary']);
    Route::get('timeline', [AnalyticsController::class, 'timeline']);
    Route::get('top-pages', [AnalyticsController::class, 'topPages']);
    Route::get('referrers', [AnalyticsController::class, 'referrers']);
});
