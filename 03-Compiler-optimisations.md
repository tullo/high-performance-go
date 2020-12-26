
```
High Performance Go Workshop

1. Benchmarking
2. Performance measurement and profiling
3. Compiler optimisations

    3.1. History of the Go compiler
    3.2. Escape analysis
    3.3. Inlining
    3.4. Dead code elimination
    3.5. Prove pass
    3.6. Compiler intrinsics
    3.7. Bounds check elimination
    3.8. Compiler flags exercises

4. Execution Tracer
5. Memory and Garbage Collector
6. Tips and trips

Dave Cheney dave@cheney.net (v379996b, 2019-07-24) 
```

# 3. Compiler optimisations

This section covers some of the optimisations that the Go compiler performs.

For example:

- `Escape analysis`
- `Inlining`

These are handled in the **front end of the compiler**, while the code is still in its **`AST` form**.

Then the code is passed to the `SSA compiler` for further optimisation like:

- `Dead code` elimination
- `Bounds check` elimination
- `Nil check` elimination

----

### 3.1. History of the Go compiler

In 2015 the Go 1.5 compiler was mechanically translated from `C` into `Go`.

A year later, Go 1.7 introduced a [new compiler backend](https://blog.golang.org/go1.7) based on [SSA](https://en.wikipedia.org/wiki/Static_single_assignment_form) (`Static single assignment`) techniques replaced the previous Plan 9 style code generation.

This new backend introduced many opportunities for **generic and architecture specific** optimistions.

----

### 3.2. Escape analysis

The Go spec does not mention the heap or the stack.

It only mentions that the language is garbage collected, and gives no hints as to how this is to be achieved.

**A compliant Go implementation** of the Go spec *could* store **every allocation on the heap**.

- That would put a lot of pressure on the the garbage collector
- But it would be in no way incorrect

A goroutine's **stack** exists as a **cheap place to store local variables**.

- There is no need to garbage collect things on the stack as they are effectively collected when the function returns.

Therefore, where it is safe to do so, **an allocation placed on the stack will be more efficient**.

In Go:

- The compiler automatically moves a **value to the heap** if it lives beyond the lifetime of the function call.
- It is said that the **value escapes to the heap**.

```go
type Foo struct {
	a, b, c, d int
}

func NewFoo() *Foo {
	return &Foo{a: 3, b: 1, c: 4, d: 7} // Foo escapes to the heap
}
```

In this example the `Foo` instance allocated in `NewFoo` will be moved to the **heap** so its contents remain valid after `NewFoo` has returned.


This has been present since the earliest days of Go. It isn't so much an optimisation as **an automatic correctness feature**.

> Accidentally returning the address of a stack allocated variable is **not possible in Go**.

But the compiler can also do the opposite:

- It can find things which would be assumed to be allocated on the heap, and **move them to stack**.

```go
// Sum returns the sum of the numbers 1 to 100.
func Sum() int {
	const count = 100
	numbers := make([]int, count) // numbers is only referenced inside Sum => does not escape.
	for i := range numbers {
		numbers[i] = i + 1
	}

	var sum int
	for _, i := range numbers {
		sum += i
	}
	return sum
}

func main() {
	answer := Sum()
    fmt.Println(answer) // fmt.Println is a variadic function ...
}
```

Because the **`numbers` slice** is only referenced inside `Sum`

- The compiler will arrange to **store the 100 integers** for that slice **on the stack**, rather than the heap.
- There is **no need to garbage collect `numbers`**, it is **automatically freed** when `Sum` returns.

----

### 3.2.1. Prove it!

To print the compilers **escape analysis decisions**, use the `-m` flag.

```sh
go build -gcflags=-m examples/esc/sum.go	# -m = escape analysis decisions

examples/esc/sum.go:22:13: inlining call to fmt.Println
examples/esc/sum.go:8:17: make([]int, count) does not escape
examples/esc/sum.go:22:13: answer escapes to heap
examples/esc/sum.go:22:13: []interface {} literal does not escape
```

Line 8 shows the compiler has correctly deduced that the result of `make([]int, 100)` does not escape to the heap.

The reason line 22 reports that `answer` escapes to the heap is because `fmt.Println` is a **`variadic function`**.

- The **parameters** to a **variadic function** are `boxed` **into a slice**, in this case a `[]interface{}` (slice of interfaces)
- `answer` is placed into an interface value because it is **referenced by the call to** `fmt.Println`

Since Go 1.6 the **garbage collector requires** all values passed via an interface to be **pointers**.

```go
// What the compiler sees is approximately:
var answer = Sum()
fmt.Println([]interface{&answer}...)
```

We can confirm this using the -gcflags="-m -m" flag.

Which returns:

```sh
go build -gcflags='-m -m' examples/esc/sum.go 2>&1 | grep sum.go:22

examples/esc/sum.go:22:13: inlining call to fmt.Println func(...interface {}) (int, error) { var fmt..autotmp_3 int; fmt..autotmp_3 = <N>; var fmt..autotmp_4 error; fmt..autotmp_4 = <N>; fmt..autotmp_3, fmt..autotmp_4 = fmt.Fprintln(io.Writer(os.Stdout), fmt.a...); return fmt..autotmp_3, fmt..autotmp_4 }
examples/esc/sum.go:22:13: answer escapes to heap:
examples/esc/sum.go:22:13:   flow: ~arg0 = &{storage for answer}:
examples/esc/sum.go:22:13:     from answer (spill) at examples/esc/sum.go:22:13
examples/esc/sum.go:22:13:     from ~arg0 = <N> (assign-pair) at examples/esc/sum.go:22:13
examples/esc/sum.go:22:13:   flow: {storage for []interface {} literal} = ~arg0:
examples/esc/sum.go:22:13:     from []interface {} literal (slice-literal-element) at examples/esc/sum.go:22:13
examples/esc/sum.go:22:13:   flow: fmt.a = &{storage for []interface {} literal}:
examples/esc/sum.go:22:13:     from []interface {} literal (spill) at examples/esc/sum.go:22:13
examples/esc/sum.go:22:13:     from fmt.a = []interface {} literal (assign) at examples/esc/sum.go:22:13
examples/esc/sum.go:22:13:   flow: {heap} = *fmt.a:
examples/esc/sum.go:22:13:     from fmt.Fprintln(io.Writer(os.Stdout), fmt.a...) (call parameter) at examples/esc/sum.go:22:13
examples/esc/sum.go:22:13: answer escapes to heap
examples/esc/sum.go:22:13: []interface {} literal does not escape
```

In short, don't worry about line 22, it's not important to this discussion.

----

### 3.2.2. Exercises

- Does this optimisation hold true for all values of count?

    ```go
	// No. Values for "count => 8192" -> escape to heap 
	const count = 1<<13 // 2^13 = 8192
    ```

    ```sh
    go build -gcflags='-m -m' examples/esc/sum.go 2>&1 | grep sum.go:8

	examples/esc/sum.go:8:17: make([]int, count) escapes to heap:
	examples/esc/sum.go:8:17:   flow: {heap} = &{storage for make([]int, count)}:
	examples/esc/sum.go:8:17:     from make([]int, count) (non-constant size) at examples/esc/sum.go:8:17
	examples/esc/sum.go:8:17: make([]int, count) escapes to heap
    ```

- Does this optimisation hold true if count is a variable, not a constant?

    ```go
	// No. non-constant size -> escape to heap 
	var count = 100
    ```

    ```sh
    go build -gcflags='-m -m' examples/esc/sum.go 2>&1 | grep sum.go:8

	examples/esc/sum.go:8:17: make([]int, count) escapes to heap:
	examples/esc/sum.go:8:17:   flow: {heap} = &{storage for make([]int, count)}:
	examples/esc/sum.go:8:17:     from make([]int, count) (non-constant size) at examples/esc/sum.go:8:17
	examples/esc/sum.go:8:17: make([]int, count) escapes to heap
    ```

- Does this optimisation hold true if count is a parameter to Sum?

    ```go
	// No. non-constant size -> escape to heap 
	func Sum(count int) int
    ```

    ```sh
    go build -gcflags='-m -m' examples/esc/sum.go 2>&1 | grep sum.go:7

	examples/esc/sum.go:7:17: make([]int, count) escapes to heap:
	examples/esc/sum.go:7:17:   flow: {heap} = &{storage for make([]int, count)}:
	examples/esc/sum.go:7:17:     from make([]int, count) (non-constant size) at examples/esc/sum.go:7:17
	examples/esc/sum.go:7:17: make([]int, count) escapes to heap
    ```

### 3.2.3. Escape analysis (continued)

This example is a little contrived. It is not intended to be real code, just an example.

```go
type Point struct{ X, Y int }

const Width = 640
const Height = 480

func Center(p *Point) {
	p.X = Width / 2
	p.Y = Height / 2
}

func NewPoint() {
	p := new(Point)
	Center(p)
	fmt.Println(p.X, p.Y)
}
```

- `NewPoint` creates a new `*Point` value, `p`.
- We pass `p` to the `Center` function which moves the point to a position in the center of the screen.
- Finally we print the values of `p.X` and `p.Y`

```sh
go build -gcflags=-m examples/esc/center.go

examples/esc/center.go:10:6: can inline Center
examples/esc/center.go:17:8: inlining call to Center
examples/esc/center.go:18:13: inlining call to fmt.Println
examples/esc/center.go:10:13: p does not escape
examples/esc/center.go:16:10: new(Point) does not escape
examples/esc/center.go:18:15: p.X escapes to heap
examples/esc/center.go:18:20: p.Y escapes to heap
examples/esc/center.go:18:13: []interface {} literal does not escape
```

Even though `p` was allocated with the `new` function

- it will not be stored on the heap,
- because `p does not escape` the `Center` function.

Write a benchmark to prove that Sum does not allocate.

```go
var Result int

func BenchmarkSum(b *testing.B) {
	b.ReportAllocs()
	var r int
	for i := 0; i < b.N; i++ {
		r = Sum()
	}
	Result = r
}

// go test -bench=. ./examples/esc/
// BenchmarkSum-12		13020309	89.8 ns/op	0 B/op	0 allocs/op
```

----

## 3.3. Inlining

In Go function calls have a fixed overhead; **stack and preemption checks**.

Some of this is ameliorated by hardware branch predictors, but it's still a **cost in terms of function size and clock cycles**.

> Inlining is the classical optimisation that avoids these costs.

Until Go 1.11 inlining only worked on `leaf functions`, a function that does not call another. 

The justification for this is:
-  If a function does a lot of work, then the preamble overhead will be negligible.
   - functions over a certain size (currently some count of instructions, plus a few operations which prevent inlining)
- **small functions** on the other hand pay a **fixed overhead** for a relatively small amount of useful work performed. 
  - These are the functions that **inlining targets** as they benefit the most.

The other reason is that heavy inlining makes stack traces harder to follow.

----

### 3.3.1. Inlining (example)

```go
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func F() {
	const a, b = 100, 20
	if Max(a, b) == b {
		panic(b)
	}
}

// go build -gcflags=-m examples/inl/max.go
func main() {
	F()
}
```

Again we use the `-gcflags=-m` flag to view the compilers optimisation decision.

```sh
go build -gcflags=-m examples/inl/max.go

examples/inl/max.go:3:6: can inline Max
examples/inl/max.go:10:6: can inline F
examples/inl/max.go:12:8: inlining call to Max
examples/inl/max.go:16:6: can inline main
examples/inl/max.go:17:3: inlining call to F
examples/inl/max.go:17:3: inlining call to Max
```

The compiler printed two lines:

 - The first at line 3, the declaration of `Max`, telling us that it can be inlined.
 - The second is reporting that the `body of Max` has been inlined into the caller at line 12.

----

### 3.3.2. What does inlining look like?

Compile `max.go` and see what the optimised version of `F()` became.

```sh
go build -gcflags=-S examples/inl/max.go 2>&1 | grep -A5 '"".F STEXT'
"".F STEXT nosplit size=1 args=0x0 locals=0x0
	0x0000 00000 (examples/inl/max.go:10)	TEXT	"".F(SB), NOSPLIT|ABIInternal, $0-0
	0x0000 00000 (examples/inl/max.go:10)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (examples/inl/max.go:10)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (examples/inl/max.go:12)	RET
```

This is the body of `F` once `Max` has been inlined into it — there's nothing happening in this function. 

I know there's a lot of text on the screen for nothing, but take my word for it, the only thing happening is the `RET`.

In effect `F` became:

```go
func F() {
        return
}
```

For the rest of the presentation I'll be using a small shell script to reduce the clutter in the assembly output.

```sh
# asm.sh
go build -gcflags=-S 2>&1 $@ | grep -v PCDATA | grep -v FUNCDATA | less

# ./asm.sh ./examples/inl/max.go
```

> What are `FUNCDATA` and `PCDATA`?
>
> The output from -S is not the final machine code that goes into your binary. The linker does some processing during the final link stage.
>
> Lines like `FUNCDATA` and `PCDATA` are **metadata for the garbage collector** which are moved elsewhere when linking.
>
> If you're reading the output of -S, **just ignore** FUNCDATA and PCDATA lines; they're **not part of the final binary**.

----

### 3.3.3. Discussion

Why did I declare `a` and `b` in `F()` to be **constants**?
- Compiler knows the size of `a+b` at compile-time; allows for if-branch elimination after inlining of `Max`.

What happens if `a` and `b` are declared as **variables**?
- Inlining of the **full body** of `Max`; no branch elimination.

What happens if `a` and `b` are passing into `F()` as **parameters**?
- Body of `F` grows quite a lot.

----

## 3.4. Dead code elimination

Why is it important that `a` and `b` are constants?

To understand what happened lets look at what the compiler sees once it has inlined `Max` into `F`.

We can't get this from the compiler easily, but it’s straight forward to do it by hand.

Berfore:

```go
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func F() {
	const a, b = 100, 20
	if Max(a, b) == b {
		panic(b)
	}
}
```

After:

```go
func F() {
	const a, b = 100, 20
	var result int
	if a > b {
		result = a
	} else {
		result = b
	}
	if result == b {
		panic(b)
	}
}
```

Because `a` and `b` are constants the compiler can prove at compile time that the branch will never be false; `100` is always greater than `20`.

So the compiler can further optimise `F` to:

```go
func F() {
	const a, b = 100, 20
	var result int
	if true {
		result = a
	} else {
		result = b
	}
	if result == b {
		panic(b)
	}
}
```

Now that the result of the branch is know then the contents of `result` are also known.

This is called `branch elimination`.

```go
func F() {
    const a, b = 100, 20
    const result = a
    if result == b {
            panic(b)
    }
}
```

Now the branch is eliminated we know that `result` is always equal to `a`, and because `a` was a constant, we know that `result` is a constant.

The compiler applies this proof to the second branch

```go
func F() {
    const a, b = 100, 20
    const result = a
    if false {
            panic(b)
    }
}
```

And using `branch elimination` again the final form of `F` is reduced to.

```go
func F() {
        const a, b = 100, 20
        const result = a
}
```

And finally just

```go
func F() {
}
```
----

### 3.4.1. Dead code elimination (cont.)

Branch elimination is one of a category of optimisations known as `dead code elimination`. In effect, using **static proofs** to show that a piece of code is never reachable, colloquially known as **dead**, therefore it **need not be compiled, optimised, or emitted** in the final binary.

We saw how `dead code elimination` **works together with** `inlining` to reduce the amount of code generated by **removing loops and branches** that are proven unreachable.

> You can take advantage of this to implement expensive **debugging**, and hide it behind:

```go
const debug = false
```

Combined with `build tags` this can be very useful.

----

### 3.4.2. Adjusting the level of inlining

Adjusting the inlining level is performed with the `-gcflags=-l` flag.

- nothing, regular inlining.
- `-gcflags=-l`, inlining disabled.
- `-gcflags='-l -l'` inlining level 2, more aggressive, might be faster, may make bigger binaries.
- `-gcflags='-l -l -l'` inlining level 3, more aggressive again, binaries definitely bigger, maybe faster again, but might also be buggy.
- `-gcflags=-l=4` (four `-l`s) in **Go 1.11** will enable the **experimental** `mid stack inlining` optimisation.

----

### 3.4.3. Mid Stack inlining

**Since Go 1.12** so called `mid stack inlining` has been **enabled**.

We can see an example of mid stack inlining in the previous example.

Because of inlining improvements `F` is now `inlined into its caller`.

This is for two reasons:
- When `Max` is inlined into `F`, `F` contains no other function calls thus it becomes a **potential leaf function**, assuming its **complexity budget** has not been exceeded.
- Because `F` is a simple function — `​inlining and dead code elimination` has **eliminated much of its complexity budget** — ​it is eligable for `mid stack inlining` irrispective of calling `Max`.

----

### 3.4.4. Further reading

- [Using // +build to switch between debug and release builds](https://dave.cheney.net/2014/09/28/using-build-to-switch-between-debug-and-release)
- [How to use conditional compilation with the go build tool](http://dave.cheney.net/2013/10/12/how-to-use-conditional-compilation-with-the-go-build-tool)

```sh
# =============================================================================
# START - Further reading notes ===============================================
# =============================================================================

go list -f '{{.GoFiles}}' os/exec
# [exec.go exec_unix.go lp_unix.go]

# =============================================================================
# build tag found at the top of a source file
// +build darwin freebsd netbsd openbsd
# constrains this file to only building on BSD systems
# =============================================================================

# =============================================================================
# A file may have multiple build tags.
# The overall constraint is the logical AND of the individual constraints
// +build linux darwin
// +build 386
# constrains the build to linux/386 or darwin/386 platforms only
# =============================================================================

# =============================================================================
// Copyright 2013 Way out enterprises. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build someos someotheros thirdos,!amd64

// Package headspin implements calculates numbers so large
// they will make your head spin.
package headspin
# =============================================================================

go help buildconstraint
https://golang.org/pkg/go/build/

// +build integration
go test -tags=integration
https://peter.bourgon.org/go-in-production/#testing-and-validation
https://stackoverflow.com/questions/25965584/separating-unit-tests-and-integration-tests-in-go
```

----

### Debugging is optional

When I committed the fix for this bug I didn’t have to spend any time removing the debug function calls inside the package.

```go
// +build debug

package sftp

import "log"

func debug(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}

// https://github.com/pkg/sftp/blob/master/debug.go
```

Adding a call to debug inside sendPacket it was easy to figure out the packet which was being incorrectly encoded.

`go test -tags debug -integration -v -run=Lstat`

```go
// +build !debug

package sftp

func debug(fmt string, args ...interface{}) {}

// https://github.com/pkg/sftp/blob/master/release.go
```

When `-tags debug` is not present, the version from release.go, effectively a no-op, is used.

#### Extra credit

This package includes integration tests which are not run by default. How the -integration test flag works is left as an exercise to the reader.

```Makefile
integration:
	go test -integration -v ./...
	go test -testserver -v ./...
	go test -integration -testserver -v ./...
	go test -integration -allocator -v ./...
	go test -testserver -allocator -v ./...
	go test -integration -testserver -allocator -v ./...

integration_w_race:
	go test -race -integration -v ./...
	go test -race -testserver -v ./...
	go test -race -integration -testserver -v ./...
	go test -race -integration -allocator -v ./...
	go test -race -testserver -allocator -v ./...
	go test -race -integration -allocator -testserver -v ./...

# https://github.com/pkg/sftp/blob/master/Makefile



# =============================================================================
# END - Further reading notes =================================================
# =============================================================================
```

----

## 3.5. [Prove pass](https://dave.cheney.net/high-performance-go-workshop/gophercon-2019.html#prove_pass)

A few releases ago the SSA backend gained a `prove pass`.

Prove establishes the `relationship between variables`.

Let's look at an example to explain what prove is doing.

```go
package main

func foo(x int32) bool {
	if x > 5 {  // At this point the compiler knows that x is greater than 5
		if x > 3 { // Therefore x is also greater than 3, the branch is always taken.
			return true
		}
		panic("x less than 3")
	}
	return false
}

func main() {
	foo(-1)
}
```

### 3.5.1. Prove it

```sh
go build -gcflags=-d=ssa/prove/debug=on examples/prove/foo.go

examples/prove/foo.go:5:10: Proved Less32
```

Line 5 `if x > 3`: The compiler is saying that it has proven that the branch will always be true.

----

## 3.6 Compiler intrinsics

Go allows you to `write functions in assembly` if required.

The technique involves a `forwarding declared function` — ​a function without a body — ​**and a corresponding** `assembly function`.

```go
// decl.go
package asm

// Add returns the sum of a and b.
func Add(a int64, b int64) int64
```

Note the `Add` function **has no body**.

To satisfy the compiler we must supply the `assembly` for this function, which we can do via a `.s file` in the same package.

```go
// add.s
TEXT ·Add(SB),$0
	MOVQ a+0(FP), AX
	ADDQ b+8(FP), AX
	MOVQ AX, ret+16(FP)
	RET
```

Now we can build, test, and use our `asm.Add` function just like normal Go code.

But there’s a problem, `assembly functions cannot be inlined`.

There have been various proposals for an inline assembly syntax for Go, but they have not been accepted by the Go developers.

Instead, Go has added `intrinsic functions`.

An `intrinsic function` is regular Go code written in regular Go, however the **compiler contains specific drop in replacements** for the functions.

The two packages that make use of this are:

 - `math/bits`
 - `sync/atomic`

These replacements are implemented in the **compiler backend**:
- if your **architecture** supports a **faster way of doing an operation**
- it will be **transparently replaced with the comparable instruction** during compilation.

As well as generating more efficient code, because `intrinsic functions` are just normal Go code, **the rules of `inlining, and mid stack inlining` apply to them**.

----

### 3.6.1. Popcnt example

Population count is an important crypto operation so modern CPUs have a **native instruction** to perform it.

The `math/bits` package provides a set of functions, `OnesCount…`​ which are recognised by the compiler and **replaced with their native equivalent**.

```go
func BenchmarkMathBitsPopcnt(b *testing.B) {
	var r int
	for i := 0; i < b.N; i++ {
		r = bits.OnesCount64(uint64(i)) // intrinsic function
	}
	Result = uint64(r)
}
```

Run the benchmark and compare the performance of the hand rolled shift implementation and `math/bits.OnesCount64`.

```sh
go test -bench=.  ./examples/popcnt-intrinsic/

BenchmarkPopcnt-12            	777447768	         1.51 ns/op
BenchmarkMathBitsPopcnt-12    	1000000000	         0.585 ns/op
```

### 3.6.2. Atomic counter example

Here's an example of an atomic counter type.

We've got methods on types, method calls several levels deep, multiple packages, etc.

You'd be forgiven for thinking this might have a lot of overhead.

```go
package main

import (
	"sync/atomic"
)

type counter uint64

func (c *counter) get() uint64 {
	return atomic.LoadUint64((*uint64)(c))
}
func (c *counter) inc() uint64 {
	return atomic.AddUint64((*uint64)(c), 1)
}
func (c *counter) reset() uint64 {
	return atomic.SwapUint64((*uint64)(c), 0)
}

var c counter

func f() uint64 {
	c.inc()
	c.get()
	return c.reset()
}

func main() {
	f()
}
```

But, because of the interation between `inlining` and `compiler intrinsics`, this **code collapses down to efficient native code on most platforms**.

```sh
bash asm.sh ./examples/counter/counter.go

"".f STEXT nosplit size=36 args=0x8 locals=0x0
        0x0000 00000 (examples/counter/counter.go:21) TEXT    "".f(SB), NOSPLIT|ABIInternal, $0-8
        0x0000 00000 (<unknown line number>)    NOP
        0x0000 00000 (examples/counter/counter.go:22) MOVL    $1, AX
        0x0005 00005 (examples/counter/counter.go:13) LEAQ    "".c(SB), CX
        0x000c 00012 (examples/counter/counter.go:13) LOCK
        0x000d 00013 (examples/counter/counter.go:13) XADDQ   AX, (CX)         # 1
        0x0011 00017 (examples/counter/counter.go:23) XCHGL   AX, AX
        0x0012 00018 (examples/counter/counter.go:10) MOVQ    "".c(SB), AX     # 2
        0x0019 00025 (<unknown line number>)    NOP
        0x0019 00025 (examples/counter/counter.go:16) XORL    AX, AX
        0x001b 00027 (examples/counter/counter.go:16) XCHGQ   AX, (CX)         # 3
        0x001e 00030 (examples/counter/counter.go:24) MOVQ    AX, "".~r0+8(SP)
        0x0023 00035 (examples/counter/counter.go:24) RET
# 1 => c.inc()
# 2 => c.get()
# 3 => c.reset()
```

Further reading:

- [Mid-stack inlining in the Go compiler](https://docs.google.com/presentation/d/1Wcblp3jpfeKwA0Y4FOmj63PW52M_qmNqlQkNaLj0P5o/edit#slide=id.p) by David Lazar
- [Proposal: Mid-stack inlining in the Go compiler](https://github.com/golang/proposal/blob/master/design/19348-midstack-inlining.md)

----

## 3.7. Bounds check elimination

Go is a bounds checked language. This means array and slice subscript operations are checked to ensure they are within the bounds of the respective types.

For arrays, this can be done at compile time. For slices, this must be done at runtime.

```go
var v = make([]int, 9)

var A, B, C, D, E, F, G, H, I int

func BenchmarkBoundsCheckInOrder(b *testing.B) {
	var a, _b, c, d, e, f, g, h, i int
	for n := 0; n < b.N; n++ {
		a = v[0]
		_b = v[1]
		c = v[2]
		d = v[3]
		e = v[4]
		f = v[5]
		g = v[6]
		h = v[7]
		i = v[8]
	}
	A, B, C, D, E, F, G, H, I = a, _b, c, d, e, f, g, h, i
}

func BenchmarkBoundsCheckOutOfOrder(b *testing.B) {
	var a, _b, c, d, e, f, g, h, i int
	for n := 0; n < b.N; n++ {
		i = v[8]
		a = v[0]
		_b = v[1]
		c = v[2]
		d = v[3]
		e = v[4]
		f = v[5]
		g = v[6]
		h = v[7]
	}
	A, B, C, D, E, F, G, H, I = a, _b, c, d, e, f, g, h, i
}

// go test -gcflags=-S -bench=BoundsCheckInOrder  ./examples/bounds/

// How many bounds check operations are performed per loop?
```

```sh
go test -bench=.  ./examples/bounds/

BenchmarkBoundsCheckInOrder-12       	471346575	         2.52 ns/op
BenchmarkBoundsCheckOutOfOrder-12    	654611022	         1.82 ns/op
```


### 3.7.1. Exercises

What happens if v is moved inside the Benchmark function?

```sh
go test -bench=.  ./examples/bounds/

BenchmarkBoundsCheckInOrder-12       	1000000000	         0.266 ns/op
BenchmarkBoundsCheckOutOfOrder-12    	1000000000	         0.256 ns/op
```

What happens if v was declared as an array, `var v [9]int`?

```sh
go test -bench=.  ./examples/bounds/

BenchmarkBoundsCheckInOrder-12       	919924052	         1.26 ns/op
BenchmarkBoundsCheckOutOfOrder-12    	954235348	         1.26 ns/op
```

----

## 3.8. Compiler flags exercises

Investigate the operation of the following compiler functions:

```sh
go test -gcflags=-S         # prints the (Go flavoured) assembly of the package being compiled.

go test -gcflags=-l         # disables inlining
go test -gcflags='-l -l'    # increases it
go test -gcflags='-l -l -l' # increases it even more

go test -gcflags=-m         # controls printing of optimisation decision
go test -gcflags='-m -m'    # prints more details about what the compiler was thinking

go test -gcflags='-l -N'    # disables all optimisations

go test -gcflags=-d=ssa/prove/debug=on  # this also takes values of 2 and above, see what prints

go test -gcflags=

# go tool compile -d help
# go tool compile -d=ssa/<phase>/<flag>[=<value>|<function_name>]
# go test -gcflags=-d=ssa/prove/debug=on -bench=.  ./examples/bounds/bounds_test.go
go test -gcflags=-d=ssa/prove/debug=2 -bench=.  ./examples/bounds/bounds_test.go

examples/bounds/bounds_test.go:14:19: Induction variable: limits [0,?), increment 1
examples/bounds/bounds_test.go:32:19: Induction variable: limits [0,?), increment 1

BenchmarkBoundsCheckInOrder-12       	1000000000	         0.256 ns/op
BenchmarkBoundsCheckOutOfOrder-12    	1000000000	         0.252 ns/op

# Disassembler
# https://go-talks.appspot.com/github.com/rakyll/talks/gcinspect/talk.slide#5
# go tool objdump -s main.main hello
# GOSSAFUNC=main go build && open ssa.html
# go build -gcflags="-S"
# go build -gcflags="-m" golang.org/x/net/context
# go build -gcflags="-l -N"
# go build -x
```

----

(Performance measurement and profiling) [prev](02-Performance-measurement-and-profiling.md) | [next](04-Execution-Tracer.md) (Execution Tracer)
