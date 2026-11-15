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

type Order struct {
	ID     int
	TotalP int
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
// orders, so the two modes start with the same small baseline.
func (api *API) FetchPage(page int) ([]Order, bool) {
	start := (page - 1) * api.pageSize
	if start < 0 || start >= api.total {
		return nil, false
	}

	end := min(start+api.pageSize, api.total)
	orders := make([]Order, end-start)

	for index := range orders {
		id := start + index
		orders[index] = Order{
			ID:     id,
			TotalP: 100 + id%10_000,
		}
	}

	api.requests++
	return orders, end < api.total
}

// FetchAllOrders is the eager version. Preallocating gives it the best
// reasonable case, but the completed slice still retains every order.
func (api *API) FetchAllOrders() []Order {
	all := make([]Order, 0, api.total)

	for page := 1; ; page++ {
		orders, more := api.FetchPage(page)
		all = append(all, orders...)

		if !more {
			return all
		}
	}
}

// Orders yields one page at a time. Once the caller finishes a page, that
// page can be reclaimed instead of becoming part of one large result slice.
func (api *API) Orders() iter.Seq[Order] {
	return func(yield func(Order) bool) {
		for page := 1; ; page++ {
			orders, more := api.FetchPage(page)

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

func (api *API) Requests() int {
	return api.requests
}

func main() {
	mode := flag.String("mode", "", "eager or iterator")
	total := flag.Int("orders", 5_000_000, "number of generated orders")
	pageSize := flag.Int("page-size", 10_000, "orders generated per page")
	flag.Parse()

	if *mode != "eager" && *mode != "iterator" {
		fmt.Fprintln(os.Stderr, "-mode must be eager or iterator")
		os.Exit(2)
	}
	if *total < 1 || *pageSize < 1 {
		fmt.Fprintln(os.Stderr, "-orders and -page-size must be positive")
		os.Exit(2)
	}

	runtime.GC()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	api := NewAPI(*total, *pageSize)
	var (
		checksum uint64
		retained []Order
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
	fmt.Printf("orders:        %d\n", *total)
	fmt.Printf("page size:     %d\n", *pageSize)
	fmt.Printf("pages fetched: %d\n", api.Requests())
	fmt.Printf("checksum:      %d\n", checksum)
	fmt.Printf("retained heap: %s\n", formatBytes(positiveDifference(after.HeapAlloc, before.HeapAlloc)))

	// Keep the eager slice live until after the post-GC measurement.
	runtime.KeepAlive(retained)
}

func runEager(api *API) (uint64, []Order) {
	orders := api.FetchAllOrders()
	return checksumSlice(orders), orders
}

func runIterator(api *API) uint64 {
	var checksum uint64
	for order := range api.Orders() {
		checksum += orderChecksum(order)
	}
	return checksum
}

func checksumSlice(orders []Order) uint64 {
	var checksum uint64
	for _, order := range orders {
		checksum += orderChecksum(order)
	}
	return checksum
}

func orderChecksum(order Order) uint64 {
	return uint64(order.ID)*31 + uint64(order.TotalP)
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
