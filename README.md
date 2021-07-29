# go-routine
A go routine wrapper for an easy approach on panic recovering by default within goroutines.

The purpose of this package is to provide a simple wrapper function for goroutines, which automatically handles panics in goroutines.
Starting a new goroutine without taking care of recovering from a possible panic in that goroutine itself could crash the whole application.

The `Go` function runs an arbitrary function *fn* in a separate goroutine, which handles the recovering from panic within that goroutine.

### Usage

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

### Example

```
package main

import (
	"fmt"
	. "github.com/sknr/go-routine"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var errChan chan string

func init() {
	errChan = make(chan string)
	// Override the default recover function in order to communicate the error via channel.
	SetDefaultRecoverFunc(func(v interface{}) {
		errChan <- fmt.Sprintf("%v", v)
	})
}

func main() {
	exitChan := make(chan struct{})
	interruptChan := make(chan os.Signal)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)

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
			fmt.Println("GoRoutine exists with error: ", err)
		case <-exitChan:
			return
		case <-interruptChan:
			return
		}
	}
}
```