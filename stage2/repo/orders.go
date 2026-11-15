// Package repo adapts the paginated API for the Stage 2 caller.
package repo

import "example.com/iterators/mockapi"

type Repository struct {
	api *mockapi.API
}

func New(api *mockapi.API) *Repository {
	return &Repository{api: api}
}

// Orders hides pagination, but cannot return until it has fetched and retained
// every page.
func (repository *Repository) Orders() []mockapi.Order {
	var all []mockapi.Order

	for page := 1; ; page++ {
		orders, more := repository.api.FetchPage(page)
		all = append(all, orders...)

		if !more {
			return all
		}
	}
}
