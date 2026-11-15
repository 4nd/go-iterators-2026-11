// demo_stage1.go is stage 1 of the pagination refactor.
//
// The caller fetches pages itself. It avoids unnecessary requests, but the
// pagination details obscure what the caller actually wants to do.
//
// Run it with:
//
//	go run demo_stage1.go
//
// It requires Go 1.23 or newer.
package main

import (
	"fmt"

	"example.com/iterators/mockapi"
)

func main() {
	fmt.Println("STAGE 1: MANUAL PAGINATION")
	api := mockapi.New()

findOrder:
	for page := 1; ; page++ {
		orders, more := api.FetchPage(page)

		for _, order := range orders {
			fmt.Printf("  inspecting %s (%s)\n", order.ID, pounds(order.TotalP))

			if order.TotalP > 5_000 {
				fmt.Println("  found one — stop")
				break findOrder
			}
		}

		if !more {
			break
		}
	}

	fmt.Printf("  API requests: %d\n", api.Requests())
}

func pounds(pence int) string {
	return fmt.Sprintf("£%d.%02d", pence/100, pence%100)
}
