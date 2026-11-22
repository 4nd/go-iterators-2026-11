---
marp: true
theme: default
paginate: true
size: 16:9
title: Go iterators
author: Andy Graham
description: A short introduction to Go iterators using Fibonacci numbers
style: |
  section {
    background: #f7f8f5;
    color: #17212b;
    font-family: "Avenir Next", Avenir, Helvetica, Arial, sans-serif;
    font-size: 29px;
    padding: 58px 72px;
  }

  section.lead {
    background: #102a43;
    color: #f7f8f5;
  }

  section.lead h1 {
    color: #63d2b3;
    font-size: 64px;
  }

  section::after {
    color: #627381;
    font-size: 16px;
  }

  section.lead::after {
    color: #b8c7d5;
  }

  h1 {
    color: #006d77;
    font-size: 43px;
    margin: 0 0 28px;
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
    font-size: 23px;
    line-height: 1.32;
    padding: 20px 26px;
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

  section :is(pre, marp-pre) .hljs-string {
    color: #e0baff;
  }

  table {
    font-size: 25px;
  }

  th {
    background: #dcefe9;
    color: #173f45;
  }

  th, td {
    padding: 14px 18px;
  }
---

<!-- _class: lead -->

# Go iterators

Andy Graham · DailyPay

<!--
We use range all the time. Today I want to show what happens when the thing on the right of range is a function.
-->

---

# What can `range` loop over?

~~~go
for i := range 10 {
    fmt.Println(i)
}

s := []string{"Hello", "World"}
for i, v := range s {
    fmt.Println(i, v)
}
~~~

Each iteration gives the loop a value to work with.

<!--
The first loop prints 0 to 9. The second gets an index and a string each time. Both are in the demo. Now imagine the values arrive from a function that generates them one by one.
-->

---

# Range over Fibonacci numbers

~~~go
for n := range fibonacci() {
    if n > 100 {
        break
    }
    fmt.Println(n)
}
~~~

What’s the last number this loop prints?

<!--
Run: go run ./numbers

Point out the integer and slice output first. The Fibonacci output ends at 89. The iterator offers 144 next, and the caller breaks before printing it. The caller decides how many numbers it wants. Next, open fibonacci.
-->

---

# How the iterator produces a value

~~~go
func fibonacci() iter.Seq[int] {
    return func(yield func(int) bool) {
        current, next := 0, 1
        for {
            if !yield(current) {
                return
            }
            current, next = next, current+next
        }
    }
}
~~~

`iter.Seq[int]` is a function type for yielding one integer at a time.

<!--
Point to current and next, then show how the last line advances them. Pause at yield(current) and ask where yield comes from.
-->

---

# What `range` does with `iter.Seq`

~~~go
type Seq[V any] func(yield func(V) bool)
~~~

Go supplies `yield`. Calling `yield(n)` runs the loop body once.

The Fibonacci loop with the callback written out

~~~go
sequence := fibonacci()
sequence(func(n int) bool {
    if n > 100 {
        return false
    }
    fmt.Println(n)
    return true
})
~~~

<!--
Start with the inside type, func(V) bool. It takes one value and tells the iterator whether to keep going. In this example V is int. The outer function is the iterator itself. It takes that callback as its argument.

The callback checks the limit before printing, just like the range loop. Ask what it returns at 89, then what happens at 144.
-->

---

# How `break` stops the iterator

| Value offered | What the loop does | `yield` returns |
|---|---|---|
| `89` | Prints `89` and finishes the iteration | `true` |
| `144` | Hits `break` before printing | `false` |

~~~go
if !yield(current) {
    return
}
~~~

<!--
At 89, the loop body finishes and the iterator advances. At 144, break makes yield return false. The iterator returns without yielding another number. Returning from the enclosing function also stops the loop.
-->

---

# A standard-library iterator · `slices.Backward`

~~~go
s := []string{"Hello", "World"}
for i, v := range slices.Backward(s) {
    fmt.Println(i, v)
}
~~~

~~~text
1 World
0 Hello
~~~

`iter.Seq2[int, string]` yields an index and a value on each iteration.

- `maps.All(m)` yields key-value pairs.
- `strings.SplitSeq(s, ",")` yields one substring at a time.

<!--
This is the same slice and the same loop body as the opening example. Only the expression after range changes. Backward visits the original slice from the last index to the first. Seq2 works like Seq, with two arguments to yield instead of one.

[Go documentation for slices.Backward](https://pkg.go.dev/slices#Backward)
[Go documentation for maps.All](https://pkg.go.dev/maps#All)
[Go documentation for strings.SplitSeq](https://pkg.go.dev/strings#SplitSeq)
-->

---

# When an iterator is useful

- **Use a slice** when you already have the values and want to index or reuse them.
- **Use an iterator** when you can produce values one at a time and the caller may stop early.

The iterator produces the values. The loop decides what to do with them and when to stop.

<!--
A tree traversal is another example. The iterator decides how to visit the nodes. The loop can stop at the first match. If you already have a slice that suits the job, just range over it.
-->
