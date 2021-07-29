package goroutine

import (
	"fmt"
	"reflect"
	"strings"
)

// The default recover function which will be used by the Go func for each goroutine.
// Can be easily overridden with SetDefaultRecoverFunc in an init function in order to change the default behavior.
var defaultRecoverFunc = func(v interface{}) {
	if err, ok := v.(error); ok {
		fmt.Printf("Error(%T): %v\n", err, err)
		return
	}
	fmt.Printf("Goroutine panic recovered: %v\n", v)
}

// The RecoverFunc type defines the signature of a recover function within a goroutine.
type RecoverFunc func(v interface{})

type goroutine struct {
	goFn  interface{} // The goFn function which will be called in a goroutine.
	recFn RecoverFunc // The recFn function which will be called if a panic has been recovered with that goroutine.
}

// The Go method starts a new goroutine, which is panic safe. A possible panic call will be gracefully recovered by the recover function g.recFn.
func (g *goroutine) Go(params ...interface{}) {
	go func() {
		defer func() {
			if r := recover(); r != nil && g.recFn != nil {
				g.recFn(r)
			}
		}()
		fnVal := reflect.ValueOf(g.goFn)
		// Check input length
		if lp, lf := len(params), fnVal.Type().NumIn(); lp != lf {
			panic(fmt.Sprintf("The number of params (%d) does not match required function params (%d). Function signature: %s", lp, lf, signature(g.goFn)))
		}
		// Convert the input params
		in := make([]reflect.Value, len(params))
		for i, param := range params {
			in[i] = reflect.ValueOf(param)
		}
		// Call the function
		fnVal.Call(in)
	}()
}

// WithRecoverFunc overrides the default recover function with fn.
// Note: If you pass nil as a RecoverFunc, the panic will be silently recovered.
func (g *goroutine) WithRecoverFunc(fn RecoverFunc) *goroutine {
	g.recFn = fn
	return g
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
func Go(fn interface{}, params ...interface{}) {
	fnVal := reflect.ValueOf(fn)
	if fnVal.Kind() != reflect.Func {
		panic(fmt.Sprintf("Param \"fn\" must be a function but is a %q", fnVal.Kind()))
	}
	Goroutine(fn).Go(params...)
}

// SetDefaultRecoverFunc can be used to override the defaultRecoverFunc which is used by Go and Goroutine if the input param is not a GoRoutiner.
// Note: Calling panic() within the recover function fn, will cause the app to crash if a panic within the goroutine arise.
func SetDefaultRecoverFunc(fn func(v interface{})) {
	defaultRecoverFunc = fn
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
