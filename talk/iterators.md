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

## Finding the latest failed deployment

Go 1.23+

<!--
Ask how many pages we need to fetch before we can report the latest failure.
-->

---

# Find the latest failed deployment

The API lists deployments newest first.

| Page | Deployments |
|---|---|
| 1 | deploy-106 succeeded, deploy-105 succeeded |
| 2 | **deploy-104 failed**, deploy-103 succeeded |
| 3 | deploy-102 failed, deploy-101 succeeded |

How many pages do we need?

<!--
There are two failures. The first one we encounter is the latest, because
the API lists deployments newest first. Page 2 gives us the answer.
-->

---

<!-- _class: tight -->

# Stage 1: manual pagination

~~~go
found := false
for page := 1; !found; page++ {
    deployments, more := api.FetchPage(page)

    for _, deployment := range deployments {
        if deployment.Status == mockapi.Failed {
            found = true
            break
        }
    }

    if !more {
        break
    }
}
~~~

**Efficient:** two requests

**Caller complexity:** page numbers, continuation state, nested loops

<!--
Run: go run ./stage1

The caller stops after two requests, but the business question is buried inside pagination control.

The found flag exists only to carry a break from the inner loop to the outer loop.
-->

---

<!-- _class: two-code -->

# Stage 2: eager repository

<div class="columns compact">

<div>

~~~go
deploymentRepo := repo.New(api)
deployments := deploymentRepo.Deployments()

for _, deployment := range deployments {
    if deployment.Status == mockapi.Failed {
        break
    }
}
~~~

</div>

<div>

~~~go
func (r *Repository) Deployments() []mockapi.Deployment {
    var all []mockapi.Deployment

    for page := 1; ; page++ {
        deployments, more := r.api.FetchPage(page)
        all = append(all, deployments...)

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

Moving pagination into a repository gives us the caller we want, but Deployments must finish before the caller can inspect anything.
-->

---

# All three pages arrive before the loop starts

~~~text
GET /deployments?page=1
GET /deployments?page=2
GET /deployments?page=3
inspecting deploy-106 (succeeded)
inspecting deploy-105 (succeeded)
inspecting deploy-104 (failed)
latest failed deployment: deploy-104
API requests: 3
deployments materialised: 6
~~~

The repository builds the complete slice before the caller can inspect the first deployment.

<!--
Point at the three GET lines. The slice already exists when the caller breaks.
-->

---

# Stage 3: range over deployments

~~~go
for deployment := range deploymentRepo.Deployments() {
    if deployment.Status == mockapi.Failed {
        break
    }
}
~~~

The stopping condition stays the same. How many pages will this fetch?

<!--
This is the payoff slide.

Run: go run ./stage3

The two loops have the same stopping condition. Ask how many GET lines this one will print.
-->

---

<!-- _class: two-code -->

# The iterator fetches page 2 when the loop needs it

<div class="columns">

<div>

### Slice

~~~text
GET /deployments?page=1
GET /deployments?page=2
GET /deployments?page=3
inspecting deploy-106 (succeeded)
inspecting deploy-105 (succeeded)
inspecting deploy-104 (failed)
latest failed deployment: deploy-104
API requests: 3
~~~

</div>

<div>

### Iterator

~~~text
GET /deployments?page=1
inspecting deploy-106 (succeeded)
inspecting deploy-105 (succeeded)
GET /deployments?page=2
inspecting deploy-104 (failed)
latest failed deployment: deploy-104
API requests: 2
~~~

</div>

</div>

<!--
Page 1 has no match. The loop keeps going, so the repository fetches page 2.
The caller breaks on deploy-104. The repository never asks for page 3.
-->

---

# One call to <code>yield</code> runs one loop iteration

~~~go
if !yield(deployment) {
    return
}
~~~

| Deployment | What the loop does | What <code>yield</code> returns |
|---|---|---|
| deploy-105 succeeded | Finishes the iteration | <code>true</code> |
| deploy-104 failed | Hits <code>break</code> | <code>false</code> |

~~~go
type Seq[V any] func(yield func(V) bool)
~~~

<!--
Go supplies yield when range calls the iterator. The producer calls yield
with a deployment. The boolean tells the producer whether the loop is still
running. Once yield returns false, the producer must stop calling it.
-->

---

<!-- _class: tight -->

# Stage 3: lazy repository

~~~go
func (r *Repository) Deployments() iter.Seq[mockapi.Deployment] {
    return func(yield func(mockapi.Deployment) bool) {
        for page := 1; ; page++ {
            deployments, more := r.api.FetchPage(page)

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
~~~

The repository returns as soon as <code>yield</code> returns <code>false</code>.

<!--
Read from the inside out. The repository offers each deployment to the caller. If the caller has broken out of the loop, yield returns false and the repository returns immediately.
-->

---

# The iterator keeps pagination in the repository

| Version | Caller handles pages | API requests |
|---|:---:|:---:|
| Manual pagination | Yes | **2** |
| Eager repository | No | 3 |
| Iterator repository | No | **2** |

All three find <code>deploy-104</code>. With the iterator, the caller's <code>break</code> also stops the repository.

<!--
The request counts come from the three runnable demos.
-->

---

<!-- _class: lead -->

# The caller stops at the latest failure

## Keep pagination in the producer

~~~go
for deployment := range deploymentRepo.Deployments() {
    if deployment.Status == mockapi.Failed {
        break
    }
}
~~~

<!--
Close on the caller code.

The practical definition for this talk: an iterator lets the caller consume values one at a time while the producer retains control of how those values are found.
-->

---

# Choosing between a slice and an iterator

| Use a slice when… | Use an iterator when… |
|---|---|
| The values already exist in memory | Values arrive from pages or a stream |
| Most callers need every value | Callers often stop early |
| The simplest API is the priority | Avoiding unnecessary work matters |

<!--
An iterator earns its complexity when values are produced incrementally and early termination can avoid meaningful work.
-->

---

<!-- _class: compact -->

# Appendix: iterator benefits

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

- Start processing when the first value arrives.
- Stop producing values when the caller breaks.
- Keep pagination inside the repository.
- Process values without collecting a full slice.

</div>

</div>

<span class="muted">The producer implementation determines which benefits apply.</span>

<!--
Iterators enable these benefits, but they do not guarantee all of them. The implementation still decides whether values are produced lazily.

The main demo will make early stopping visible. The full-scan example measures retained heap.
-->

---

<!-- _class: two-code -->

# Appendix: full-scan memory

<div class="columns">

<div>

~~~text
mode:          eager
deployments:   5,000,000
pages fetched: 500
checksum:      387525420000000
retained heap: 76.3 MiB
~~~

</div>

<div>

~~~text
mode:          iterator
deployments:   5,000,000
pages fetched: 500
checksum:      387525420000000
retained heap: a few KiB
~~~

</div>

</div>

Both modes scanned every deployment and produced the same result.

<span class="muted">Live heap after GC, with the eager slice held in memory.</span>

<!--
Run these as separate commands:

go run ./memorydemo -mode=eager
go run ./memorydemo -mode=iterator

The eager implementation preallocates its result slice, so this is already a favourable case for it. The iterator fetches the same 500 pages but does not retain a collection after the scan.
-->
