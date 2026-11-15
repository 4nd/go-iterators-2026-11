// demo_stage2.go is stage 2 of the pagination refactor.
//
// The repository hides pagination from the caller, but it fetches and stores
// every order before the caller can inspect the first one.
//
// Run it with:
//
//	go run ./stage2
//
// It requires Go 1.23 or newer.
package main

import (
	"fmt"

	"example.com/iterators/mockapi"
	"example.com/iterators/stage2/repo"
)

func main() {
	fmt.Println("STAGE 2: EAGER REPOSITORY")
	api := mockapi.New()
	orderRepo := repo.New(api)

	orders := orderRepo.Orders()

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

func pounds(pence int) string {
	return fmt.Sprintf("£%d.%02d", pence/100, pence%100)
}
