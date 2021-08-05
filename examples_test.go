package goroutine_test

import (
	"fmt"
	"github.com/sknr/goroutine"
)

func ExampleGo() {
	// Instead of
	go func() {
		values := [3]int{1, 2, 3}
		for i := 0; i < 4; i++ {
			fmt.Println(values[i])
		}
	}()

	// simply call
	goroutine.Go(func() {
		values := [3]int{1, 2, 3}
		for i := 0; i < 4; i++ {
			fmt.Println(values[i])
		}
	})
}

func ExampleGo_withInputParam() {
	// Functions with input params need to be wrapped by an anonymous function.

	// Instead of
	go func(s string) {
		panic(s)
	}("Hello World")

	// simply call
	goroutine.Go(func() {
		func(s string) {
			panic(s)
		}("Hello World")
	})
}

func ExampleNew() {
	err := <-goroutine.New(func() {
		values := [3]int{1, 2, 3}
		for i := 0; i < 4; i++ {
			fmt.Println(values[i])
		}
	}).Go()

	fmt.Println(err)
	// Output:
	// 1
	// 2
	// 3
	// panic in goroutine recovered: runtime error: index out of range [3] with length 3
}

func ExampleGoroutine_WithRecover() {
	err := <-goroutine.New(func() {
		values := [3]int{1, 2, 3}
		for i := 0; i < 4; i++ {
			fmt.Println(values[i])
		}
	}).WithRecover(func(v interface{}, done chan<- error) {
		if err, ok := v.(error); ok {
			done <- err
			return
		}
		done <- fmt.Errorf("recovered: %v", v)
	}).Go()
	fmt.Println(err)
	// Output:
	// 1
	// 2
	// 3
	// runtime error: index out of range [3] with length 3
}
