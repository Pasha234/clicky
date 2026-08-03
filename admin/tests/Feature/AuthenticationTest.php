<?php

use App\Models\User;
use Illuminate\Foundation\Http\Middleware\ValidateCsrfToken;
use Illuminate\Support\Facades\Hash;

beforeEach(function () {
    // Authentication is the subject under test; CSRF is Laravel middleware
    // and browser form-token behavior is outside this feature test's scope.
    $this->withoutMiddleware(ValidateCsrfToken::class);
});

test('the health endpoint is available without authentication', function () {
    $this->get('/healthz')->assertNoContent();
});

test('the root page sends guests to registration', function () {
    $this->get('/')->assertRedirect('/register');
});

test('the root page sends authenticated users to the admin panel', function () {
    $user = User::factory()->create();

    $this->actingAs($user)->get('/')->assertRedirect('/admin');
});

test('a visitor can register and is authenticated', function () {
    $response = $this->post('/register', [
        'name' => 'Pavel',
        'email' => 'pavel@example.test',
        'password' => 'correct-horse-battery-staple',
        'password_confirmation' => 'correct-horse-battery-staple',
    ]);

    $user = User::query()->where('email', 'pavel@example.test')->firstOrFail();

    $response->assertRedirect('/admin');
    $this->assertAuthenticatedAs($user);
    expect(Hash::check('correct-horse-battery-staple', $user->password))->toBeTrue();
});

test('a visitor can log in with valid credentials', function () {
    $user = User::factory()->create([
        'email' => 'pavel@example.test',
        'password' => 'correct-horse-battery-staple',
    ]);

    $this->post('/login', [
        'email' => $user->email,
        'password' => 'correct-horse-battery-staple',
    ])->assertRedirect('/admin');

    $this->assertAuthenticatedAs($user);
});

test('a visitor cannot log in with an invalid password', function () {
    $user = User::factory()->create([
        'email' => 'pavel@example.test',
        'password' => 'correct-horse-battery-staple',
    ]);

    $this->from('/login')->post('/login', [
        'email' => $user->email,
        'password' => 'not-the-password',
    ])->assertRedirect('/login')
        ->assertSessionHasErrors('email');

    $this->assertGuest();
});
