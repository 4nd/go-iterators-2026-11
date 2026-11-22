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

// Deployments yields deployments newest first. If the caller exits the loop
// early, yield returns false and the repository stops before fetching another page.
func (repository *Repository) Deployments() iter.Seq[mockapi.Deployment] {
	return func(yield func(mockapi.Deployment) bool) {
		for page := 1; ; page++ {
			deployments, more := repository.api.FetchPage(page)

			for _, deployment := range deployments {
				if !yield(deployment) {
					return
				}
			}

			if !more {
				return
			}
		}
	}
}
