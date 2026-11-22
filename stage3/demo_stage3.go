// demo_stage3.go is stage 3 of the pagination refactor.
//
// The repository iterator hides pagination and stops fetching when the search
// returns from the loop.
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
	deploymentRepo := repo.New(api)

	latestFailure, found := findLatestFailedDeployment(deploymentRepo)
	if found {
		fmt.Printf("  latest failed deployment: %s\n", latestFailure.ID)
	}

	fmt.Printf("  API requests: %d\n", api.Requests())
}

func findLatestFailedDeployment(deploymentRepo *repo.Repository) (mockapi.Deployment, bool) {
	for deployment := range deploymentRepo.Deployments() {
		fmt.Printf("  inspecting %s (%s)\n", deployment.ID, deployment.Status)

		if deployment.Status == mockapi.Failed {
			return deployment, true
		}
	}

	return mockapi.Deployment{}, false
}
