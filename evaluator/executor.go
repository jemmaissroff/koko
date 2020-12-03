package evaluator

import (
	"io/ioutil"
	"koko/lexer"
	"koko/object"
	"koko/parser"
	"strings"
)

func LoadProgramFromFile(fileLocation string, env *object.Environment) object.Object {
	data, err := ioutil.ReadFile(fileLocation)

	if err != nil {
		return newError("File reading error %v", fileLocation)
	}
	return LoadProgram(string(data), env)
}

func ExecuteProgram(programStr string) string {
	env := object.NewEnvironment()
	evaluated := LoadProgram(programStr, env)

	if evaluated != nil {
		return evaluated.Inspect()
	}
	return ""
}

func LoadProgram(programStr string, env *object.Environment) object.Object {
	l := lexer.New(programStr)
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return newError("Eeeeek! Parsing Errors Found: \n" + strings.Join(p.Errors(), "\n"))
	}

	return Eval(program, env)
}
