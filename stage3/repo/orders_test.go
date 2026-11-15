package repo

import (
	"testing"

	"example.com/iterators/mockapi"
)

func TestOrdersStopsFetchingWhenCallerBreaks(t *testing.T) {
	api := mockapi.New()
	inspected := 0

	for order := range New(api).Orders() {
		inspected++

		if order.TotalP > 5_000 {
			break
		}
	}

	if inspected != 2 {
		t.Fatalf("caller inspected %d orders, want 2", inspected)
	}
	if api.Requests() != 1 {
		t.Fatalf("Orders made %d requests, want 1", api.Requests())
	}
}
