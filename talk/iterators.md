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

  section :is(pre, marp-pre) .hljs-title {
    color: #63d2b3;
  }

  section :is(pre, marp-pre) :is(.hljs-keyword, .hljs-type) {
    color: #ff9b91;
  }

  section :is(pre, marp-pre) :is(.hljs-literal, .hljs-number) {
    color: #9ccaff;
  }

  section :is(pre, marp-pre) .hljs-built_in {
    color: #ffd18a;
  }

  section :is(pre, marp-pre) .hljs-string {
    color: #e0baff;
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

  section.repository-trace .columns {
    grid-template-columns: minmax(0, 5fr) minmax(0, 3fr);
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

Andy Graham · DailyPay

---

# Clean code and optimisation

**Clean code** makes behaviour easier to understand, review and change.

- Make common operations easy to reuse.
- Keep implementation details behind a clear interface.
- Keep control flow easy to follow.

**Optimisation** starts with finding where time, memory or requests are being spent.

- Measure before making a change.
- Avoid work the result doesn’t need.
- Measure again to verify the improvement.

> Can we keep a function easy to read and still optimised?

<!--
Readability matters every time someone comes back to change the code. Performance matters when the code is doing work we can avoid. Sometimes the shape of the code affects both. To see how, let's look at a small example and count the work each version does.
-->

---

# Example: Find the latest failed deployment

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
            fmt.Printf("  latest failed deployment: %s\n", deployment.ID)
            found = true
            break
        }
    }

    if !more {
        break
    }
}
~~~

<!--
Run: go run ./stage1

The search runs directly in main and stops after two requests. It tracks page numbers and continuation state in addition to checking each deployment.

The found flag exists only to carry a break from the inner loop to the outer loop.
-->

---

<!-- _class: tight -->

# Stage 2: eager repository

~~~go
func findLatestFailedDeployment(deploymentRepo *repo.Repository) (mockapi.Deployment, bool, int) {
    deployments := deploymentRepo.Deployments()

    for _, deployment := range deployments {
        if deployment.Status == mockapi.Failed {
            return deployment, true, len(deployments)
        }
    }

    return mockapi.Deployment{}, false, len(deployments)
}
~~~

<!--
Run: go run ./stage2

The search receives the repository, calls Deployments and loops over a slice.
-->

---

<!-- _class: two-code repository-trace -->

# All three pages arrive before the loop starts

<div class="columns">

<div>

~~~go
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
~~~

</div>

<div>

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

</div>

</div>

The repository builds the complete slice before the caller can inspect the first deployment.

<!--
Point to the append in Deployments, then to the three GET lines. The repository has built the slice by the time the search returns.
-->

---

<!-- _class: tight -->

# Stage 3: range over deployments

~~~go
func findLatestFailedDeployment(deploymentRepo *repo.Repository) (mockapi.Deployment, bool) {
    for deployment := range deploymentRepo.Deployments() {
        if deployment.Status == mockapi.Failed {
            return deployment, true
        }
    }

    return mockapi.Deployment{}, false
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
The search returns on deploy-104. The repository never asks for page 3.
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
| deploy-104 failed | Returns from the search function | <code>false</code> |

~~~go
type Seq[V any] func(yield func(V) bool)
~~~

Leaving the loop early with <code>break</code> or <code>return</code> makes <code>yield</code> return <code>false</code>.

<!--
Go supplies yield when range calls the iterator. The producer calls yield
with a deployment. The boolean tells the producer whether the loop is still
running. A break or return exits the loop. A continue asks for the next value.
Once yield returns false, the producer must stop calling it.
-->

---

<!-- _class: tight -->

# Stage 3: lazy repository

~~~go
func (repository *Repository) Deployments() iter.Seq[mockapi.Deployment] {
    return func(yield func(mockapi.Deployment) bool) {
        for page := 1; ; page++ {
            deployments, more := repository.api.FetchPage(page)

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
Read from the inside out. The repository offers each deployment to the caller. If the caller exits the loop early, yield returns false and the repository returns immediately.
-->

---

# The iterator keeps pagination in the repository

| Version | Caller handles pages | API requests |
|---|:---:|:---:|
| Manual pagination | Yes | **2** |
| Eager repository | No | 3 |
| Iterator repository | No | **2** |

All three find <code>deploy-104</code>. With the iterator, returning from the search also stops the repository.

<!--
The request counts come from the three runnable demos.
-->

---

# Choosing between a slice and an iterator

| Use a slice when… | Use an iterator when… |
|---|---|
| The values already exist in memory | Values arrive from pages or a stream |
| Most callers need every value | Callers often stop early |
| The simplest API is the priority | Avoiding unnecessary work matters |

- **Readability.** The search loop shows what qualifies as a match and when to stop.
- **Reuse.** Other searches can use the same repository method to traverse deployments.

<!--
Both repository versions keep pagination in one place. An iterator earns its complexity when values are produced incrementally and early termination can avoid meaningful work.
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
- Stop producing values when the caller exits the loop early.
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
