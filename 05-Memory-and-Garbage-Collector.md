```
High Performance Go Workshop

1. Benchmarking
2. Performance measurement and profiling
3. Compiler optimisations
4. Execution Tracer
5. Memory and Garbage Collector

    5.1. Garbage collector world view
    5.2. Garbage collector design
    5.3. Minimise allocations
    5.4. Using sync.Pool
    5.5. Rearrange fields for better packing
    5.6. Exercises

6. Tips and trips

Dave Cheney dave@cheney.net (v379996b, 2019-07-24) 
```

# 5. Memory and Garbage Collector

> Go is a garbage collected language. This is a design principle, it will not change.

As a garbage collected language, the **performance of Go programs** is often determined by their **interaction with the garbage collector**.

Next to your choice of algorithms, **memory consumption** is the most **important factor** that determines the performance and scalability of your application.

This section discusses the operation of the garbage collector, how to measure the memory usage of your program and **strategies for lowering memory usage** `if garbage collector performance is a bottleneck`.

## 5.1 Garbage collector world view

Over time the Go GC has moved from a pure stop the world collector to a `concurrent, non compacting, collector`.

This is because the Go GC is designed for **low latency servers** and **interactive applications**.

## 5.2 Garbage collector design

The design of the Go GC has changed over the years:

- Go 1.0, **stop the world mark sweep collector** based heavily on tcmalloc.
- Go 1.3, fully precise collector, wouldn't mistake big numbers on the heap for pointers, thus leaking memory.
- Go 1.5, new GC design, focusing on **latency over throughput**.
- Go 1.6, GC improvements, **handling larger heaps with lower latency**.
- Go 1.7, small GC improvements, mainly refactoring.
- Go 1.8, further work to **reduce STW times**, now down to the **100 microsecond** range.
- Go 1.10+, **move away from pure cooprerative goroutine scheduling** to lower the latency when triggering a full GC cycle.
- Go 1.13 **Scavenger rewritten**.

----

## 5.2.1. Garbage collector tuning

The Go runtime provides one environment variable to tune the GC, [GOGC](https://dave.cheney.net/high-performance-go-workshop/gophercon-2019.html#memory-and-gc).

The formula for GOGC is:

```txt
goal = reachable * (1 + GOGC/100)
```

For example, if we currently have a 256MB heap, and `GOGC=100` (the default), when the heap fills up it will grow to:

```txt
goal = reachable * (1 + GOGC/100)
---------------------------------
goal = 256MB * (1 + 100/100)
goal = 256MB * 2
goal = 512MB
```

- Values of `GOGC` **greater than 100** causes the heap to grow faster, `reducing` the **pressure on the GC**.
- Values of `GOGC` **less than 100** cause the heap to grow slowly, `increasing` the **pressure on the GC**.

The default value of `100 is just_a_guide`.

You should `choose your own value after profiling your application with production loads`.

----

## 5.2.2. RSS and the scavenger

Many applications operate through distict phases:
- setup, 
- steady-state, 
- and (optionally) shutdown.

The phases have different memory profiles:

- **Setup** may process or summarise **large amounts of data**.
- **Steady-state** may consume memory **proportional to connected clients or request rate**.
- **Shutdown** may consume memory proportional to the amount of data processed during steady state to **summarise or pasivate data to disk**.

In practical terms your application may use more memory on startup than during the rest of it's life, then its **heap will be larger than necessary**, but mostly unused.

It would be useful if the **Go runtime** could **tell the operating system** [which parts of the, mostly unoccupied, heap are not needed](https://github.com/golang/proposal/blob/master/design/30333-smarter-scavenging.md#scavenging) (scavenging).

Scavenging is especially useful in dealing with page-level external fragmentation, since we can give these fragments back to the OS, reducing the process' **resident set size** (`RSS`). That is, the amount of memory that is backed by physical memory in the application’s address space.

> The Scavenging Process
> 
> The scavenger has remained mostly unchanged since it was first implemented in Go 1.1.
>
> As of Go 1.11, the only scavenging process in the Go runtime was a `periodic scavenger which runs every 2.5 minutes`. This scavenger combs over all the free spans in the heap and scavenge them if they have been `unused for at least 5 minutes`.
>
> As of Go 1.12, in addition to the periodic scavenger, the Go runtime also performs heap-growth scavenging.
> 
>In Go 1.13 scavenging moves to something demand driven, thus processes which do not benefit from scavenging do not pay for it whereas long running programs where memory allocation varys widely should return memory to the operating system more effectively.
>
> However, some of the CLs related to scavenging have not been committed. It is possible that this work will not be completed until Go 1.14.
>
> - [Design Document](https://github.com/golang/proposal/blob/master/design/30333-smarter-scavenging.md) Smarter Scavenging
> - [Proposal](https://github.com/golang/go/issues/30333) Smarter Scavenging

----

## 5.2.3. Garbage collector monitoring

A simple way to obtain a general idea of how hard the garbage collector is working is to enable the output of GC logging.

These **stats are always collected**, but normally suppressed, you can **enable their display** by setting the `GODEBUG` environment variable.

```go
GODEBUG=gctrace=1 godoc -http=:8080

gc 1 @0.012s 2%: 0.026+0.39+0.10 ms clock, 0.21+0.88/0.52/0+0.84 ms cpu, 4->4->0 MB, 5 MB goal, 8 P
gc 2 @0.016s 3%: 0.038+0.41+0.042 ms clock, 0.30+1.2/0.59/0+0.33 ms cpu, 4->4->1 MB, 5 MB goal, 8 P
gc 3 @0.020s 4%: 0.054+0.56+0.054 ms clock, 0.43+1.0/0.59/0+0.43 ms cpu, 4->4->1 MB, 5 MB goal, 8 P
gc 4 @0.025s 4%: 0.043+0.52+0.058 ms clock, 0.34+1.3/0.64/0+0.46 ms cpu, 4->4->1 MB, 5 MB goal, 8 P
gc 5 @0.029s 5%: 0.058+0.64+0.053 ms clock, 0.46+1.3/0.89/0+0.42 ms cpu, 4->4->1 MB, 5 MB goal, 8 P
gc 6 @0.034s 5%: 0.062+0.42+0.050 ms clock, 0.50+1.2/0.63/0+0.40 ms cpu, 4->4->1 MB, 5 MB goal, 8 P
gc 7 @0.038s 6%: 0.057+0.47+0.046 ms clock, 0.46+1.2/0.67/0+0.37 ms cpu, 4->4->1 MB, 5 MB goal, 8 P
gc 8 @0.041s 6%: 0.049+0.42+0.057 ms clock, 0.39+1.1/0.57/0+0.46 ms cpu, 4->4->1 MB, 5 MB goal, 8 P
gc 9 @0.045s 6%: 0.047+0.38+0.042 ms clock, 0.37+0.94/0.61/0+0.33 ms cpu, 4->4->1 MB, 5 MB goal, 8 P

Where the fields are as follows:
	'gc #'        the GC number, incremented at each GC
	'@#s'         time in seconds since program start
	'#%'          percentage of time spent in GC since program start
	'#+...+#'     wall-clock/CPU times for the phases of the GC
	'#->#-># MB'  heap size at GC start, at GC end, and live heap
	'# MB goal'   goal heap size
	'# P'         number of processors used

The phases of GC are:
(1) STW sweep termination => (2) concurrent mark and scan => (3) STW mark termination
    0.21+0.88/                   0.52/                           0+0.84 ms cpu

// STW => stop-the-world
```
The trace output gives a general measure of GC activity. The [output format of gctrace=1](https://golang.org/pkg/runtime/#hdr-Environment_Variables) is described in the `runtime` package documentation.

> Use the `GODEBUG` in **production**, it has **no performance impact**. 

Using `GODEBUG=gctrace=1` is good **when you know there is a problem**
- but for **general telemetry** on your Go application I recommend the `net/http/pprof` interface.

```go
import _ "net/http/pprof"
```

Importing the `net/http/pprof` package will register a handler at `/debug/pprof` with **various runtime metrics**, including:

- A list of all the **running goroutines**, `/debug/pprof/heap?debug=1`
- A report on the **memory allocation statistics**, `/debug/pprof/heap?debug=1`

> Be careful as these endpoints will be visible if you use `http.ListenAndServe(address, nil)`.

----

## 5.3 Minimise allocations

Memory allocation is not free, this is true reguardless of if your language is garbage collected or not.

Memory allocation can be an overhead spread throughout your codebase; each represents a tiny fraction of the total runtime, but collectively they represent a sizeable cost.

Because that cost is spread across many places, identifying the biggest offenders can be complicated and often requires reworking APIs.

**Each allocation should pay its way.**

> Analogy: if you move to a larger house because you’re planning on having a family, that’s a good use of your capital. If you move to a larger house because someone asked you to mind their child for an afternoon, that’s a poor use of your capital.


## 5.3.1 strings vs []bytes

In Go `string` values are **immutable**, `[]byte` are **mutable**.

Most programs prefer to work with `string`, but **most IO is done** with `[]byte`.

**Avoid `[]byte` to `string` conversions** wherever possible, this normally means picking one representation, either a `string` or a `[]byte` for a value.

Often this will be `[]byte` if you read the data from the **network or disk**.

The [bytes](https://golang.org/pkg/bytes/) package contains many of the same operations — `Split`, `Compare`, `HasPrefix`, `Trim`, etc — as the [strings](https://golang.org/pkg/strings/) package.

Under the hood `strings` uses **same assembly primitives** as the `bytes` package.

---

## 5.3.2 Using []byte as a map key

It is very common to use a `string` as a map key, but often you have a `[]byte`.

Alterntative you may have a key in `[]byte` form, but slices do not have a defined equality operator so **cannot be used as map keys**.

The compiler implements a specific optimisation for this case:

```go
var bytes []byte{'F', 'r', 'a', 'n', 'c', 'e'}

var m map[string]string
v, ok := m[string(bytes)] // compiler specific optimisation
```

**This will avoid the conversion** of the byte slice to a string for the map lookup.

This is very specific, it won't work if you do something like:

```go
key := string(bytes)
val, ok := m[key] // No compiler optimization
```

Write a benchmark comparing these two methods of using a `[]byte` as a `string` map key.

```sh
go test -run=^$ -bench=. -benchmem ./examples/benchmap/

BenchmarkMapLookup-12     	64933486	        18.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkMapLookup2-12    	54124168	        21.8 ns/op	       0 B/op	       0 allocs/op
```

----

## 5.3.3 []byte to string conversions

Just like `[]byte` to `string` conversions are necessary for map keys, comparing two `[]byte` slices for equality either

- requires a conversion — potentially a copy — to a `string`
- or the use of the `bytes.Equal` function.

The good news is in 1.13 the compiler has improved to the point that `[]byte` to `string` conversions for the purpose of **equality testing avoids** the allocation.

```go
func BenchmarkBytesEqualInline(b *testing.B) {
	x := bytes.Repeat([]byte{'a'}, 1<<20)
	y := bytes.Repeat([]byte{'a'}, 1<<20)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if string(x) != string(y) { // inline []byte to string conversions (avoids allocs)
			b.Fatal("x != y")
		}
	}
}

func BenchmarkBytesEqualExplicit(b *testing.B) {
	x := bytes.Repeat([]byte{'a'}, 1<<20)
	y := bytes.Repeat([]byte{'a'}, 1<<20)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		q := string(x)
		r := string(y)
		if q != r {
			b.Fatal("x != y")
		}
	}
}
```

```sh
go test -run=^$ -bench=. -benchmem ./examples/byteseq

BenchmarkBytesEqualInline-12      	   31782	     37031 ns/op	       0 B/op	       0 allocs/op
BenchmarkBytesEqualExplicit-12    	    4924	    241077 ns/op	 2097155 B/op	       2 allocs/op
```

However, **the copy is only elided in simple cases**.

----

## 5.3.4 Avoid string concatenation

Go strings are immutable. Concatenating two strings generates a third.

Which of the following is fastest?

```go
// Concatenating strings
s := request.ID
s += " " + client.Addr().String()
s += " " + time.Now().String()
r = s
```

```go
// fmt.Fprintf
var b bytes.Buffer
fmt.Fprintf(&b, "%s %v %v", request.ID, client.Addr(), time.Now())
r = b.String()
```

```go
// fmt.Sprintf
r = fmt.Sprintf("%s %v %v", request.ID, client.Addr(), time.Now())
```

```go
// []byte to string conversion
b := make([]byte, 0, 40)
b = append(b, request.ID...)
b = append(b, ' ')
b = append(b, client.Addr().String()...)
b = append(b, ' ')
b = time.Now().AppendFormat(b, "2006-01-02 15:04:05.999999999 -0700 MST")
r = string(b)
```

```go
// strings.Builder
var b strings.Builder
b.WriteString(request.ID)
b.WriteString(" ")
b.WriteString(client.Addr().String())
b.WriteString(" ")
b.WriteString(time.Now().String())
r = b.String()
```

```go
// go test -run=^$ -bench=. -benchmem ./examples/concat/
BenchmarkConcatenate-12       	 1562970	       754 ns/op	     272 B/op	      10 allocs/op
BenchmarkFprintf-12           	 1000000	      1272 ns/op	     432 B/op	      13 allocs/op
BenchmarkSprintf-12           	 1000000	      1090 ns/op	     304 B/op	      11 allocs/op
BenchmarkStrconv-12           	 2278242	       540 ns/op	     165 B/op	       5 allocs/op
BenchmarkStringsBuilder-12    	 1445193	       813 ns/op	     280 B/op	      11 allocs/op
```

----

## 5.3.5 Don't force allocations on the callers of your API

Make sure your APIs allow the caller to reduce the amount of garbage generated.

Consider these two Read methods:

```go
func (r *Reader) Read() ([]byte, error)

func (r *Reader) Read(buf []byte) (int, error)
```

The first Read method takes no arguments and returns some data as a `[]byte`.

 - This Read method `will always allocate a buffer`, putting pressure on the GC.

The second Read method takes a `[]byte buffer` and returns the amount of bytes read.
 - This Read method `fills the buffer it was given`.

----

## 5.3.6 Preallocate slices if the length is known

Append is convenient, but wasteful.

Slices grow by doubling up to 1024 elements, then by approximately 25% after that.

What is the capacity of b after we append one more item to it?

```go
func main() {
	b := make([]int, 1024)
	b = append(b, 99)
	fmt.Println("len:", len(b), "cap:", cap(b))
}
// len: 1025 cap: 1280 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 99]

// If you use the append pattern you could be copying a lot of data and creating a lot of garbage.
```

If you use the append pattern you could be copying a lot of data and creating a lot of garbage.

If know know the length of the slice beforehand, then pre-allocate the target to avoid copying and to make sure the target is exactly the right size.

```go
// Before:
var s []string
for _, v := range fn() {
        s = append(s, v)
}
return s

// After:
vals := fn()
s := make([]string, len(vals))
for i, v := range vals {
        s[i] = v
}
return s
```

## 5.4 Using sync.Pool

The sync package comes with a `sync.Pool` type which is used to **reuse** common objects.

> sync.Pool is not a cache.
> 
> It can and will be emptied at_any_time. Do not place important items in a sync.Pool, they will be discarded. 

`sync.Pool` has no fixed size or maximum capacity. You add to it and take from it until a GC happens, then it is emptied unconditionally. This is by design:

> If before garbage collection is too early and after garbage collection too late, then the right time to drain the pool must be during garbage collection. That is, the semantics of the Pool type must be that it drains at each garbage collection. — Russ Cox

```go
// examples/pool/pool.go

var pool = sync.Pool{New: func() interface{} { return make([]byte, 4096) }}

func fn() {
	buf := pool.Get().([]byte) // takes from pool or calls New
	// do work
	pool.Put(buf) // returns buf to the pool
}
```


## 5.5 Rearrange fields for better packing

Consider this struct declaration

```go
type S struct {
    a bool
	b float64
    c int32
}
```

How many **bytes of memory** does a value of this type consume?

```go
var s S
fmt.Println(unsafe.Sizeof(s)) 
// 24 bytes on 64-bit platforms, 16 on 32-bit platforms.
```

Why? The answer has to do with **padding and alignment**.

On platforms that do support so called **unaligned access**, there is usually a **cost to access** these fields.

> Even on platforms that allow unaligned access, `sync/atomic` requires the values be naturally aligned.
> 
> This is because atomic operations are implemented in the various **L1, L2, L3 caching layers**, which always work in amounts known as `cache lines` (normally 32-64 bytes wide). `Atomic access cannot span cache lines`, so they must be correctly aligned. 

```go
// The CPU expects fields which are 4 bytes long
// to be alligned on 4 byte boundaries (4*0, 4*1, 4*2, ...),
// 8 byte values on 8 byte boundaries (8*0, 8*1, 8*2, ...), 
// and so on.

type S struct {
    a bool      // 1 byte
                // 7 bytes padding 
	b float64   // 8 bytes wide on all platforms
    c int32     // 4 bytes
                // 4 bytes padding
                // 24 bytes
}

// We can infer how the compiler is going to lay out these fields in memory:
type S struct {
	a bool    // [0] ==> start
	_ [7]byte // padding 
	b float64 // [8] (8*1)
	c int32   // [16] (4*4) ==> start + 4 bytes
    _ [4]byte // padding 
              // [24] <== end index ([16]+4+4)
}

// 7 bytes of padding is required to ensure b float64 starts on an 8 byte boundary.
// 4 bytes of padding are required to ensure that arrays (or slices) of S's are correctly aligned in memory.
```

> Exercise: rearrange the fields in S to reduce its overall size.

```go
type S struct {
	b float64   // 8 bytes wide on all platforms
    c int32     // 4 bytes
    a bool      // 1 byte
                // 13 bytes
}
```

## 5.6 Exercises

- Using godoc (or another program) observe the results of changing `GOGC` using `GODEBUG=gctrace=1`
- Benchmark byte’s string(byte) map keys
- Benchmark allocs from different concat strategies

----

(Execution Tracer) [prev](04-Execution-Tracer.md) | [next]()