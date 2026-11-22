package main

import "iter"

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
