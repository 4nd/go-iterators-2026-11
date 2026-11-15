---
marp: true
theme: default
paginate: true
size: 16:9
title: Go iterators
description: A practical introduction using a paginated API
style: |
  section {
    background: #f7f8f5;
    color: #17212b;
    font-family: "Avenir Next", Avenir, Helvetica, Arial, sans-serif;
    font-size: 29px;
    padding: 58px 72px;
  }

  section::after {
    color: #6c7a86;
    font-size: 16px;
  }

  section.lead {
    background: #102a43;
    color: #f7f8f5;
    text-align: left;
  }

  section.lead h1 {
    color: #63d2b3;
    font-size: 64px;
    margin-bottom: 18px;
  }

  section.lead h2 {
    color: #d9e7ef;
    font-size: 32px;
    font-weight: 500;
  }

  h1 {
    color: #006d77;
    font-size: 43px;
    margin: 0 0 28px;
  }

  h2 {
    color: #334e68;
    font-size: 28px;
  }

  strong {
    color: #006d77;
  }

  code {
    font-family: Menlo, Monaco, "Courier New", monospace;
  }

  pre {
    background: #102a43;
    border-radius: 8px;
    color: #f5f7fa;
    font-size: 20px;
    line-height: 1.36;
    padding: 22px 26px;
  }

  pre code {
    color: inherit;
  }

  table {
    border-collapse: collapse;
    font-size: 24px;
    width: 100%;
  }

  th {
    background: #dcefe9;
    color: #173f45;
  }

  th, td {
    border: 1px solid #b8c8cf;
    padding: 14px 18px;
  }

  blockquote {
    border-left: 6px solid #f4a261;
    color: #334e68;
    font-size: 31px;
    margin: 30px 0;
    padding: 4px 0 4px 24px;
  }

  .columns {
    align-items: start;
    display: grid;
    gap: 42px;
    grid-template-columns: 1fr 1fr;
  }

  .compact pre {
    font-size: 18px;
  }

  section.tight pre {
    font-size: 18px;
    line-height: 1.24;
    padding: 16px 22px;
  }

  section.tight p {
    margin: 10px 0;
  }

  section.two-code pre {
    font-size: 16px;
    line-height: 1.25;
    padding: 18px;
  }

  .trace {
    font-family: Menlo, Monaco, "Courier New", monospace;
    font-size: 26px;
    line-height: 1.65;
  }

  .result {
    color: #006d77;
    font-size: 34px;
    font-weight: 700;
    margin-top: 24px;
  }

  .muted {
    color: #607786;
  }
---

<!-- _class: lead -->

# Go iterators

## Pagination without leaking pagination

Go 1.23+

<!--
Open with the problem, not the iterator type.

We have a paginated API and want the first matching order. The useful question is whether the producer can stop when the caller stops.
-->

---

<!-- _class: compact -->

# Iterator benefits

<div class="columns">

<div>

~~~go
for value := range values() {
    use(value)

    if done(value) {
        break
    }
}
~~~

</div>

<div>

**Avoid a full result slice**  
Values can arrive as needed.

**Begin processing sooner**  
The caller can start before the producer finishes.

**Stop unnecessary work**  
A <code>break</code> can stop future production.

**Hide traversal details**  
The producer owns paging or streaming.

</div>

</div>

<span class="muted">The producer implementation determines which benefits apply.</span>

<!--
This is the high-level map for the talk.

Iterators enable these benefits, but they do not guarantee all of them. The implementation still decides whether values are produced lazily.

The main demo will make early stopping visible. The appendix measures the retained-memory benefit.
-->

---

# The job

<div class="columns">

<div>

Find the first order worth more than **£50**.

Stop as soon as it appears.

> The first match is on page 1.

</div>

<div>

~~~text
page 1   £12.50
         £62.00  ← match
         £9.00

page 2   £45.00, £21.00, £75.00
page 3   £32.50, £11.00
~~~

</div>

</div>

<div class="result">The ideal result is one API request.</div>

<!--
The mock API has three pages and eight orders.

The second order on the first page is the match. Any request after page 1 is wasted for this caller.
-->

---

<!-- _class: tight -->

# Stage 1: manual pagination

~~~go
found := false
for page := 1; !found; page++ {
    orders, more := api.FetchPage(page)

    for _, order := range orders {
        if order.TotalP > 5_000 {
            found = true
            break
        }
    }

    if !more {
        break
    }
}
~~~

**Efficient:** one request

**Caller complexity:** page numbers, continuation state, nested loops

<!--
Run: go run ./stage1

The caller stops after one request, but the business question is buried inside pagination control.

The found flag exists only to carry a break from the inner loop to the outer loop.
-->

---

<!-- _class: two-code -->

# Stage 2: eager repository

<div class="columns compact">

<div>

~~~go
orderRepo := repo.New(api)
orders := orderRepo.Orders()

for _, order := range orders {
    if order.TotalP > 5_000 {
        break
    }
}
~~~

</div>

<div>

~~~go
func (r *Repository) Orders() []mockapi.Order {
    var all []mockapi.Order

    for page := 1; ; page++ {
        orders, more := r.api.FetchPage(page)
        all = append(all, orders...)

        if !more {
            return all
        }
    }
}
~~~

</div>

</div>

<!--
Run: go run ./stage2

Moving pagination into a repository gives us the caller we want, but Orders must finish before the caller can inspect anything.
-->

---

# The eager repository cannot see <code>break</code>

~~~text
GET /orders?page=1
GET /orders?page=2
GET /orders?page=3

inspecting A-100 (£12.50)
inspecting A-101 (£62.00)
found one — stop

API requests: 3
orders materialised: 8
~~~

The caller stops. The producer has already finished.

<!--
Pause on the order of the output. All three GETs happen before the first inspecting line.

The repository hides pagination, but its eager slice removes the caller's control over when fetching stops.
-->

---

# The API we want

~~~go
for order := range orderRepo.Orders() {
    fmt.Printf("inspecting %s\n", order.ID)

    if order.TotalP > 5_000 {
        break
    }
}
~~~

The caller sees **orders**, not pages.

The producer should see the caller's **break**.

<!--
This is the payoff slide.

The loop looks like a range over a slice. The difference is that values arrive only as the loop asks for them.
-->

---

# <code>iter.Seq</code> is a function

~~~go
type Seq[V any] func(yield func(V) bool)
~~~

<div class="columns">

<div>

### The producer

Calls <code>yield(value)</code> for each value it finds.

</div>

<div>

### The caller

Returns <code>true</code> while ranging.

A <code>break</code> makes <code>yield</code> return <code>false</code>.

</div>

</div>

<!--
Keep this explanation mechanical.

Seq is a function. Go supplies the yield function when range starts consuming it.

The bool is the cancellation signal. True means keep going. False means the loop stopped.
-->

---

<!-- _class: tight -->

# Stage 3: lazy repository

~~~go
func (r *Repository) Orders() iter.Seq[mockapi.Order] {
    return func(yield func(mockapi.Order) bool) {
        for page := 1; ; page++ {
            orders, more := r.api.FetchPage(page)

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
~~~

**Pagination:** repository &nbsp;&nbsp; **Stopping:** caller

<!--
Run: go run ./stage3

Read from the inside out. The repository offers each order to the caller. If the caller has broken out of the loop, yield returns false and the repository returns immediately.
-->

---

# <code>break</code> reaches the producer

<div class="trace">

1. Fetch page 1

2. Yield A-100 &nbsp; <span class="muted">caller continues</span>

3. Yield A-101 &nbsp; <span class="muted">caller breaks</span>

4. <code>yield</code> returns <code>false</code>

5. Producer returns before page 2

</div>

<div class="result">API requests: 1</div>

<!--
The important property is control flow, not syntax.

Break stops the consumer loop and also tells the producer to stop generating values.
-->

---

# The iterator keeps both benefits

| Version | Caller handles pages | API requests | Stores all orders |
|---|:---:|:---:|:---:|
| Manual pagination | Yes | **1** | No |
| Eager repository | No | 3 | Yes |
| Iterator repository | No | **1** | No |

The iterator keeps the clean caller without doing work the caller does not need.

<!--
These are the observed results from the three runnable demos.

The iterator does not make every loop faster. It matters here because producing the next value requires another external request.
-->

---

# Choosing between a slice and an iterator

| Use a slice when… | Use an iterator when… |
|---|---|
| The values already exist in memory | Values arrive from pages or a stream |
| Most callers need every value | Callers often stop early |
| The simplest API is the priority | Avoiding unnecessary work matters |

<!--
Do not replace slices everywhere.

An iterator earns its complexity when values are produced incrementally and early termination can avoid meaningful work.
-->

---

<!-- _class: lead -->

# Let <code>range</code> describe the job

## Keep pagination in the producer

~~~go
for order := range orderRepo.Orders() {
    if order.TotalP > 5_000 {
        break
    }
}
~~~

<!--
Close on the caller code.

The practical definition for this talk: an iterator lets the caller consume values one at a time while the producer retains control of how those values are found.
-->

---

<!-- _class: two-code -->

# Appendix: full-scan memory

<div class="columns">

<div>

~~~text
mode:          eager
orders:        5,000,000
pages fetched: 500
checksum:      387525420000000
retained heap: 76.3 MiB
~~~

</div>

<div>

~~~text
mode:          iterator
orders:        5,000,000
pages fetched: 500
checksum:      387525420000000
retained heap: 11 KiB
~~~

</div>

</div>

Both modes scanned every order and produced the same result.

<span class="muted">Metric: live heap retained after GC, measured in separate processes.</span>

<!--
Run these as separate commands:

go run ./memorydemo -mode=eager
go run ./memorydemo -mode=iterator

The eager implementation preallocates its result slice, so this is already a favourable case for it. The iterator fetches the same 500 pages but does not retain a collection after the scan.

The iterator result varied between 10.8 KiB and 11.6 KiB across two runs, so the slide rounds it to 11 KiB.
-->
