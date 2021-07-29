package examples

import (
	"fmt"
	. "github.com/sknr/goroutine"
	"log"
	"time"
)

func One() {
	errChan := make(chan string)
	SetDefaultRecoverFunc(func(v interface{}) {
		errChan <- fmt.Sprintf("%v", v)
	})

	exitChan := make(chan struct{})

	Go(func() {
		for i := 3; i > 0; i-- {
			Go(func(a, b int) {
				fmt.Println("Divide by zero", a/b)
			}, 1, 0)
			time.Sleep(1 * time.Second)
		}
		fmt.Println("Exit goroutine")
		close(exitChan)
	})

	for {
		select {
		case err := <-errChan:
			fmt.Println("Goroutine exits with error: ", err)
		case <-exitChan:
			return
		}
	}
}

func Two() {
	oldRecoverFunc := GetDefaultRecoverFunc()
	errChan := make(chan string, 0)
	SetDefaultRecoverFunc(func(v interface{}) {
		errChan <- fmt.Sprintf("%v", v)
	})

	fmt.Print("\nGoroutines with custom recover function:\n\n")
	Goroutine(func(a, b int) {
		fmt.Println("Divide by zero", a/b)
	}).Go(1, 0)
	fmt.Println("Goroutine exits with error: ", <-errChan)

	Goroutine(func(a, b int) {
		fmt.Println("Divide by zero", a/b)
	}).WithRecoverFunc(func(v interface{}) {
		log.Printf("Goroutine recovered from panic: %v", v)
		fmt.Println("The panic in this goroutine has been successfully recovered and error has been logged.")
		errChan <- fmt.Sprintf("%v", v)
	}).Go(1, 0)
	fmt.Println("Error:",<-errChan)

	Goroutine(func(a, b int) {
		fmt.Println(a, "/", b, "=", a/b)
	}).Go(10, 5)

	SetDefaultRecoverFunc(oldRecoverFunc)

	fmt.Print("\nGoroutines with default recover function:\n\n")
	Goroutine(func(a, b int) {
		fmt.Println("Divide by zero", a/b)
	}).Go(1, 0)
	Goroutine(func(a, b int) {
		fmt.Println(a, "/", b, "=", a/b)
	}).Go(10, 5)

	time.Sleep(500 * time.Millisecond)
}
