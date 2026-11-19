package repo

import (
	"testing"

	"example.com/iterators/mockapi"
)

func TestDeploymentsStopsFetchingWhenCallerBreaks(t *testing.T) {
	api := mockapi.New()
	inspected := 0
	var latestFailure string

	for deployment := range New(api).Deployments() {
		inspected++

		if deployment.Status == mockapi.Failed {
			latestFailure = deployment.ID
			break
		}
	}

	if latestFailure != "deploy-104" {
		t.Fatalf("latest failure = %q, want deploy-104", latestFailure)
	}
	if inspected != 3 {
		t.Fatalf("caller inspected %d deployments, want 3", inspected)
	}
	if api.Requests() != 2 {
		t.Fatalf("Deployments made %d requests, want 2", api.Requests())
	}
}
