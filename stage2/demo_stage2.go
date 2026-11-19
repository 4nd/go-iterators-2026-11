// demo_stage2.go is stage 2 of the pagination refactor.
//
// The repository hides pagination from the caller, but it fetches and stores
// every deployment before the caller can inspect the first one.
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
	deploymentRepo := repo.New(api)

	deployments := deploymentRepo.Deployments()

	for _, deployment := range deployments {
		fmt.Printf("  inspecting %s (%s)\n", deployment.ID, deployment.Status)

		if deployment.Status == mockapi.Failed {
			fmt.Printf("  latest failed deployment: %s\n", deployment.ID)
			break
		}
	}

	fmt.Printf("  API requests: %d\n", api.Requests())
	fmt.Printf("  deployments materialised: %d\n", len(deployments))
}
