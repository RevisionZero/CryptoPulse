package connection

// BreakerState models the standard closed/open/half-open circuit breaker lifecycle.
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

// incrementFails records a failed operation and recomputes breaker state.
func (cb *CircuitBreaker) incrementFails() {
	cb.failCount++
	cb.updateState()
}

// decrementFails allows callers to compensate fail counts when needed.
func (cb *CircuitBreaker) decrementFails() {
	cb.failCount--
	cb.updateState()
}

// incrementSuccesses records a successful operation and recomputes breaker state.
func (cb *CircuitBreaker) incrementSuccesses() {
	cb.successCount++
	cb.updateState()
}

// decrementSuccesses allows callers to compensate success counts when needed.
func (cb *CircuitBreaker) decrementSuccesses() {
	cb.successCount--
	cb.updateState()
}

// setDialState updates breaker connectivity knowledge after a dial attempt.
func (cb *CircuitBreaker) setDialState(dialErr error) {
	if dialErr != nil {
		cb.dialSuccess = false
	} else {
		cb.dialSuccess = true
	}
	cb.updateState()
}

// updateState transitions breaker state according to configured fail/success thresholds.
func (cb *CircuitBreaker) updateState() {
	if cb.state == Closed {
		if cb.failCount >= cb.failThreshold {
			cb.state = Open
			cb.failCount = 0
			cb.successCount = 0
		}
	} else if cb.state == HalfOpen {
		if cb.successCount >= cb.successNeeded && cb.failCount <= 0 {
			cb.state = Closed
			cb.successCount = 0
			cb.failCount = 0
		} else if cb.failCount >= cb.failThreshold {
			cb.state = Open
			cb.failCount = 0
			cb.successCount = 0
		}
	} else {
		if cb.dialSuccess {
			cb.state = HalfOpen
			cb.successCount = 0
			cb.failCount = 0
		}
	}
}

// requestPermission reports whether the caller may proceed with a request.
func (cb *CircuitBreaker) requestPermission() bool {
	if cb.state == Closed || cb.state == HalfOpen {
		return true
	} else {
		return false
	}
}
