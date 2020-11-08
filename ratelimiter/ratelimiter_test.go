package ratelimiter

import (
	"testing"
	"time"
)

func Test_rateLimiter_Allow(t *testing.T) {
	limit := 2
	r := NewRateLimiter(limit)

	tests := []struct {
		name  string
		want1 bool
		want2 *RateLimitStatus
	}{
		{
			"1 - OK",
			true,
			&RateLimitStatus{
				limit,
				1,
				time.Now().Unix(),
				1,
			},
		},
		{
			"2 - OK",
			true,
			&RateLimitStatus{
				limit,
				0,
				time.Now().Unix(),
				2,
			},
		},
		{
			"3 - Error",
			false,
			&RateLimitStatus{
				limit,
				0,
				time.Now().Unix(),
				2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got1, got2 := r.Allow()
			if got1 != tt.want1 {
				t.Errorf("[isAllow] rateLimiter.Allow() got = %v, want %v", got1, tt.want1)
			}
			if got2.RatelimitLimitUsed != tt.want2.RatelimitLimitUsed {
				t.Errorf(
					"[RatelimitLimitUsed] rateLimiter.Allow() got = %v, want %v",
					got2.RatelimitLimitUsed,
					tt.want2.RatelimitLimitUsed,
				)
			}
			if got2.RatelimitLimitRemaining != tt.want2.RatelimitLimitRemaining {
				t.Errorf(
					"[RatelimitLimitRemaining] rateLimiter.Allow() got = %v, want %v",
					got2.RatelimitLimitRemaining,
					tt.want2.RatelimitLimitRemaining,
				)
			}
		})
	}
}