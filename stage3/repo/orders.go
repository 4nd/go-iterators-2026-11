// Package repo adapts the paginated API for the Stage 3 caller.
package repo

import (
	"iter"

	"example.com/iterators/mockapi"
)

type Repository struct {
	api *mockapi.API
}

func New(api *mockapi.API) *Repository {
	return &Repository{api: api}
}

// Orders hides pagination and yields each order as soon as it arrives. When
// the caller breaks, yield returns false and the repository stops before
// fetching a new page.
func (repository *Repository) Orders() iter.Seq[mockapi.Order] {
	return func(yield func(mockapi.Order) bool) {
		for page := 1; ; page++ {
			orders, more := repository.api.FetchPage(page)

			for _, order := range orders {
				if !yield(order) {
					return
				}
			}

			if !more {
				return
			}
		}
	}
}
