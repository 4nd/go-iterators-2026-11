// Package repo adapts the paginated API for the Stage 2 caller.
package repo

import "example.com/iterators/mockapi"

type Repository struct {
	api *mockapi.API
}

func New(api *mockapi.API) *Repository {
	return &Repository{api: api}
}

// Deployments hides pagination, but cannot return until it has fetched and retained
// every page.
func (repository *Repository) Deployments() []mockapi.Deployment {
	var all []mockapi.Deployment

	for page := 1; ; page++ {
		deployments, more := repository.api.FetchPage(page)
		all = append(all, deployments...)

		if !more {
			return all
		}
	}
}
