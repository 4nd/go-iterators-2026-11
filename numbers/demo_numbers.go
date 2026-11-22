package main

import (
	"fmt"
	"iter"
	"slices"
)

func main() {
	for i := range 10 {
		fmt.Println(i)
	}

	s := []string{"Hello", "World"}
	for i, v := range s {
		fmt.Println(i, v)
	}

	for n := range fibonacci() {
		fmt.Println(n)
		if n == 13 {
			break
		}
	}

	slices.All()
}

func fibonacci() iter.Seq[int] {
	return func(yield func(int) bool) {
		current, next := 0, 1
		for {
			if !yield(current) {
				return
			}
			current, next = next, current+next
		}
	}
}
