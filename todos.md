Figure out why this isn't working as expected:
let fib = fn(x) { if (x == 1) { 1 } else { if (x ==0) { 1} else { fib(x - 1) + fib(x - 2) }}
fib("a")
=> 1
???


