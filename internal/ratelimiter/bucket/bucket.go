package bucket

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity      int
	tokens        int // текущее кол-во
	rate          int // скорость пополнения
	lastCheckTime time.Time
	mux           sync.Mutex
}

// NewTokenBucket - конструктор TokenBucket
func NewTokenBucket(rate, burst int) *TokenBucket {
	return &TokenBucket{
		capacity:      burst,
		tokens:        burst,
		rate:          rate,
		lastCheckTime: time.Now(),
	}
}

// Allow - проверяет и обновляет количество токенов для пропуска запроса
func (tb *TokenBucket) Allow() bool {
	tb.mux.Lock()
	defer tb.mux.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastCheckTime) // временная дельта с последней проверки
	tb.lastCheckTime = now

	tokensToAdd := int(elapsed.Seconds() * float64(tb.rate))
	if tokensToAdd > 0 {
		// следим за переполнением токенов в бакете
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		} else {
			tb.tokens += tokensToAdd
		}
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}
