package repo

import (
	"testing"

	"example.com/iterators/mockapi"
)

func TestDeploymentsFetchesEveryPage(t *testing.T) {
	api := mockapi.New()

	deployments := New(api).Deployments()

	if len(deployments) != 6 {
		t.Fatalf("Deployments returned %d deployments, want 6", len(deployments))
	}
	if api.Requests() != 3 {
		t.Fatalf("Deployments made %d requests, want 3", api.Requests())
	}
}
