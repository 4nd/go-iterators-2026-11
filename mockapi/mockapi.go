// Package mockapi provides the unchanged paginated API used by every demo stage.
package mockapi

import "fmt"

type Order struct {
	ID     string
	TotalP int
}

type API struct {
	pages    [][]Order
	requests int
}

func New() *API {
	return &API{
		pages: [][]Order{
			{
				{ID: "A-100", TotalP: 1_250},
				{ID: "A-101", TotalP: 6_200},
				{ID: "A-102", TotalP: 900},
			},
			{
				{ID: "A-103", TotalP: 4_500},
				{ID: "A-104", TotalP: 2_100},
				{ID: "A-105", TotalP: 7_500},
			},
			{
				{ID: "A-106", TotalP: 3_250},
				{ID: "A-107", TotalP: 1_100},
			},
		},
	}
}

func (api *API) FetchPage(page int) ([]Order, bool) {
	api.requests++
	fmt.Printf("  GET /orders?page=%d\n", page)

	index := page - 1
	return api.pages[index], index < len(api.pages)-1
}

// Requests returns the number of pages fetched.
func (api *API) Requests() int {
	return api.requests
}
