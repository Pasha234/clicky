<?php

use App\Http\Controllers\AnalyticsController;
use App\Http\Controllers\AuthController;
use Illuminate\Support\Facades\Route;

// Kubernetes liveness probe. Keep this independent of sessions and external
// services so a temporary dependency outage does not restart PHP itself.
Route::get('/healthz', fn () => response()->noContent());

Route::get('/', function () {
    return redirect(auth()->check() ? '/admin' : '/register');
});

Route::middleware('guest')->group(function (): void {
    Route::get('/register', [AuthController::class, 'create'])->name('register');
    Route::post('/register', [AuthController::class, 'store']);
    Route::get('/login', [AuthController::class, 'login'])->name('login');
    Route::post('/login', [AuthController::class, 'authenticate']);
});

Route::post('/logout', [AuthController::class, 'destroy'])->middleware('auth')->name('logout');

Route::middleware('auth')->prefix('api/sites/{site}/analytics')->group(function (): void {
    Route::get('summary', [AnalyticsController::class, 'summary']);
    Route::get('timeline', [AnalyticsController::class, 'timeline']);
    Route::get('top-pages', [AnalyticsController::class, 'topPages']);
    Route::get('referrers', [AnalyticsController::class, 'referrers']);
});
