// Package Goroutine provides a small wrapper around go's goroutines, in order to easily create panic safe goroutines.
// Starting a new goroutine without taking care of recovering from a possible panic in that goroutine itself could
// crash the whole application. Therefore, in case of a panic, triggered by the goroutine which was created by the
// Go method, the panic will be automatically recovered and the error will be notified via the done channel.
package goroutine

// The default recover function which will be used by the Go method.
// Can be easily overridden with SetDefaultRecoverFunc in order to change the default behavior.
var defaultRecoverFunc RecoverFunc = func(v interface{}, done chan<- error) {
	done <- ErrPanicRecovered.WithValue(v)
}

// The RecoverFunc type defines the signature of a recover function within a Goroutine.
type RecoverFunc func(v interface{}, done chan<- error)

// Goroutine type contains the function f to run within that goroutine and the recover function rf.
// The recover function rf will be called in case of a panic in f within that goroutine.
type Goroutine struct {
	f  func()      // Will be called in a separate goroutine.
	rf RecoverFunc // Will be called if a panic has been recovered within that goroutine.
}

// The Go method starts a new goroutine which is panic safe.
// A possible panic will be recovered by the recover function, either set by SetDefaultRecoverFunc or WithRecover.
func (g *Goroutine) Go() <-chan error {
	done := make(chan error, 1) // The done channel indicates when a Goroutine has either finished normally or recovered from panic.
	go func() {
		defer func() {
			if r := recover(); r != nil && g.rf != nil {
				// We wrap the recover function in order to prevent an application crash due to a possible panic
				// within the recover function. This ensures, that the app could not crash anymore because of a goroutine panic.
				panicSafeRecover(func() { g.rf(r, done) }, done)
			}
			close(done) // Lastly we need to close the done channel in order to prevent memory leakage.
		}()
		g.f()
	}()
	return done
}

// WithRecover overrides the default recover function with rf.
//  Note: If you pass nil as a RecoverFunc, the panic will be silently recovered.
func (g *Goroutine) WithRecover(rf RecoverFunc) *Goroutine {
	g.rf = rf
	return g
}

// New creates a new panic safe Goroutine, with the defaultRecoverFunc as recover function.
func New(f func()) *Goroutine {
	return &Goroutine{
		f:  f,
		rf: defaultRecoverFunc,
	}
}

// Go runs a function f in a separate goroutine, which does automatically handle the recovering from a panic within that goroutine.
func Go(f func()) <-chan error {
	return New(f).Go()
}

// GetDefaultRecoverFunc returns the current default recover function for goroutines used by the Go method.
func GetDefaultRecoverFunc() RecoverFunc {
	return defaultRecoverFunc
}

// SetDefaultRecoverFunc can be used to override the defaultRecoverFunc which is used by Go method.
//  Note: If you pass nil as a RecoverFunc, the panic will be silently recovered.
func SetDefaultRecoverFunc(rf RecoverFunc) {
	defaultRecoverFunc = rf
}

// panicSafeRecover does guarantee that the goroutine recover function will not crash the application even if it panics.
func panicSafeRecover(f func(), done chan<- error) {
	defer func() {
		if r := recover(); r != nil {
			done <- ErrRecoverFuncPanicRecovered.WithValue(r)
		}
	}()
	f()
}
