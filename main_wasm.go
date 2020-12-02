package main

import (
	"fmt"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"strings"
	"syscall/js"
)

func executeProgram(programStr string) string {
	env := object.NewEnvironment()

	l := lexer.New(programStr)
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return "Eeeeek! Parsing Errors Found: \n" + strings.Join(p.Errors(), "\n")
	}

	evaluated := evaluator.Eval(program, env)
	if evaluated != nil {
		return evaluated.Inspect()
	}
	return ""
}

func main() {
	// register an empty channel
	c := make(chan struct{}, 0)
	js.Global().Set("executeProgram", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			// Simplified code
			programStr := args[0].String()
			js.Global().Set("output", executeProgram(programStr))
		}()
		return nil
	}))
	fmt.Println("interpreter successfully loaded")
	// wait forever without spinning (I think)
	<-c
}
