This is a programming language implemented in Go loosely following Thornsten Ball's *[Writing an Interpreter in Go][interpreter-book]*. Notable divergences are:

* Float type
* Various builtin functions
* Pure functions

This is the documentation of the programming language. If there is any missing documentation, incorrect functionality, or any other issue, [please file an issue][new-issue].

To run a repl, execute the following command after cloning the repo:

`go run main.go`


## Types

### Bool

### String

### Integer

### Float

### Array 

### Hash

## Comparison

`==` returns true if two values are equal, else false


`!=` returns false if two values are equal, else true

## Variables

The syntax for assignment to variables is:

`let name = value`

For instance:

`let i = 1`

`let str = "some string"` 

## If statements

This language has `if` and `else` available. It does not have a concept of `else if` as that would be superfluous with the existing `if` and `else`. Syntax for `if` blocks is as follows:

`if (condition) { code } else { code }`

For example:

`if (1 == 2) { "the same" } else { "different" }`

## Functions

Functions are defined using 

`fn(arguments) { code }`

They can take any number of comma-separated arguments. They can also be composed.

For example:

```
let multiplier = fn(x) { fn(n) { x * n } }
let double = multiplier(2)
double(10)
=> 20
``` 

### Pure Functions

TODO: Peter to fill in this section!


## Builtins

These are all builtin functions which are not specific to a type. Any type-specific builtin functions are documented above in the `Type` section. Like all functions, builtins are called with some number of arguments in parentheses. 

### builtins()

Returns an alphabetically sorted array of all available builtin functions, including those which are type-speficific.

### type(obj)

Returns the type of an object. For example:

```
type(true)
BOOLEAN
```

### rando(int)

Returns a random integer in the range [0,int)



## Whitespace

Whitespace is not relevant in this programming language. Use as much, or as little, as you'd like. There is not (yet) any existing style guide. Although I really like Go's formatter, so might bake some auto-formatted style into the language at some future point!

[interpreter-book]: https://interpreterbook.com/
[new-issue]: https://github.com/jemmaissroff/go-interpreter/issues/new