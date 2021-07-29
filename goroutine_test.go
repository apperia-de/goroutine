package goroutine

import (
	"fmt"
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

	t.Run("Goroutine with a two param function which recovered from a panic", func(t *testing.T) {
		SetDefaultRecoverFunc(func(v interface{}) { resultChan <- fmt.Sprintf("%v", v) })
		Go(fn2, 2, 0)
		got := <-resultChan
		want := "runtime error: integer divide by zero"
		assertOutput(t, got, want)
	})
}

func assertOutput(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
