package goroutine

import (
	"fmt"
)

var (
	// ErrPanicRecovered is returned via done channel if a panic arises within the goroutine.
	ErrPanicRecovered = &goroutineError{msg: "panic recovered"}

	// ErrRecoverFunctionPanicked is returned via done channel if a panic arises within the
	// goroutine and also within the recover function.
	ErrRecoverFunctionPanicked = &goroutineError{msg: "recover function panicked"}

	// The default recover function which will be used by the Go and Goroutine functions for each new goroutine.
	// Can be easily overridden with SetDefaultRecoverFunc in an init function in order to change the default behavior.
	defaultRecoverFunc = func(v interface{}, done chan<- error) {
		done <- ErrPanicRecovered.WithValue(v)
	}
)

// Goroutine creates a new panic safe goroutine, with the defaultRecoverFunc as recover function.
func Goroutine(f func()) *goroutine {
	return &goroutine{
		f:    f,
		recF: defaultRecoverFunc,
	}
}

// Go runs a function f in a separate goroutine, which does handle the recovering from panic within that goroutine.
// Starting a new goroutine without taking care of recovering from a possible panic in that goroutine itself could crash the whole application.
// Functions with more than one input param must be wrapped within a func(){} like in the example below.
// Instead of running:
//
// go func(s string) {
//   panic(s)
// }("Hello World")
//
// simply call:
//
// Go(func() {
//   func(s string) {
//     panic(s)
//   }("Hello World")
// })
//
func Go(f func()) <-chan error {
	return Goroutine(f).Go()
}

// SetDefaultRecoverFunc can be used to override the defaultRecoverFunc which is used by Go and Goroutine functions.
// Note: Calling panic() within the recover function recF, will cause the application to crash.
//       If you pass nil as a RecoverFunc, the panic will be silently recovered.
func SetDefaultRecoverFunc(recF RecoverFunc) {
	defaultRecoverFunc = recF
}

// GetDefaultRecoverFunc returns the current default recover function for goroutines used by the Go and Goroutine functions.
func GetDefaultRecoverFunc() RecoverFunc {
	return defaultRecoverFunc
}

// The RecoverFunc type defines the signature of a recover function within a goroutine.
type RecoverFunc func(v interface{}, done chan<- error)

// goroutineError is the goroutine specific error type
type goroutineError struct {
	msg string      // Goroutine error type
	v   interface{} // Goroutine error value
}

// WithValue allows to attach the error received by the recover function
func (r *goroutineError) WithValue(v interface{}) *goroutineError {
	r.v = v
	return r
}

// Error implements the error interface
func (r *goroutineError) Error() string {
	return fmt.Sprintf("[goroutine] %s: %v", r.msg, r.v)
}

// goroutine type contains the function to run in that go routine and the recover function recF.
// The recover function recF will be called in case of a panic in f within that goroutine.
type goroutine struct {
	f    func()      // The f function which will be called in a goroutine.
	recF RecoverFunc // The recF function which will be called if a panic has been recovered with that goroutine.
}

// The Go method starts a new goroutine, which is panic safe. A possible panic call will be gracefully recovered by the recover function g.recF.
func (g *goroutine) Go() <-chan error {
	done := make(chan error, 1) // The done channel indicates when a goroutine has either finished or recovered from panic.
	go func() {
		defer func() {
			if r := recover(); r != nil && g.recF != nil {
				// We wrap the recover function in order to prevent an application crash due to a possible panic
				// within the recover function. This ensures, that the app could not crash anymore because of a goroutine panic.
				panicSafeRecover(func() { g.recF(r, done) }, done)
			}
			// Lastly we need to close the done channel in order to prevent memory leakage.
			close(done)
		}()
		g.f()
	}()
	return done
}

// WithRecoverFunc overrides the default recover function with recF.
// Note: Calling panic() within the recover function recF, will cause the application to crash.
//       If you pass nil as a RecoverFunc, the panic will be silently recovered.
func (g *goroutine) WithRecoverFunc(recF RecoverFunc) *goroutine {
	g.recF = recF
	return g
}

// panicSafeRecover does guarantee that the recover function will not crash the application even if it panics.
func panicSafeRecover(recF func(), done chan<- error) {
	defer func() {
		if r := recover(); r != nil {
			done <- ErrRecoverFunctionPanicked.WithValue(r)
		}
	}()
	recF()
}
