package ipratelimiter

import (
	"encoding/binary"
	"net"
	"sync"
	"testing"
)

func int2ip(nn int) string {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, uint32(nn))
	return ip.String()
}

func Test_IPRateLimiter_goroutine_safed(t *testing.T) {
	limit := 100000
	ipr := NewIPRateLimiter(limit)

	var wg sync.WaitGroup
	wg.Add(limit)
	for i := 0; i < limit; i++ {
		go func(IPNum int) {
			defer wg.Done()
			ipr.GetLimiter(int2ip(IPNum))
		}(i) // should pass i into function or will be captured by func literal
	}
	wg.Wait()

	if got := ipr.totalIP(); got != limit {
		t.Errorf("IPRateLimiter.totalIP() = %v, want %v", got, limit)
	}
}

func Test_IPRateLimiter_totalIP(t *testing.T) {
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
