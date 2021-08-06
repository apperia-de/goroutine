# goroutine

[![Go Report Card](https://goreportcard.com/badge/github.com/sknr/goroutine)](https://goreportcard.com/report/github.com/sknr/goroutine)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/sknr/goroutine?style=flat)
![GitHub](https://img.shields.io/github/license/sknr/goroutine)

A goroutine wrapper for creating and running panic safe goroutines.

The purpose of this package is to provide a simple wrapper function for goroutines, which automatically handles panics
in goroutines. Starting a new goroutine without taking care of recovering from a possible panic in that goroutine itself
could crash the whole application.

The `Go` function runs an arbitrary function `f` in a separate goroutine, which handles the recovering from panic within
that goroutine.

## Installation

`go get -u github.com/sknr/goroutine`

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
    "github.com/sknr/goroutine"
    "log"
)

func init() {
    // Override the default recover function.
    goroutine.SetDefaultRecoverFunc(func(v interface{}, done chan<- error) {
        log.Printf("%v", v)
        done <- fmt.Errorf("panic in goroutine successfully recovered")
    })
}

func main() {
    for i := -3; i <= 3; i++ {
        err := <-goroutine.Go(func() {
            func(a, b int) {
                log.Println(a, "/", b, "=", a/b)
            }(10, i)
        })
        if err != nil {
            log.Println(err)
        }
    }
}
```

### Set a new default recover function

In order to override the default recover function for new goroutines (created with `Go(func())` or `New(func())`),
simply set a custom recover function of type `RecoverFunc` with `SetDefaultRecoverFunc`

### Create a new Goroutine instance with a custom recover function

If you need different recover functions for different goroutines, you can simply call

```
goroutine.New(func() {
    func(name string) {
        panic(fmt.Sprintln("Hallo", name))
    }("Welt")
}).WithRecoverFunc(func(v interface{}, done chan<- error) {
    log.Printf("Custom recover function in goroutine, with error: %v", v)
}).Go()
```
