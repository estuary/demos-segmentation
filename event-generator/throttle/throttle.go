package throttle

import "time"

type Throttler struct {
	ticker *time.Ticker
}

func New(interval time.Duration) Throttler {
	return Throttler{ticker: time.NewTicker(interval)}
}

func PerSecond(maxPerSecond int64) Throttler {
	var interval time.Duration = time.Duration(int64(time.Second) / maxPerSecond)

	return New(interval)
}

// Blocking check for readiness. When this returns, it will reset the throttler's timer.
func (t *Throttler) WaitUntilReady() {
	<-t.ticker.C
}

// Non-blocking check for readiness. When this returns `true`, it will reset the throttler's timer.
func (t *Throttler) IsReady() bool {
	select {
	case <-t.ticker.C:
		return true
	default:
		return false
	}
}
