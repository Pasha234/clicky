<?php

use App\Services\TokenGenerator;

test('it generates a Clicky tracking token', function () {
    $token = app(TokenGenerator::class)->generate();

    expect($token)
        ->toMatch('/^clk_[a-f0-9]{64}$/')
        ->and(strlen($token))->toBe(68);
});

test('it generates different tokens on each call', function () {
    $generator = app(TokenGenerator::class);

    expect($generator->generate())->not->toBe($generator->generate());
});
