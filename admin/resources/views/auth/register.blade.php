<!doctype html>
<html lang="en">
<body>
<h1>Create an account</h1>
<form method="post" action="/register">
    @csrf
    <label>Name <input name="name" value="{{ old('name') }}" required></label>
    <label>Email <input type="email" name="email" value="{{ old('email') }}" required></label>
    <label>Password <input type="password" name="password" required></label>
    <label>Confirm password <input type="password" name="password_confirmation" required></label>
    <button type="submit">Register</button>
</form>
<a href="/login">Log in</a>
</body>
</html>
