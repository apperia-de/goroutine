# goroutine

[![Go Report Card](https://goreportcard.com/badge/github.com/sknr/goroutine)](https://goreportcard.com/report/github.com/sknr/goroutine)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/sknr/goroutine?style=flat)
![GitHub](https://img.shields.io/github/license/sknr/goroutine)

A goroutine wrapper for creating and running panic safe goroutines.

The purpose of this package is to provide a simple wrapper function for goroutines, which automatically handles panics
in goroutines. Starting a new goroutine without taking care of recovering from a possible panic in that goroutine itself
could crash the whole application.

The `Go` function runs an arbitrary function *f* in a separate goroutine, which handles the recovering from panic within
that goroutine.

## Usage (with dot import)

Instead of running

```
go func() {
    panic("Panic raised in goroutine")
}()
```

simply call

```
Go(func() {
    panic("Panic raised in goroutine")
})
```
in order to create a panic safe goroutine.

Functions with multiple input params must be wrapped within an anonymous function. 

Instead of running


```
go func(a, b int) {
    panic(a+b)
}(21,21)
```

simply call

```
Go(func() {
    func(a, b int) {
        panic(a+b)
    }(21,21)
})
```

## Examples

```
package main

import (
	"fmt"
	. "github.com/sknr/goroutine"
)

var errChan chan string

func init() {
	errChan = make(chan string)

	// Override the default recover function.
	SetDefaultRecoverFunc(func(v interface{}) {
		errChan <- fmt.Sprintf("%v", v)
	})
}

func main() {
	exitChan := make(chan struct{})
	cnt := 1
	for {
	    select {
		case <-Go(func() {
			func(a, b int) {
				fmt.Println("Divide by zero", a/b)
			}(1, 0)
		}):
			fmt.Printf("Never reached due to panic in goroutine")
		case err := <-errChan:
			fmt.Println("Goroutine exits with error: ", err)

		case <-exitChan:
			fmt.Println("Exit program")
			return
		}

		if cnt >= 3 {
			close(exitChan)
			fmt.Println("Channel closed")
		}
		cnt++
	}
}

```

### Create a new goroutine with a custom recover function

In order to override the default recover function for goroutines, simply provide an alternative implementation for
with `goroutine.SetDefaultRecoverFunc`.

If you need different recover functions for different goroutines, you can simply call

```
Goroutine(func() {
    func(name string) {
        panic(fmt.Sprintln("Hallo", name))
    }("Welt")
}).WithRecoverFunc(func(v interface{}, done chan<- error) {
    log.Printf("Custom recover function in goroutine, with error: %v", v)
}).Go()
```
