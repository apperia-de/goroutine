package goroutine

import (
	"fmt"
	"reflect"
	"strings"
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
func Goroutine(fn interface{}) *goroutine {
	fnVal := reflect.ValueOf(fn)
	if fnVal.Kind() != reflect.Func {
		panic(fmt.Sprintf("Param \"fn\" must be a function but is a %q", fnVal.Kind()))
	}

	return &goroutine{
		goFn:  fn,
		recFn: defaultRecoverFunc,
		done:  make(chan struct{}),
	}
}

// Go runs an arbitrary function fn in a separate goroutine, which does handle the recovering from panic within that goroutine.
// Starting a new goroutine without taking care of recovering from a possible panic in that goroutine itself could crash the whole application.
// The input param fn must be a generic function, where params contains the possible parameter for that function, otherwise a panic occurs.
// Instead of running:
//
// go func(s string) {
//   panic(s)
// }("Hello World")
//
// Simply call:
//
// Go(func(s string) {
//   panic(s)
// }, "Hello World")
//
func Go(fn interface{}, params ...interface{}) DoneChan {
	fnVal := reflect.ValueOf(fn)
	if fnVal.Kind() != reflect.Func {
		panic(fmt.Sprintf("Param \"fn\" must be a function but is a %q", fnVal.Kind()))
	}

	gr := Goroutine(fn)
	gr.Go(params...)
	return gr.done
}

// SetDefaultRecoverFunc can be used to override the defaultRecoverFunc which is used by Go and Goroutine functions.
// Note: Calling panic() within the recover function fn, will cause the application to crash if a panic within the goroutine arise.
//       If you pass nil as a RecoverFunc, the panic will be silently recovered.
func SetDefaultRecoverFunc(fn RecoverFunc) {
	defaultRecoverFunc = fn
}

// GetDefaultRecoverFunc returns the current default recover function for goroutines used by the Go and Goroutine functions.
func GetDefaultRecoverFunc() RecoverFunc {
	return defaultRecoverFunc
}

type goroutine struct {
	goFn  interface{}   // The goFn function which will be called in a goroutine.
	recFn RecoverFunc   // The recFn function which will be called if a panic has been recovered with that goroutine.
	done  chan struct{} // The done channel indicates when a goroutine has either finished or recovered from panic.
}

// The Go method starts a new goroutine, which is panic safe. A possible panic call will be gracefully recovered by the recover function g.recFn.
func (g *goroutine) Go(params ...interface{}) DoneChan {
	go func() {
		defer func() {
			if r := recover(); r != nil && g.recFn != nil {
				g.recFn(r)
				g.done <- struct{}{}
			}
		}()
		fnVal := reflect.ValueOf(g.goFn)
		// Check input length
		if lp, lf := len(params), fnVal.Type().NumIn(); lp != lf {
			panic(fmt.Sprintf("Function signature: %s | The number of params (%d) does not match required function params (%d)\n", signature(g.goFn), lp, lf))
		}
		// Convert the input params
		in := make([]reflect.Value, len(params))
		for i, param := range params {
			in[i] = reflect.ValueOf(param)
		}
		// Call the function
		fnVal.Call(in)
		g.done <- struct{}{}
	}()
	return g.done
}

// WithRecoverFunc overrides the default recover function with fn.
// Note: Calling panic() within the recover function fn, will cause the application to crash if a panic within the goroutine arise.
//       If you pass nil as a RecoverFunc, the panic will be silently recovered.
func (g *goroutine) WithRecoverFunc(fn RecoverFunc) *goroutine {
	g.recFn = fn
	return g
}

// signature returns the signature string of a given function.
// Credits for that function goes to András Belicza (https://github.com/icza).
func signature(fn interface{}) string {
	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func {
		return "<not a function>"
	}

	buf := strings.Builder{}
	buf.WriteString("func(")
	for i := 0; i < t.NumIn(); i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(t.In(i).String())
	}
	buf.WriteString(")")
	if numOut := t.NumOut(); numOut > 0 {
		if numOut > 1 {
			buf.WriteString(" (")
		} else {
			buf.WriteString(" ")
		}
		for i := 0; i < t.NumOut(); i++ {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(t.Out(i).String())
		}
		if numOut > 1 {
			buf.WriteString(")")
		}
	}

	return buf.String()
}
