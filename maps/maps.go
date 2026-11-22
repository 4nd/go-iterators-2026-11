package main

import (
	"fmt"
	"iter"
)

type Collection []int

func (c Collection) Map(transformation func(int) int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, x := range c {
			if !yield(transformation(x)) {
				return
			}
		}
	}
}

func main() {

	coll := Collection{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	squared := coll.Map(func(x int) int {
		return x * x
	})

	for s := range squared {
		fmt.Println(s)
	}
}
