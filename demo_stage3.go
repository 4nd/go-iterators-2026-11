// demo_stage3.go is stage 3 of the pagination refactor.
//
// The iterator hides pagination while still allowing the caller's break to
// stop the producer before it makes another request.
//
// Run it with:
//
//	go run demo_stage3.go
//
// It requires Go 1.23 or newer.
package main

import (
	"fmt"
	"iter"

	"example.com/iterators/mockapi"
)

func main() {
	fmt.Println("STAGE 3: LAZY ITERATOR")
	api := mockapi.New()

	for order := range orders(api) {
		fmt.Printf("  inspecting %s (%s)\n", order.ID, pounds(order.TotalP))

		if order.TotalP > 5_000 {
			fmt.Println("  found one — stop")
			break
		}
	}

	fmt.Printf("  API requests: %d\n", api.Requests())
}

// orders hides pagination and yields each order as soon as it arrives. When
// the caller breaks, yield returns false and orders stops before fetching a
// new page.
func orders(api *mockapi.API) iter.Seq[mockapi.Order] {
	return func(yield func(mockapi.Order) bool) {
		for page := 1; ; page++ {
			pageOrders, more := api.FetchPage(page)

			for _, order := range pageOrders {
				if !yield(order) {
					return
				}
			}

			if !more {
				return
			}
		}
	}
}

func pounds(pence int) string {
	return fmt.Sprintf("£%d.%02d", pence/100, pence%100)
}
