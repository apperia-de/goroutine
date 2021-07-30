# goroutine
[![Go Report Card](https://goreportcard.com/badge/github.com/sknr/goroutine)](https://goreportcard.com/report/github.com/sknr/goroutine)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/sknr/goroutine?style=flat)
![GitHub](https://img.shields.io/github/license/sknr/goroutine)


A goroutine wrapper for creating and running panic safe goroutines.

The purpose of this package is to provide a simple wrapper function for goroutines, which automatically handles panics
in goroutines. Starting a new goroutine without taking care of recovering from a possible panic in that goroutine itself
could crash the whole application.

The `Go` function runs an arbitrary function *fn* in a separate goroutine, which handles the recovering from panic
within that goroutine.

## Usage (with dot import)

Instead of running

```
go func(s string) {
    panic(s)
}("Hello World")
```

simply call

```
Go(func(s string) {
    panic(s)
}, "Hello World")
```

in order to create a panic safe goroutine.

## Examples

```
package main

import (
    "fmt"
    . "github.com/sknr/goroutine"
    "os"
    "os/signal"
    "syscall"
    "time"
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
```

### Create a new goroutine with a custom recover function

In order to override the default recover function for goroutines, simply provide an alternative implementation for
with `goroutine.SetDefaultRecoverFunc`.

If you need different recover functions for different goroutines, you can simply call

```
Goroutine(func(name string) {
    panic(fmt.Sprintln("Hallo", name))
}).WithRecoverFunc(func(v interface{}) {
    log.Printf("Custom recover function in goroutine, with error: %v", v)
}).Go("Welt")
```
