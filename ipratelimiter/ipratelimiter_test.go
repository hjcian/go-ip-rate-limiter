package ipratelimiter

import (
	"testing"
)

func TestIPRateLimiter_totalIP(t *testing.T) {
	reqLimitPerMin := 2
	i := NewIPRateLimiter(reqLimitPerMin)
	tests := []struct {
		name string
		ip   string
		want int
	}{
		{
			"Get 1.1.1.1",
			"1.1.1.1",
			1,
		},
		{
			"Get 1.1.1.1",
			"1.2.3.4",
			2,
		},
		{
			"Get 1.1.1.1",
			"1.1.1.4",
			3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i.GetLimiter(tt.ip)
			if got := i.totalIP(); got != tt.want {
				t.Errorf("IPRateLimiter.totalIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
