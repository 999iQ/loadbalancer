package limiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity int
	tokens   int
	fillRate time.Duration
	lastFill time.Time
	mux      sync.Mutex
}

func NewTokenBucket(capacity int, fillRate time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		tokens:   capacity,
		fillRate: fillRate,
		lastFill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mux.Lock()
	defer tb.mux.Unlock()

	// Пополнение токенов
	now := time.Now()
	elapsed := now.Sub(tb.lastFill)
	tokensToAdd := int(elapsed / tb.fillRate)

	if tokensToAdd > 0 {
		tb.tokens = min(tb.tokens+tokensToAdd, tb.capacity)
		tb.lastFill = now
	}

	// Проверка токенов
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}
