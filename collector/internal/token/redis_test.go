package token

import "testing"

func TestRedisTokenKey(t *testing.T) {
	const token = "clk_example"
	const want = "collector:token:6f1197acbb1dd6e493fd704af2af3d53eda4ddfcbe93a0a1fc8fda909b62bdbe"

	if got := redisTokenKey(token); got != want {
		t.Errorf("redisTokenKey() = %q, want %q", got, want)
	}
}
