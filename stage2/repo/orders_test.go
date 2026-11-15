package repo

import (
	"testing"

	"example.com/iterators/mockapi"
)

func TestOrdersFetchesEveryPage(t *testing.T) {
	api := mockapi.New()

	orders := New(api).Orders()

	if len(orders) != 8 {
		t.Fatalf("Orders returned %d orders, want 8", len(orders))
	}
	if api.Requests() != 3 {
		t.Fatalf("Orders made %d requests, want 3", api.Requests())
	}
}
