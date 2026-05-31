package auth

import (
	"sync"
	"time"
)

type rateLimiter struct {
	mu        sync.Mutex
	entries   map[string]*rateEntry
	max       int
	window    time.Duration
	lockout   time.Duration
	lastSweep time.Time
	now       func() time.Time
}

type rateEntry struct {
	count       int
	windowStart time.Time
	lockedUntil time.Time
}

func newRateLimiter(max int, window, lockout time.Duration) *rateLimiter {
	return &rateLimiter{
		entries: make(map[string]*rateEntry),
		max:     max,
		window:  window,
		lockout: lockout,
		now:     time.Now,
	}
}

func (l *rateLimiter) retryAfter(key string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.entries[key]
	if e == nil {
		return 0
	}
	now := l.now()
	if now.Before(e.lockedUntil) {
		return e.lockedUntil.Sub(now)
	}
	return 0
}

func (l *rateLimiter) fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.sweep(now)
	e := l.entries[key]
	if e == nil || now.Sub(e.windowStart) > l.window {
		e = &rateEntry{windowStart: now}
		l.entries[key] = e
	}
	e.count++
	if e.count >= l.max {
		e.lockedUntil = now.Add(l.lockout)
		e.count = 0
		e.windowStart = now
	}
}

func (l *rateLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

func (l *rateLimiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < l.window {
		return
	}
	l.lastSweep = now
	for k, e := range l.entries {
		if now.After(e.lockedUntil) && now.Sub(e.windowStart) > l.window {
			delete(l.entries, k)
		}
	}
}
