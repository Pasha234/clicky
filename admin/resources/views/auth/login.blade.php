<!doctype html>
<html lang="en">
<body>
<h1>Log in</h1>
<form method="post" action="/login">
    @csrf
    <label>Email <input type="email" name="email" value="{{ old('email') }}" required></label>
    <label>Password <input type="password" name="password" required></label>
    <label><input type="checkbox" name="remember" value="1"> Remember me</label>
    <button type="submit">Log in</button>
</form>
<a href="/register">Create an account</a>
</body>
</html>
