package auth

import (
	"testing"
	"time"
)

func TestRateLimiterLocksAfterMaxFailures(t *testing.T) {
	now := time.Now()
	l := newRateLimiter(3, time.Minute, 5*time.Minute)
	l.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		if d := l.retryAfter("k"); d != 0 {
			t.Fatalf("locked too early after %d failures", i)
		}
		l.fail("k")
	}

	if d := l.retryAfter("k"); d <= 0 {
		t.Fatal("expected lockout after reaching max failures")
	}

	now = now.Add(5*time.Minute + time.Second)
	if d := l.retryAfter("k"); d != 0 {
		t.Fatal("expected lockout to expire")
	}
}

func TestRateLimiterResetClearsFailures(t *testing.T) {
	now := time.Now()
	l := newRateLimiter(3, time.Minute, 5*time.Minute)
	l.now = func() time.Time { return now }

	l.fail("k")
	l.fail("k")
	l.reset("k")
	l.fail("k")
	l.fail("k")

	if d := l.retryAfter("k"); d != 0 {
		t.Fatal("reset should have cleared prior failures")
	}
}

func TestRateLimiterWindowExpiry(t *testing.T) {
	now := time.Now()
	l := newRateLimiter(3, time.Minute, 5*time.Minute)
	l.now = func() time.Time { return now }

	l.fail("k")
	l.fail("k")
	now = now.Add(2 * time.Minute)
	l.fail("k")
	l.fail("k")

	if d := l.retryAfter("k"); d != 0 {
		t.Fatal("failures outside the window should not trigger lockout")
	}
}
