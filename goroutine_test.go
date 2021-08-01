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
	f0 := func() {
		resultChan <- "Hallo Welt"
	}
	f1 := func() {
		func(name string) {
			resultChan <- fmt.Sprintf("Hallo %s", name)
		}("GoWorld")
	}
	f2 := func() {
		func(a, b int) {
			resultChan <- fmt.Sprintf("%d / %d = %d", a, b, a/b)
		}(42, 2)
	}
	f3 := func() {
		func(a, b int) {
			resultChan <- fmt.Sprintf("%d / %d = %d", a, b, a/b)
		}(42, 0)
	}
	f4 := func() {
		panic("Error in goroutine")
	}
	recF0 := func(v interface{}) {
		resultChan <- fmt.Sprintf("%v", v)
	}
	recF1 := func(v interface{}) {
		panic(nil)
	}

	tests := []struct {
		name string
		f    func()
		want string
	}{
		{"Goroutine with a zero param function", f0, "Hallo Welt"},
		{"Goroutine with a one param function", f1, "Hallo GoWorld"},
		{"Goroutine with a two param function", f2, "42 / 2 = 21"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Goroutine(test.f).Go()
			got := <-resultChan
			assertOutput(t, got, test.want)
		})
	}

	t.Run("Goroutine with a two param function and a custom recover function which recovered from a panic", func(t *testing.T) {
		Goroutine(f3).WithRecoverFunc(recF0).Go()
		got := <-resultChan
		want := "runtime error: integer divide by zero"
		assertOutput(t, got, want)
	})

	t.Run("Goroutine with recover function which panics should never raise an application crash", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("The code did panic")
			}
		}()
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
		SetDefaultRecoverFunc(func(v interface{}) { panic("Panic in recover func") })
		Go(f)
		SetDefaultRecoverFunc(originalRecoverFunc)
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

func assertOutput(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
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
