package goroutine

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestGoroutine(t *testing.T) {
	// Given some example functions
	f0 := func() {
		fmt.Println("Hallo Welt")
	}
	f1 := func() {
		func(name string) {
			fmt.Printf("Hello %s", name)
		}("GoWorld")
	}
	f2 := func() {
		func(a, b int) {
			fmt.Printf("%d / %d = %d", a, b, a/b)
		}(42, 2)
	}
	f3 := func() {
		func(a, b int) {
			fmt.Printf("%d / %d = %d", a, b, a/b)
		}(42, 0)
	}
	f4 := func() {
		panic("Error in goroutine")
	}
	recF0 := func(v interface{}, done chan<- error) {
		done <- fmt.Errorf("%v", v)
	}
	recF1 := func(v interface{}, done chan<- error) {
		panic("OH NO! Panic in recover function")
	}

	tests := []struct {
		name     string
		f        func()
		want     error
		expected string
	}{
		{"Goroutine with a zero param function", f0, nil, "Hallo Welt\n"},
		{"Goroutine with a one param function", f1, nil, "Hello GoWorld"},
		{"Goroutine with a two param function", f2, nil, "42 / 2 = 21"},
		{"Goroutine with a two param function", f3, ErrPanicRecovered, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := recordStdOut(func() {
				got := <-Goroutine(test.f).Go()
				assertError(t, got, test.want)
			})
			assertOutput(t, result, test.expected)
		})
	}

	t.Run("Goroutine with a two param function and a custom recover function which recovered from a panic", func(t *testing.T) {
		got := <-Goroutine(f3).WithRecoverFunc(recF0).Go()
		want := "runtime error: integer divide by zero"
		assertOutput(t, got.Error(), want)
	})

	t.Run("Goroutine with recover function which panics should never raise an application crash", func(t *testing.T) {
		<-Goroutine(f4).WithRecoverFunc(recF1).Go()
	})
}

func TestGo(t *testing.T) {
	resultChan := make(chan string)
	// Example function which panicked in goroutine
	f := func() {
		func(a, b int) {
			resultChan <- fmt.Sprintf("%d / %d = %d", a, b, a/b)
		}(10, 0)
	}

	originalRecoverFunc := GetDefaultRecoverFunc()

	t.Run("Goroutine with a two param function which panicked in recover func and recovered", func(t *testing.T) {
		SetDefaultRecoverFunc(func(v interface{}, done chan<- error) { panic("panic in recover func") })
		got := <-Go(f)
		want := ErrRecoverFunctionPanicked.WithValue("panic in recover func")
		if got == nil {
			t.Errorf("Expected a goroutineError, but got none")
		}
		assertError(t, got, want)
		assertOutput(t, got.Error(), "[goroutine] recover function panicked: panic in recover func")
	})

	// Restore defaultRecoverFunc
	SetDefaultRecoverFunc(originalRecoverFunc)
}

func TestDefaultRecoverFunc(t *testing.T) {
	t.Run("defaultRecoverFunc: panic value is an error", func(t *testing.T) {
		done := make(chan error, 1)
		defaultRecoverFunc(errors.New("panic value is an error"), done)
		got := <-done
		want := &goroutineError{
			msg: "panic recovered",
			v:   errors.New("panic value is an error"),
		}
		assertError(t, got, want)
	})

	t.Run("defaultRecoverFunc: panic value is an error message", func(t *testing.T) {
		done := make(chan error, 1)
		defaultRecoverFunc("panic value is an error message", done)
		got := <-done
		want := &goroutineError{
			msg: "panic recovered",
			v:   "panic value is an error message",
		}
		assertError(t, got, want)
	})
}

func assertOutput(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertError(t *testing.T, got, want error) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func recordStdOut(f func()) string {
	old := os.Stdout // Keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f() // Run the function

	outC := make(chan string)
	// Copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()
	// Back to normal state
	_ = w.Close()
	os.Stdout = old // Restoring the real stdout
	return <-outC
}
