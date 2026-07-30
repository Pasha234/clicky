<!doctype html>
<html lang="{{ str_replace('_', '-', app()->getLocale()) }}">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Create your account · {{ config('app.name', 'Clicky') }}</title>
    <style>
        :root { color-scheme: light; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
        * { box-sizing: border-box; }
        body { min-height: 100vh; margin: 0; color: #e8edf7; background: #0b1020; }
        .page { min-height: 100vh; display: grid; grid-template-columns: minmax(0, 1.05fr) minmax(420px, .95fr); }
        .intro { position: relative; overflow: hidden; padding: clamp(2rem, 6vw, 6rem); display: flex; flex-direction: column; justify-content: space-between; background: linear-gradient(145deg, #151b3e, #0c1128 62%, #091f35); }
        .intro::before, .intro::after { position: absolute; content: ""; border-radius: 999px; filter: blur(4px); opacity: .6; }
        .intro::before { width: 33rem; height: 33rem; right: -12rem; top: -12rem; background: radial-gradient(circle, #5968ff 0%, transparent 67%); }
        .intro::after { width: 27rem; height: 27rem; left: -14rem; bottom: -14rem; background: radial-gradient(circle, #08b6a1 0%, transparent 67%); }
        .brand, .intro-copy, .features { position: relative; z-index: 1; }
        .brand { color: #fff; font-size: 1.25rem; font-weight: 750; letter-spacing: -.04em; text-decoration: none; }
        .brand-mark { display: inline-grid; place-items: center; width: 1.8rem; height: 1.8rem; margin-right: .55rem; border-radius: .55rem; color: #101735; background: #aab4ff; font-size: 1rem; }
        .eyebrow { margin: 0 0 1rem; color: #aab4ff; font-size: .75rem; font-weight: 700; letter-spacing: .13em; text-transform: uppercase; }
        h1 { max-width: 13ch; margin: 0; color: #fff; font-size: clamp(2.6rem, 5vw, 4.75rem); line-height: 1.02; letter-spacing: -.065em; }
        .intro-copy p:last-child { max-width: 34rem; margin: 1.5rem 0 0; color: #b7c0dd; font-size: 1.05rem; line-height: 1.65; }
        .features { display: grid; gap: .85rem; max-width: 30rem; }
        .feature { display: flex; gap: .75rem; align-items: center; color: #d5dcf5; font-size: .92rem; }
        .check { display: grid; place-items: center; flex: 0 0 auto; width: 1.35rem; height: 1.35rem; border-radius: 50%; color: #9dfff2; background: #ffffff14; font-size: .8rem; }
        .form-panel { display: grid; place-items: center; padding: 2rem; background: #f7f8fc; color: #17203b; }
        .card { width: min(100%, 27rem); }
        .card-header { margin-bottom: 2rem; }
        h2 { margin: 0; font-size: 1.8rem; letter-spacing: -.045em; }
        .card-header p { margin: .55rem 0 0; color: #68728a; line-height: 1.5; }
        .field { margin-top: 1.2rem; }
        label { display: block; margin-bottom: .45rem; color: #34405b; font-size: .875rem; font-weight: 650; }
        input { width: 100%; border: 1px solid #d9deea; border-radius: .7rem; outline: none; padding: .78rem .9rem; color: #17203b; background: #fff; font: inherit; transition: border-color .15s, box-shadow .15s; }
        input::placeholder { color: #a0a8b9; }
        input:focus { border-color: #5363e8; box-shadow: 0 0 0 4px #5363e820; }
        input[aria-invalid="true"] { border-color: #dd5562; }
        .error { margin: .4rem 0 0; color: #bb3040; font-size: .8rem; }
        .button { width: 100%; margin-top: 1.65rem; border: 0; border-radius: .7rem; padding: .85rem 1rem; cursor: pointer; color: #fff; background: #4858d8; font: inherit; font-weight: 700; transition: background .15s, transform .15s; }
        .button:hover { background: #3949c8; }
        .button:active { transform: translateY(1px); }
        .login { margin: 1.5rem 0 0; color: #68728a; text-align: center; font-size: .9rem; }
        .login a { color: #4353d0; font-weight: 700; text-decoration: none; }
        .login a:hover { text-decoration: underline; }
        .terms { margin: 1.5rem auto 0; max-width: 21rem; color: #8a93a6; text-align: center; font-size: .75rem; line-height: 1.5; }
        @media (max-width: 800px) { .page { display: block; } .intro { min-height: auto; padding: 2rem 1.5rem; } .intro-copy { margin: 4rem 0; } .features { display: none; } .form-panel { min-height: 68vh; padding: 2.5rem 1.5rem; } }
    </style>
</head>
<body>
<main class="page">
    <section class="intro" aria-label="About {{ config('app.name', 'Clicky') }}">
        <a class="brand" href="/"><span class="brand-mark">↗</span>{{ config('app.name', 'Clicky') }}</a>

        <div class="intro-copy">
            <p class="eyebrow">Privacy-friendly analytics</p>
            <h1>Know what your visitors do.</h1>
            <p>Set up your site in a few minutes, collect the events that matter, and see the signal without the noise.</p>
        </div>

        <div class="features" aria-label="Benefits">
            <div class="feature"><span class="check">✓</span> Simple event collection for your websites</div>
            <div class="feature"><span class="check">✓</span> Clear dashboards for pages and referrers</div>
            <div class="feature"><span class="check">✓</span> Your data stays yours</div>
        </div>
    </section>

    <section class="form-panel">
        <div class="card">
            <header class="card-header">
                <h2>Create your account</h2>
                <p>Start tracking your first site today.</p>
            </header>

            <form method="post" action="{{ route('register') }}">
                @csrf

                <div class="field">
                    <label for="name">Name</label>
                    <input id="name" name="name" type="text" value="{{ old('name') }}" autocomplete="name" placeholder="Pavel Petrov" required autofocus aria-invalid="{{ $errors->has('name') ? 'true' : 'false' }}" aria-describedby="name-error">
                    @error('name') <p id="name-error" class="error">{{ $message }}</p> @enderror
                </div>

                <div class="field">
                    <label for="email">Email address</label>
                    <input id="email" name="email" type="email" value="{{ old('email') }}" autocomplete="email" placeholder="you@example.com" required aria-invalid="{{ $errors->has('email') ? 'true' : 'false' }}" aria-describedby="email-error">
                    @error('email') <p id="email-error" class="error">{{ $message }}</p> @enderror
                </div>

                <div class="field">
                    <label for="password">Password</label>
                    <input id="password" name="password" type="password" autocomplete="new-password" placeholder="At least 8 characters" required aria-invalid="{{ $errors->has('password') ? 'true' : 'false' }}" aria-describedby="password-error">
                    @error('password') <p id="password-error" class="error">{{ $message }}</p> @enderror
                </div>

                <div class="field">
                    <label for="password_confirmation">Confirm password</label>
                    <input id="password_confirmation" name="password_confirmation" type="password" autocomplete="new-password" placeholder="Repeat your password" required>
                </div>

                <button class="button" type="submit">Create account</button>
            </form>

            <p class="login">Already have an account? <a href="{{ route('login') }}">Log in</a></p>
            <p class="terms">By creating an account, you agree to use this project responsibly.</p>
        </div>
    </section>
</main>
</body>
</html>
