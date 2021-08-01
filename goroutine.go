package goroutine

import (
	"fmt"
)

// The RecoverFunc type defines the signature of a recover function within a goroutine.
type RecoverFunc func(v interface{})

// The DoneChan type is defines the channel which receives only an empty struct.
type DoneChan <-chan struct{}

// The default recover function which will be used by the Go func for each goroutine.
// Can be easily overridden with SetDefaultRecoverFunc in an init function in order to change the default behavior.
var defaultRecoverFunc = func(v interface{}) {
	if err, ok := v.(error); ok {
		fmt.Printf("Error(%T): %v\n", err, err)
		return
	}
	fmt.Printf("Goroutine panic recovered: %v\n", v)
}

// Goroutine creates a new panic safe goroutine, with the defaultRecoverFunc as recover function.
func Goroutine(f func()) *goroutine {
	return &goroutine{
		f:    f,
		recF: defaultRecoverFunc,
	}
}

// Go runs an arbitrary function f in a separate goroutine, which does handle the recovering from panic within that goroutine.
// Starting a new goroutine without taking care of recovering from a possible panic in that goroutine itself could crash the whole application.
// The input param f must be a generic function, where params contains the possible parameter for that function, otherwise a panic occurs.
// Instead of running:
//
// go func(s string) {
//   panic(s)
// }("Hello World")
//
// Simply call:
//
// Go(func() {
//   func(s string) {
//     panic(s)
//   }("Hello World")
// })
//
func Go(f func()) DoneChan {
	return Goroutine(f).Go()
}

// SetDefaultRecoverFunc can be used to override the defaultRecoverFunc which is used by Go and Goroutine functions.
// Note: Calling panic() within the recover function f, will cause the application to crash if a panic within the goroutine arise.
//       If you pass nil as a RecoverFunc, the panic will be silently recovered.
func SetDefaultRecoverFunc(f RecoverFunc) {
	defaultRecoverFunc = f
}

// GetDefaultRecoverFunc returns the current default recover function for goroutines used by the Go and Goroutine functions.
func GetDefaultRecoverFunc() RecoverFunc {
	return defaultRecoverFunc
}

// goroutine type defines contains the function to run in that go routine and the recover function recF,
// which will be called in case of a panic which might occur in f.
type goroutine struct {
	f    func()      // The f function which will be called in a goroutine.
	recF RecoverFunc // The recF function which will be called if a panic has been recovered with that goroutine.
}

// The Go method starts a new goroutine, which is panic safe. A possible panic call will be gracefully recovered by the recover function g.recF.
func (g *goroutine) Go() DoneChan {
	done := make(chan struct{}) // The done channel indicates when a goroutine has either finished or recovered from panic.
	go func() {
		defer func() {
			if r := recover(); r != nil && g.recF != nil {
				// We wrap the recover function in order to prevent an application crash due to a possible panic
				// within the recover function. Therefore the app could not crash anymore because of a goroutine panic.
				<-Goroutine(func() { g.recF(r) }).
					WithRecoverFunc(func(v interface{}) { fmt.Printf("Recover function panicked: %v\n", v) }).
					Go()
			}
			done <- struct{}{}
		}()
		g.f()
		done <- struct{}{}
	}()
	return done
}

// WithRecoverFunc overrides the default recover function with f.
// Note: Calling panic() within the recover function f, will cause the application to crash if a panic within the goroutine arise.
//       If you pass nil as a RecoverFunc, the panic will be silently recovered.
func (g *goroutine) WithRecoverFunc(f RecoverFunc) *goroutine {
	g.recF = f
	return g
}
