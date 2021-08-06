package goroutine

import "fmt"

var (
	// ErrPanicRecovered is returned when a goroutine has panicked.
	ErrPanicRecovered = &panicError{message: "panic in goroutine recovered", value: nil}

	// ErrRecoverFuncPanicRecovered is returned when the recover function of a goroutine has panicked.
	ErrRecoverFuncPanicRecovered = &panicError{message: "panic in recover function of goroutine recovered", value: nil}
)

// panicError indicates recovered panic values as errors which might occur in the Goroutine.
type panicError struct {
	message string      // Custom error message
	value   interface{} // Recovered panic value
}

// Error returns the error as a string.
func (pe *panicError) Error() string {
	if pe.value == nil {
		return pe.message
	}
	return fmt.Sprintf("%s: %v", pe.message, pe.value)
}

// WithValue returns a copy of the current panicError with a custom value.
func (pe *panicError) WithValue(v interface{}) *panicError {
	pe.value = v
	return pe
}
