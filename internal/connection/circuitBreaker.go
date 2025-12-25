package connection

import (
	"math/rand/v2"
	"time"
)

type BreakerState int

const (
	Closed BreakerState = iota
	Open
	HalfOpen
)

type CircuitBreaker struct {
	state         BreakerState
	failThreshold int  // Number of fails before circuit breaker trips
	successNeeded int  // Number of successes needed for circuit breaker to close after being half-open
	failCount     int  // Current fail count
	successCount  int  // Current success count
	dialSuccess   bool // Status of successful connection
}

func (cb *CircuitBreaker) incrementFails() {
	cb.failCount++
}

func (cb *CircuitBreaker) decrementFails() {
	cb.failCount--
}

func (cb *CircuitBreaker) incrementSuccesses() {
	cb.successCount++
}

func (cb *CircuitBreaker) decrementSuccesses() {
	cb.successCount--
}

func (cb *CircuitBreaker) updateState() {
	if cb.state == Closed {
		if cb.failCount >= cb.failThreshold {
			cb.state = Open
			cb.failCount = 0
		}
	} else if cb.state == HalfOpen {
		if cb.successCount >= cb.successNeeded && cb.failCount <= 0 {
			cb.state = Closed
			cb.successCount = 0
			cb.failCount = 0
		}
	} else {
		if cb.dialSuccess {
			cb.state = HalfOpen
			cb.successCount = 0
			cb.failCount = 0
		}
	}
}

func (cb *CircuitBreaker) waitForPermission(conn Connection) {
	if cb.state == Closed {
		return
	} else if cb.state == Open {

	} else {
		cb.failCount = 0
		cb.successCount = 0

		maxWait := 60000
		baseWait := 1000
		for baseWait > 0 {
			waitTime := rand.Float64() * float64(baseWait)
			time.Sleep(time.Duration(waitTime) * time.Millisecond)
			dialErr := error(nil)
			dialErr = conn.dial()
			if dialErr == nil {
				cb.dialSuccess = true
				cb.updateState()
			} else {
				cb.dialSuccess = false
			}
			if baseWait < maxWait {
				baseWait *= 2
			}
		}
	}
}
