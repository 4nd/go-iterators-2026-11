// demo_stage1.go is stage 1 of the pagination refactor.
//
// The caller fetches pages itself. It avoids unnecessary requests, but the
// pagination details obscure what the caller actually wants to do.
//
// Run it with:
//
//	go run ./stage1
//
// It requires Go 1.23 or newer.
package main

import (
	"fmt"

	"example.com/iterators/paging/mockapi"
)

func main() {
	fmt.Println("STAGE 1: MANUAL PAGINATION")
	api := mockapi.New()

	found := false
	for page := 1; !found; page++ {
		deployments, more := api.FetchPage(page)

		for _, deployment := range deployments {
			fmt.Printf("  inspecting %s (%s)\n", deployment.ID, deployment.Status)

			if deployment.Status == mockapi.Failed {
				fmt.Printf("  latest failed deployment: %s\n", deployment.ID)
				found = true
				break
			}
		}

		if !more {
			break
		}
	}

	fmt.Printf("  API requests: %d\n", api.Requests())
}
