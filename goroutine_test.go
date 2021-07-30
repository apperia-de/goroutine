package goroutine

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestGoroutine(t *testing.T) {
	resultChan := make(chan string)
	// Given some example functions
	fn0 := func() {
		resultChan <- "Hallo Welt"
	}
	fn1 := func(name string) {
		resultChan <- fmt.Sprintf("Hallo %s", name)
	}
	fn2 := func(a, b int) {
		resultChan <- fmt.Sprintf("%d / %d = %d", a, b, a/b)
	}
	recFn1 := func(v interface{}) {
		resultChan <- fmt.Sprintf("%v", v)
	}

	tests := []struct {
		name   string
		fn     interface{}
		params []interface{}
		want   string
	}{
		{"Goroutine with a zero param function", fn0, []interface{}{}, "Hallo Welt"},
		{"Goroutine with a one param function", fn1, []interface{}{"GoWorld"}, "Hallo GoWorld"},
		{"Goroutine with a two param function", fn2, []interface{}{42, 2}, "42 / 2 = 21"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Goroutine(test.fn).Go(test.params...)
			got := <-resultChan
			assertOutput(t, got, test.want)
		})
	}

	t.Run("Goroutine with a two param function and a custom recover function which recovered from a panic", func(t *testing.T) {
		Goroutine(fn2).WithRecoverFunc(recFn1).Go(2, 0)
		got := <-resultChan
		want := "runtime error: integer divide by zero"
		assertOutput(t, got, want)
	})

	t.Run("Starting a Goroutine with wrong type which should raise a panic", func(t *testing.T) {
		assertPanic(t, func() { Goroutine(1) })
	})
}

func TestGo(t *testing.T) {
	resultChan := make(chan string)
	// Given some example functions
	fn0 := func() {
		resultChan <- "Hallo Welt"
	}
	fn1 := func(name string) {
		resultChan <- fmt.Sprintf("Hallo %s", name)
	}
	fn2 := func(a, b int) {
		resultChan <- fmt.Sprintf("%d / %d = %d", a, b, a/b)
	}

	tests := []struct {
		name   string
		fn     interface{}
		params []interface{}
		want   string
	}{
		{"Goroutine with a zero param function", fn0, []interface{}{}, "Hallo Welt"},
		{"Goroutine with a one param function", fn1, []interface{}{"GoWorld"}, "Hallo GoWorld"},
		{"Goroutine with a two param function", fn2, []interface{}{10, 2}, "10 / 2 = 5"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Go(test.fn, test.params...)
			got := <-resultChan
			assertOutput(t, got, test.want)
		})
	}

	originalRecoverFunc := GetDefaultRecoverFunc()

	t.Run("Goroutine with a two param function which recovered from a panic", func(t *testing.T) {
		SetDefaultRecoverFunc(func(v interface{}) { resultChan <- fmt.Sprintf("%v", v) })
		Go(fn2, 2, 0)
		got := <-resultChan
		want := "runtime error: integer divide by zero"
		assertOutput(t, got, want)
		SetDefaultRecoverFunc(originalRecoverFunc)
	})

	t.Run("Goroutine with a wrong number of params which recovered from a panic", func(t *testing.T) {
		SetDefaultRecoverFunc(func(v interface{}) { resultChan <- fmt.Sprintf("%v", v) })
		Go(fn0, 1)
		got := <-resultChan
		want := "Function signature: func() | The number of params (1) does not match required function params (0)\n"
		assertOutput(t, got, want)
		SetDefaultRecoverFunc(originalRecoverFunc)
	})

	t.Run("Starting a Goroutine with wrong type which should raise a panic", func(t *testing.T) {
		assertPanic(t, func() { Go(1) })
	})
}

func TestDefaultRecoverFunc(t *testing.T) {
	t.Run("defaultRecoverFunc: panic value is an error", func(t *testing.T) {
		got := recordStdOut(func() { defaultRecoverFunc(errors.New("panic value is an error")) })
		want := "Error(*errors.errorString): panic value is an error\n"
		assertOutput(t, got, want)
	})

	t.Run("defaultRecoverFunc: panic value is an error", func(t *testing.T) {
		got := recordStdOut(func() { defaultRecoverFunc("panic value is a string message") })
		want := "Goroutine panic recovered: panic value is a string message\n"
		assertOutput(t, got, want)
	})
}

func TestSignature(t *testing.T) {
	fn0 := struct{}{}
	fn1 := func() {}
	fn2 := func(a, b int) {}
	fn3 := func(a, b int) string { return "" }
	fn4 := func(a int) (string, error) { return "", nil }

	tests := []struct {
		name string
		fn   interface{}
		want string
	}{
		{"Is not a function", fn0, "<not a function>"},
		{"Is a function without input params", fn1, "func()"},
		{"Is a function with one input param", fn2, "func(int, int)"},
		{"Is a function with two input params and one output param", fn3, "func(int, int) string"},
		{"Is a function with one input params and two output params", fn4, "func(int) (string, error)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := signature(test.fn)
			assertOutput(t, got, test.want)
		})
	}
}

func assertOutput(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertPanic(t *testing.T, fn func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	fn() // Run the panic function
}

func recordStdOut(fn func()) string {
	old := os.Stdout // Keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn() // Run the function

	outC := make(chan string)
	// Copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, _  = io.Copy(&buf, r)
		outC <- buf.String()
	}()
	// Back to normal state
	_ = w.Close()
	os.Stdout = old // Restoring the real stdout
	return <-outC
}
