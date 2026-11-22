// memorydemo compares the retained heap of an eager result with an iterator.
//
// Run each mode in a separate process:
//
//	go run ./memorydemo -mode=eager
//	go run ./memorydemo -mode=iterator
package main

import (
	"flag"
	"fmt"
	"iter"
	"os"
	"runtime"
)

type Deployment struct {
	ID         int
	DurationMS int
}

type API struct {
	total    int
	pageSize int
	requests int
}

func NewAPI(total, pageSize int) *API {
	return &API{
		total:    total,
		pageSize: pageSize,
	}
}

// FetchPage generates one page on demand. The API does not retain generated
// deployments, so the two modes start with the same small baseline.
func (api *API) FetchPage(page int) ([]Deployment, bool) {
	start := (page - 1) * api.pageSize
	if start < 0 || start >= api.total {
		return nil, false
	}

	end := min(start+api.pageSize, api.total)
	deployments := make([]Deployment, end-start)

	for index := range deployments {
		id := start + index
		deployments[index] = Deployment{
			ID:         id,
			DurationMS: 100 + id%10_000,
		}
	}

	api.requests++
	return deployments, end < api.total
}

// FetchAllDeployments is the eager version. Preallocating gives it the best
// reasonable case, but the completed slice still retains every deployment.
func (api *API) FetchAllDeployments() []Deployment {
	all := make([]Deployment, 0, api.total)

	for page := 1; ; page++ {
		deployments, more := api.FetchPage(page)
		all = append(all, deployments...)

		if !more {
			return all
		}
	}
}

// Deployments yields deployments one at a time. Once the caller finishes a page, that
// page can be reclaimed instead of becoming part of one large result slice.
func (api *API) Deployments() iter.Seq[Deployment] {
	return func(yield func(Deployment) bool) {
		for page := 1; ; page++ {
			deployments, more := api.FetchPage(page)

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

func (api *API) Requests() int {
	return api.requests
}

func main() {
	//mode := flag.String("mode", "", "eager or iterator")
	mode := new("eager")
	total := flag.Int("deployments", 5_000_000, "number of generated deployments")
	pageSize := flag.Int("page-size", 10_000, "deployments generated per page")
	flag.Parse()

	if *mode != "eager" && *mode != "iterator" {
		fmt.Fprintln(os.Stderr, "-mode must be eager or iterator")
		os.Exit(2)
	}
	if *total < 1 || *pageSize < 1 {
		fmt.Fprintln(os.Stderr, "-deployments and -page-size must be positive")
		os.Exit(2)
	}

	runtime.GC()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	api := NewAPI(*total, *pageSize)
	var (
		checksum uint64
		retained []Deployment
	)

	switch *mode {
	case "eager":
		checksum, retained = runEager(api)
	case "iterator":
		checksum = runIterator(api)
	}

	runtime.GC()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	fmt.Printf("mode:          %s\n", *mode)
	fmt.Printf("deployments:   %d\n", *total)
	fmt.Printf("page size:     %d\n", *pageSize)
	fmt.Printf("pages fetched: %d\n", api.Requests())
	fmt.Printf("checksum:      %d\n", checksum)
	fmt.Printf("retained heap: %s\n", formatBytes(positiveDifference(after.HeapAlloc, before.HeapAlloc)))

	// Keep the eager slice live until after the post-GC measurement.
	runtime.KeepAlive(retained)
}

func runEager(api *API) (uint64, []Deployment) {
	deployments := api.FetchAllDeployments()
	return checksumSlice(deployments), deployments
}

func runIterator(api *API) uint64 {
	var checksum uint64
	for deployment := range api.Deployments() {
		checksum += deploymentChecksum(deployment)
	}
	return checksum
}

func checksumSlice(deployments []Deployment) uint64 {
	var checksum uint64
	for _, deployment := range deployments {
		checksum += deploymentChecksum(deployment)
	}
	return checksum
}

func deploymentChecksum(deployment Deployment) uint64 {
	return uint64(deployment.ID)*31 + uint64(deployment.DurationMS)
}

func positiveDifference(after, before uint64) uint64 {
	if after <= before {
		return 0
	}
	return after - before
}

func formatBytes(bytes uint64) string {
	const (
		kib = 1024
		mib = 1024 * kib
	)

	if bytes >= mib {
		return fmt.Sprintf("%.1f MiB", float64(bytes)/mib)
	}
	return fmt.Sprintf("%.1f KiB", float64(bytes)/kib)
}
