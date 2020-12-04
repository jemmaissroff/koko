package main

import (
	"fmt"
	"koko/executor"
	"syscall/js"
)

func main() {
	// register an empty channel
	c := make(chan struct{}, 0)
	js.Global().Set("executeProgram", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			// Simplified code
			programStr := args[0].String()
			js.Global().Set("output", executor.ExecuteProgram(programStr))
		}()
		return nil
	}))
	fmt.Println("interpreter successfully loaded")
	// wait forever without spinning (I think)
	<-c
}
