<?php

namespace App\Services;

final class TokenGenerator
{
    public function generate(): string
    {
        return 'clk_'.bin2hex(random_bytes(32));
    }
}
