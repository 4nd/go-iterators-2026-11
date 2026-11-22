package main

import (
	"fmt"
)

func main() {
	for i := range 10 {
		fmt.Print(i)
	}

	s := []string{"Hello", "World"}
	for _, v := range s {
		fmt.Println(v)
	}

}
