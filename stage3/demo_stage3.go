// demo_stage3.go is stage 3 of the pagination refactor.
//
// The repository iterator hides pagination while still allowing the caller's
// break to stop the producer before it makes another request.
//
// Run it with:
//
//	go run ./stage3
//
// It requires Go 1.23 or newer.
package main

import (
	"fmt"

	"example.com/iterators/mockapi"
	"example.com/iterators/stage3/repo"
)

func main() {
	fmt.Println("STAGE 3: LAZY REPOSITORY")
	api := mockapi.New()
	orderRepo := repo.New(api)

	for order := range orderRepo.Orders() {
		fmt.Printf("  inspecting %s (%s)\n", order.ID, pounds(order.TotalP))

		if order.TotalP > 5_000 {
			fmt.Println("  found one — stop")
			break
		}
	}

	fmt.Printf("  API requests: %d\n", api.Requests())
}

func pounds(pence int) string {
	return fmt.Sprintf("£%d.%02d", pence/100, pence%100)
}
