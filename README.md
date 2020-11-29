This is a programming language implemented in Go loosely following Thornsten Ball's *[Writing an Interpreter in Go][interpreter-book]*. Notable divergences are:

* Float type
* Various builtin functions
* Pure functions

This is the documentation of the programming language. If there is any missing documentation, incorrect functionality, or any other issue, [please file an issue][new-issue].

To run a repl, execute the following command after cloning the repo:

`go run main.go`


## Types

### Boolean

* There are two boolean values: `true` and `false`

* `bool(value)` casts any value to its boolean value. Notably, each type has a zero or empty value which will evaluate to `false`

```
>> bool(true)
true
```

```
>> bool(0)
false
```
### String

* Strings are written using `""` 

`"This is a string"`. 

* `+` if an argument on either side of a `+` side is a string, the result will be a concatenation of the string and the other argument's string value

```
>> "two " + "strings"
two strings

>> 1 + "two"
1two
```

* `*` If an integer and a string are multiplied, the result will be the string repeated the integer number of times

```
>> "string" * 3
stringstringstring
```

* `string(value)` casts any value to its string value

### Integer

* Integers are non-decimal numbers
* They support a few basic arithmetic operations: `+`, `-`, `*`, `/`. They follow basic order of operations. Dividing two integers will result in a float.

```
>> 1 + 2 * 3
7
>> 1 / 1
1.0
>> type(2 / 2)
FLOAT
```


* `int(value)` casts any value to its integer value

```
>> int("3") == 3
true
```

### Float

* Floats are decimal numbers
* They support basic arithmetic operations: `+`, `-`, `*`, `/`. Combining a float and integer using artihmetic will result in a float type

```
>> 1.0 + 2
3.0
>> 1.5 * (2 / 3)
1.0
```

* `float(value)` casts any value to its float value

```
>> float(7)
7.0
>> float(true)
1.0
```

### Array

* Arrays are comma-separated lists of values denoted with `[]`. For example: `[1, "array", false]`

* `first(array)` returns the first element of an array

```
>> first([1,2,3])
1
```
* `last(array)` returns the first element of an array

```
>> last([1,2,3])
3
```
* `rest(array)` returns the array without its first element

```
>> rest([1,2,3])
[2,3]
```
* `take(array,n)` returns the first n elements of an array

```
>> take([1,2,3],2)
[1,2]
```
* `drop(array,n)` returns the array without its first n elements

```
>> drop([1,2,3],2)
[3]
```
* `push(array,arg)` returns the array with `arg` appended to the end

```
>> push([1,2,3],"four")
[1,2,3,"four"]
```

* `+` returns the addition of two arrays

```
>> [1,2] + [3,4]
[1,2,3,4]
```

### Hash

* Hashes are maps from keys to values denoted with `{}`, with keys and their values separated by `:`, and pairs separated by `,`
* Hash keys can be integers, floats, booleans or strings.

```
>> { "a" : 1, 1: [3,4,5], true: 73.2 }
```


* `keys` returns an array of the keys of a hash. No ordering is guaranteed

```
>> keys({ "a" : 1, 1: [3,4,5], true: 73.2 })
[1, true, a]
```
* `values` returns an array of the values of a hash. No ordering is guaranteed

```
>> values({ "a" : 1, 1: [3,4,5], true: 73.2 })
[1, [3,4,5], 73.2]
```


* `+` returns the addition of two hashes, preferring the value of the second hash, for any keys they share

```
>> { 1: 1, 2: 2 } + { 1: "second", 3: 3 }
{ 1: second, 2: 2, 3: 3 }
```

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

Returns the type of an object

```
>> type(true)
BOOLEAN
>> type([])
ARRAY
```

### len(obj)

Returns the length of an object

```
>> len([1,2,3])
3
>> len("string")
6
>> len({ 1: 1, 2: true })
2
```

### rando(int)

Returns a random integer in the range [0,int)

```
>> rando(10)
3
```


## Whitespace

Whitespace is not relevant in this programming language. Use as much, or as little, as you'd like. There is not (yet) any existing style guide. Although I really like Go's formatter, so might bake some auto-formatted style into the language at some future point!

[interpreter-book]: https://interpreterbook.com/
[new-issue]: https://github.com/jemmaissroff/go-interpreter/issues/new