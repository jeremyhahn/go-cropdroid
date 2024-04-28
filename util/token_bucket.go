package util

import (
	"math"
	"time"
)

type TokenBucket struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64
	lastRefillTime time.Time
}

func NewTokenBucket() *TokenBucket {
	return CreateTokenBucket(25, 1) // 25 requests per second
}

func CreateTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	duration := now.Sub(tb.lastRefillTime)
	tokensToAdd := tb.refillRate * duration.Seconds()
	tb.tokens = math.Min(tb.tokens+tokensToAdd, tb.maxTokens)
	tb.lastRefillTime = now
}

func (tb *TokenBucket) Request(tokens float64) bool {
	tb.refill()
	if tokens <= tb.tokens {
		tb.tokens -= tokens
		return true
	}
	return false
}
