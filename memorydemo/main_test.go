package main

import (
	"runtime"
	"testing"
)

func TestEagerAndIteratorProduceTheSameResult(t *testing.T) {
	const (
		total    = 25_001
		pageSize = 10_000
	)

	eagerAPI := NewAPI(total, pageSize)
	eagerChecksum, retained := runEager(eagerAPI)

	iteratorAPI := NewAPI(total, pageSize)
	iteratorChecksum := runIterator(iteratorAPI)

	if eagerChecksum != iteratorChecksum {
		t.Fatalf("checksums differ: eager=%d iterator=%d", eagerChecksum, iteratorChecksum)
	}

	if len(retained) != total {
		t.Fatalf("eager result contains %d deployments, want %d", len(retained), total)
	}

	const expectedRequests = 3
	if eagerAPI.Requests() != expectedRequests {
		t.Fatalf("eager requests = %d, want %d", eagerAPI.Requests(), expectedRequests)
	}
	if iteratorAPI.Requests() != expectedRequests {
		t.Fatalf("iterator requests = %d, want %d", iteratorAPI.Requests(), expectedRequests)
	}

	runtime.KeepAlive(retained)
}
