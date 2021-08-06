package goroutine_test

import (
	"bytes"
	"fmt"
	"github.com/sknr/goroutine"
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
		panic("panicError in Goroutine")
	}
	rf0 := func(v interface{}, done chan<- error) {
		done <- fmt.Errorf("%v", v)
	}
	rf1 := func(v interface{}, done chan<- error) {
		panic("OH NO! Panic in recover function")
	}
	rf2 := func(v interface{}, done chan<- error) {
		done <- goroutine.ErrPanicRecovered.WithValue(nil)
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := recordStdOut(func() {
				got := <-goroutine.New(test.f).Go()
				assertError(t, got, test.want)
			})
			assertOutput(t, result, test.expected)
		})
	}

	t.Run("Goroutine with a two param function and a custom recover function which recovered from a panic", func(t *testing.T) {
		got := <-goroutine.New(f3).WithRecover(rf0).Go()
		want := "runtime error: integer divide by zero"
		assertOutput(t, got.Error(), want)
	})

	t.Run("Goroutine with recover function which panics should never raise an application crash", func(t *testing.T) {
		<-goroutine.New(f4).WithRecover(rf1).Go()
	})

	t.Run("Goroutine panic recovered with plain ErrPanicRecovered", func(t *testing.T) {
		got := <-goroutine.New(f4).WithRecover(rf2).Go()
		want := "panic in goroutine recovered"
		assertOutput(t, got.Error(), want)
	})
}

func TestGo(t *testing.T) {
	resultChan := make(chan string)
	// Example function which panicked in Goroutine
	f := func() {
		func(a, b int) {
			resultChan <- fmt.Sprintf("%d / %d = %d", a, b, a/b)
		}(10, 0)
	}

	originalRecoverFunc := goroutine.GetDefaultRecoverFunc()

	t.Run("Goroutine with a two param function which panicked in recover func and recovered", func(t *testing.T) {
		goroutine.SetDefaultRecoverFunc(func(v interface{}, done chan<- error) { panic("panic in recover func") })
		got := <-goroutine.Go(f)
		want := goroutine.ErrRecoverFuncPanicRecovered.WithValue("panic in recover func")
		if got == nil {
			t.Errorf("Expected a panicError, but got none")
		}
		assertError(t, got, want)
	})

	// Restore defaultRecoverFunc
	goroutine.SetDefaultRecoverFunc(originalRecoverFunc)
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
