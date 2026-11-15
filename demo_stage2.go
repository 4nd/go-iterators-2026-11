// demo_stage2.go is stage 2 of the pagination refactor.
//
// fetchAllOrders hides pagination from the caller, but it fetches and stores
// every order before the caller can inspect the first one.
//
// Run it with:
//
//	go run demo_stage2.go
//
// It requires Go 1.23 or newer.
package main

import (
	"fmt"

	"example.com/iterators/mockapi"
)

func main() {
	fmt.Println("STAGE 2: EAGER HELPER")
	api := mockapi.New()

	orders := fetchAllOrders(api)

	for _, order := range orders {
		fmt.Printf("  inspecting %s (%s)\n", order.ID, pounds(order.TotalP))

		if order.TotalP > 5_000 {
			fmt.Println("  found one — stop")
			break
		}
	}

	fmt.Printf("  API requests: %d\n", api.Requests())
	fmt.Printf("  orders materialised: %d\n", len(orders))
}

// fetchAllOrders hides pagination, but cannot return until it has fetched and
// retained every page.
func fetchAllOrders(api *mockapi.API) []mockapi.Order {
	var all []mockapi.Order

	for page := 1; ; page++ {
		orders, more := api.FetchPage(page)
		all = append(all, orders...)

		if !more {
			return all
		}
	}
}

func pounds(pence int) string {
	return fmt.Sprintf("£%d.%02d", pence/100, pence%100)
}
