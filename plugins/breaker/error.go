package breaker

type CircuitError struct {
	Message string
}

func (e CircuitError) Error() string {
	return "circuit breaker: " + e.Message
}
