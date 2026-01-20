package mouser

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	// DefaultRequestsPerMinute is the default rate limit per minute.
	DefaultRequestsPerMinute = 30

	// DefaultRequestsPerDay is the default rate limit per day.
	DefaultRequestsPerDay = 1000
)

// RateLimiter implements a dual rate limiter for minute and daily limits.
type RateLimiter struct {
	mu sync.Mutex

	// Minute rate limiting
	requestsPerMinute int
	minuteTokens      int
	lastMinuteReset   time.Time

	// Daily rate limiting
	requestsPerDay int
	dailyTokens    int
	lastDayReset   time.Time

	// Server-indicated backoff (from Retry-After header)
	blockedUntil time.Time
}

// NewRateLimiter creates a new RateLimiter with the specified limits.
func NewRateLimiter(requestsPerMinute, requestsPerDay int) *RateLimiter {
	now := time.Now()
	return &RateLimiter{
		requestsPerMinute: requestsPerMinute,
		minuteTokens:      requestsPerMinute,
		lastMinuteReset:   now,
		requestsPerDay:    requestsPerDay,
		dailyTokens:       requestsPerDay,
		lastDayReset:      now,
	}
}

// Wait blocks until a request can be made or the context is cancelled.
// It returns an error if the daily limit is exceeded or the context is cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		now := time.Now()

		// Check server-indicated backoff first
		if now.Before(r.blockedUntil) {
			waitTime := r.blockedUntil.Sub(now)
			r.mu.Unlock()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
				continue
			}
		}

		// Reset minute tokens if a minute has passed
		if now.Sub(r.lastMinuteReset) >= time.Minute {
			r.minuteTokens = r.requestsPerMinute
			r.lastMinuteReset = now
		}

		// Reset daily tokens if a day has passed
		if now.Sub(r.lastDayReset) >= 24*time.Hour {
			r.dailyTokens = r.requestsPerDay
			r.lastDayReset = now
		}

		// Check daily limit first
		if r.dailyTokens <= 0 {
			timeUntilReset := r.lastDayReset.Add(24 * time.Hour).Sub(now)
			r.mu.Unlock()
			return fmt.Errorf("%w: resets in %v", ErrDailyLimitExceeded, timeUntilReset.Round(time.Minute))
		}

		// Check minute limit
		if r.minuteTokens > 0 {
			r.minuteTokens--
			r.dailyTokens--
			r.mu.Unlock()
			return nil
		}

		// Calculate wait time until minute reset
		waitTime := r.lastMinuteReset.Add(time.Minute).Sub(now)
		r.mu.Unlock()

		// Wait for either the timer or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue loop to try again
		}
	}
}

// TryAcquire attempts to acquire a rate limit token without blocking.
// Returns true if successful, false if rate limited.
func (r *RateLimiter) TryAcquire() (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Check server-indicated backoff
	if now.Before(r.blockedUntil) {
		return false, ErrRateLimitExceeded
	}

	// Reset minute tokens if a minute has passed
	if now.Sub(r.lastMinuteReset) >= time.Minute {
		r.minuteTokens = r.requestsPerMinute
		r.lastMinuteReset = now
	}

	// Reset daily tokens if a day has passed
	if now.Sub(r.lastDayReset) >= 24*time.Hour {
		r.dailyTokens = r.requestsPerDay
		r.lastDayReset = now
	}

	// Check daily limit
	if r.dailyTokens <= 0 {
		return false, ErrDailyLimitExceeded
	}

	// Check minute limit
	if r.minuteTokens <= 0 {
		return false, ErrRateLimitExceeded
	}

	r.minuteTokens--
	r.dailyTokens--
	return true, nil
}

// UpdateFromResponse updates the rate limiter based on server response.
// retryAfterSeconds is the value from the Retry-After header (0 if not present).
func (r *RateLimiter) UpdateFromResponse(retryAfterSeconds int) {
	if retryAfterSeconds <= 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Cap the backoff to a reasonable maximum (5 minutes)
	if retryAfterSeconds > 300 {
		retryAfterSeconds = 300
	}

	blockedUntil := time.Now().Add(time.Duration(retryAfterSeconds) * time.Second)
	if blockedUntil.After(r.blockedUntil) {
		r.blockedUntil = blockedUntil
	}
}

// RemainingMinute returns the number of requests remaining in the current minute.
func (r *RateLimiter) RemainingMinute() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if now.Sub(r.lastMinuteReset) >= time.Minute {
		return r.requestsPerMinute
	}
	return r.minuteTokens
}

// RemainingDaily returns the number of requests remaining today.
func (r *RateLimiter) RemainingDaily() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if now.Sub(r.lastDayReset) >= 24*time.Hour {
		return r.requestsPerDay
	}
	return r.dailyTokens
}

// Stats returns current rate limit statistics.
type RateLimitStats struct {
	MinuteRemaining int
	DailyRemaining  int
	BlockedUntil    time.Time
}

// Stats returns current rate limit statistics.
func (r *RateLimiter) Stats() RateLimitStats {
	r.mu.Lock()
	defer r.mu.Unlock()

	return RateLimitStats{
		MinuteRemaining: r.minuteTokens,
		DailyRemaining:  r.dailyTokens,
		BlockedUntil:    r.blockedUntil,
	}
}
