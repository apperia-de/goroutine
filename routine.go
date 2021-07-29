package routine

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

// Implement the GoRoutiner interface methods in order to provide custom Run and Recover a functions for your goroutine
type GoRoutiner interface {
	Run() (fn interface{})
	Recover(v interface{})
}

// The GoRoutine struct implements the GoRoutiner interface
type GoRoutine struct {
	RunFn interface{}         // The Run function which will be called in a goroutine.
	RecFn func(v interface{}) // The Recover function which will be called if a panic has been recovered.
}

// The Run method gets called within a go routine. It can have a panic which will then be gracefully recovered by the Recover method
func (g *GoRoutine) Run() (fn interface{}) {
	return g.RunFn
}

// The Recover method receives the recovered value and processes it
func (g *GoRoutine) Recover(v interface{}) {
	g.RecFn(v)
}

// Go runs an arbitrary function fn in a separate goroutine, which does handle the recovering from panic within that goroutine.
// Starting a new goroutine without taking care of recovering from a possible panic in that goroutine itself could crash the whole application.
// The input param fn must be either GoRoutiner or a generic function, where params contains the possible parameter for that function.
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
	g, ok := fn.(GoRoutiner)
	if !ok {
		fnVal := reflect.ValueOf(fn)
		if fnVal.Kind() != reflect.Func {
			panic(fmt.Sprintf("Param \"fn\" must be a function but is a %q", fnVal.Kind()))
		}
		g = &GoRoutine{
			RunFn: fn,
			RecFn: defaultRecoverFunc,
		}
	}
	// Start the goroutine with recovering in case of a panic.
	go routine(g, params...)
}

// SetDefaultRecoverFunc can be used to override the defaultRecoverFunc which is used by Go if the input param is not a GoRoutiner.
// Note: Calling panic() within the recover function fn, will cause the app to crash if a panic within the goroutine arise.
func SetDefaultRecoverFunc(fn func(v interface{})) {
	defaultRecoverFunc = fn
}

// routine is the helper function which handles the recovering in case the goroutine panics.
func routine(g GoRoutiner, params ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			g.Recover(r)
		}
	}()

	fn := g.Run()
	fnVal := reflect.ValueOf(fn)

	// Check if fn is a function
	if fnVal.Kind() != reflect.Func {
		panic(fmt.Sprintf("Param \"fn\" must be a function but is a %q", fnVal.Kind()))
	}
	// Check input length
	if lp, lf := len(params), fnVal.Type().NumIn(); lp != lf {
		panic(fmt.Sprintf("The number of params (%d) does not match required function params (%d). fn signature: %s", lp, lf, signature(fn)))
	}
	// Convert the input params
	in := make([]reflect.Value, len(params))
	for i, param := range params {
		in[i] = reflect.ValueOf(param)
	}
	// Call the function
	fnVal.Call(in)
}

// signature returns the signature string of a given function
// Credits for that function goes to András Belicza (https://github.com/icza)
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
