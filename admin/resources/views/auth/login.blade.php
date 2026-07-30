<!doctype html>
<html lang="{{ str_replace('_', '-', app()->getLocale()) }}">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Log in · {{ config('app.name', 'Clicky') }}</title>
    <style>
        :root { color-scheme: light; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
        * { box-sizing: border-box; }
        body { min-height: 100vh; margin: 0; color: #e8edf7; background: #0b1020; }
        .page { min-height: 100vh; display: grid; grid-template-columns: minmax(0, 1.05fr) minmax(420px, .95fr); }
        .intro { position: relative; overflow: hidden; padding: clamp(2rem, 6vw, 6rem); display: flex; flex-direction: column; justify-content: space-between; background: linear-gradient(145deg, #151b3e, #0c1128 62%, #091f35); }
        .intro::before, .intro::after { position: absolute; content: ""; border-radius: 999px; filter: blur(4px); opacity: .6; }
        .intro::before { width: 33rem; height: 33rem; right: -12rem; top: -12rem; background: radial-gradient(circle, #5968ff 0%, transparent 67%); }
        .intro::after { width: 27rem; height: 27rem; left: -14rem; bottom: -14rem; background: radial-gradient(circle, #08b6a1 0%, transparent 67%); }
        .brand, .intro-copy, .quote { position: relative; z-index: 1; }
        .brand { color: #fff; font-size: 1.25rem; font-weight: 750; letter-spacing: -.04em; text-decoration: none; }
        .brand-mark { display: inline-grid; place-items: center; width: 1.8rem; height: 1.8rem; margin-right: .55rem; border-radius: .55rem; color: #101735; background: #aab4ff; font-size: 1rem; }
        .eyebrow { margin: 0 0 1rem; color: #aab4ff; font-size: .75rem; font-weight: 700; letter-spacing: .13em; text-transform: uppercase; }
        h1 { max-width: 12ch; margin: 0; color: #fff; font-size: clamp(2.6rem, 5vw, 4.75rem); line-height: 1.02; letter-spacing: -.065em; }
        .intro-copy p:last-child { max-width: 34rem; margin: 1.5rem 0 0; color: #b7c0dd; font-size: 1.05rem; line-height: 1.65; }
        .quote { max-width: 28rem; padding-left: 1rem; border-left: 2px solid #7e8cff; color: #d5dcf5; font-size: .95rem; font-style: italic; line-height: 1.55; }
        .form-panel { display: grid; place-items: center; padding: 2rem; background: #f7f8fc; color: #17203b; }
        .card { width: min(100%, 27rem); }
        .card-header { margin-bottom: 2rem; }
        h2 { margin: 0; font-size: 1.8rem; letter-spacing: -.045em; }
        .card-header p { margin: .55rem 0 0; color: #68728a; line-height: 1.5; }
        .field { margin-top: 1.2rem; }
        label { display: block; margin-bottom: .45rem; color: #34405b; font-size: .875rem; font-weight: 650; }
        input[type="email"], input[type="password"] { width: 100%; border: 1px solid #d9deea; border-radius: .7rem; outline: none; padding: .78rem .9rem; color: #17203b; background: #fff; font: inherit; transition: border-color .15s, box-shadow .15s; }
        input::placeholder { color: #a0a8b9; }
        input:focus { border-color: #5363e8; box-shadow: 0 0 0 4px #5363e820; }
        input[aria-invalid="true"] { border-color: #dd5562; }
        .error { margin: .4rem 0 0; color: #bb3040; font-size: .8rem; }
        .remember { display: flex; gap: .55rem; align-items: center; margin-top: 1.15rem; color: #556078; font-size: .875rem; }
        .remember input { width: 1rem; height: 1rem; accent-color: #4858d8; }
        .button { width: 100%; margin-top: 1.65rem; border: 0; border-radius: .7rem; padding: .85rem 1rem; cursor: pointer; color: #fff; background: #4858d8; font: inherit; font-weight: 700; transition: background .15s, transform .15s; }
        .button:hover { background: #3949c8; }
        .button:active { transform: translateY(1px); }
        .signup { margin: 1.5rem 0 0; color: #68728a; text-align: center; font-size: .9rem; }
        .signup a { color: #4353d0; font-weight: 700; text-decoration: none; }
        .signup a:hover { text-decoration: underline; }
        @media (max-width: 800px) { .page { display: block; } .intro { min-height: auto; padding: 2rem 1.5rem; } .intro-copy { margin: 4rem 0; } .quote { display: none; } .form-panel { min-height: 62vh; padding: 2.5rem 1.5rem; } }
    </style>
</head>
<body>
<main class="page">
    <section class="intro" aria-label="About {{ config('app.name', 'Clicky') }}">
        <a class="brand" href="/"><span class="brand-mark">↗</span>{{ config('app.name', 'Clicky') }}</a>

        <div class="intro-copy">
            <p class="eyebrow">Welcome back</p>
            <h1>Make every visit count.</h1>
            <p>Log in to see what is happening on your sites and turn events into useful decisions.</p>
        </div>

        <p class="quote">“Simple, focused analytics for the moments that matter.”</p>
    </section>

    <section class="form-panel">
        <div class="card">
            <header class="card-header">
                <h2>Log in to your account</h2>
                <p>Enter your details to continue to your dashboard.</p>
            </header>

            <form method="post" action="{{ route('login') }}">
                @csrf

                <div class="field">
                    <label for="email">Email address</label>
                    <input id="email" name="email" type="email" value="{{ old('email') }}" autocomplete="email" placeholder="you@example.com" required autofocus aria-invalid="{{ $errors->has('email') ? 'true' : 'false' }}" aria-describedby="email-error">
                    @error('email') <p id="email-error" class="error">{{ $message }}</p> @enderror
                </div>

                <div class="field">
                    <label for="password">Password</label>
                    <input id="password" name="password" type="password" autocomplete="current-password" placeholder="Your password" required>
                </div>

                <label class="remember" for="remember">
                    <input id="remember" type="checkbox" name="remember" value="1" @checked(old('remember'))>
                    Keep me signed in
                </label>

                <button class="button" type="submit">Log in</button>
            </form>

            <p class="signup">New to {{ config('app.name', 'Clicky') }}? <a href="{{ route('register') }}">Create an account</a></p>
        </div>
    </section>
</main>
</body>
</html>
