// Package mockapi provides the paginated deployment history used by all three stages.
package mockapi

import "fmt"

type Status string

const (
	Succeeded Status = "succeeded"
	Failed    Status = "failed"
)

type Deployment struct {
	ID     string
	Status Status
}

type API struct {
	pages    [][]Deployment
	requests int
}

func New() *API {
	return &API{
		pages: [][]Deployment{
			{
				{ID: "deploy-106", Status: Succeeded},
				{ID: "deploy-105", Status: Succeeded},
			},
			{
				{ID: "deploy-104", Status: Failed},
				{ID: "deploy-103", Status: Succeeded},
			},
			{
				{ID: "deploy-102", Status: Failed},
				{ID: "deploy-101", Status: Succeeded},
			},
		},
	}
}

// FetchPage returns deployments in newest-first order and reports whether more pages exist.
func (api *API) FetchPage(page int) ([]Deployment, bool) {
	api.requests++
	fmt.Printf("  GET /deployments?page=%d\n", page)

	index := page - 1
	return api.pages[index], index < len(api.pages)-1
}

// Requests returns the number of pages fetched.
func (api *API) Requests() int {
	return api.requests
}
