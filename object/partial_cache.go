package object

import (
	"fmt"
)

// This is the basic and bad O(n^2) solution
// We can at least get O(n) read time with a basic trie (although still O(n^2) write time)
// I'm still 50/50 that theres a totally O(n) way
type CacheLine struct {
	args  map[string]string
	Value Object
}

type PartialCache struct {
	lines []CacheLine
}

func (c *PartialCache) Get(args map[string]string) (Object, bool) {
	fmt.Printf("%+v\n", c.lines)
	for _, line := range c.lines {
		didMatch := true
		for indx, val := range line.args {
			if val != args[indx] {
				didMatch = false
				break
			}
		}
		if didMatch {
			return line.Value, true
		}
	}
	return nil, false
}

func (c *PartialCache) Set(args map[string]string, deps map[string]bool, val Object) {
	cachedArgs := make(map[string]string)
	for d, v := range deps {
		if v {
			cachedArgs[d] = args[d]
		}
	}
	c.lines = append(c.lines, CacheLine{cachedArgs, val})
}
