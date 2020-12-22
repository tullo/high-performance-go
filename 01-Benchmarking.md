
```
High Performance Go Workshop

1. Benchmarking

    1.1. Benchmarking ground rules
    1.2. Using the testing package for benchmarking
    1.3. Comparing benchmarks with benchstat
    1.4. Avoiding benchmarking start up costs
    1.5. Benchmarking allocations
    1.6. Watch out for compiler optimisations
    1.7. Benchmark mistakes
    1.8. Benchmarking math/rand
    1.9. Profiling benchmarks
    1.10. Discussion

2. Performance measurement and profiling
3. Compiler optimisations
4. Execution Tracer
5. Memory and Garbage Collector
6. Tips and trips

Dave Cheney dave@cheney.net (v379996b, 2019-07-24) 
```

# 1. Benchmarking

> Measure twice and cut once. — Ancient proverb

Before we attempt to improve the performance of a piece of code, first we must know its current performance.

This section focuses on how to construct useful benchmarks using the Go testing framework, and gives practical tips for avoiding the pitfalls.

## 1.1 Benchmarking ground rules

Before you benchmark, you must have a stable environment to get repeatable results.

- The machine must be idle — ​don’t profile on shared hardware, don’t browse the web while waiting for a long benchmark to run.
- Watch out for power saving and thermal scaling. These are almost unavoidable on modern laptops (hot labs (;-] ).
- Avoid virtual machines and shared cloud hosting; they can be too noisy for consistent measurements.

If you can afford it:
- buy dedicated performance test hardware. 
- rack it, disable all the power management and thermal scaling,
- and never update the software on those machines.

The last point is poor advice from a system adminstration point of view, but if a software update changes the way the kernel or library performs — ​think the Spectre patches — ​this will invalidate any previous benchmarking results.

For the rest of us:
- have a before and after sample
- and run them multiple times

to get consistent results.

## 1.2 Using the testing package for benchmarking

The testing package has built in support for writing benchmarks.

If we have a simple function like this:

```go
func Fib3(n int) int {
	switch n {
	case 0:
		return 0
	case 1:
		return 1
	case 2:
		return 1
	default:
		return Fib(n-1) + Fib(n-2)
	}
}

// https://en.wikipedia.org/wiki/Fibonacci_number
```

The we can use the testing package to write a benchmark for the function using this form.

```go
func BenchmarkFib1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Fib(1) // run the Fib function b.N times
	}
}

func BenchmarkFib28(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Fib(28) // run the Fib function b.N times
	}
}
```

> The benchmark function lives alongside your tests in a _test.go file. 

Benchmarks are similar to tests, the only real difference is they take a *testing.B rather than a *testing.T.

Both of these types implement the testing.TB interface which provides crowd favorites like Errorf(), Fatalf(), and FailNow().

### 1.2.1. Running a package’s benchmarks

As benchmarks use the testing package they are executed via the go test subcommand.

However, by default when you invoke go test, **benchmarks are excluded**.

To explicitly run benchmarks in a package use the -bench flag. 
- -bench takes a regular expression that matches the names of the benchmarks you want to run, 
- so the most common way to invoke all benchmarks in a package is `-bench=.`

Here is an example:

```sh
make fib-bench-all
go test -bench=. ./examples/fib/

goos: linux
goarch: amd64
BenchmarkFib1-12     	746697031	         1.58 ns/op
BenchmarkFib20-12    	    32812	     36629    ns/op
PASS
```

> go test will also run all the tests in a package before matching benchmarks, 
> 
> so if you have a lot of tests in a package, or they take a long time to run, 
> 
> you can exclude them by providing go test’s `-run flag with a regex that matches nothing; ie.
>
> `go test -run=^$`

### 1.2.2. How benchmarks work

Each benchmark function is called with different value for b.N, this is the number of iterations the benchmark should run for.

b.N starts at 1, if the benchmark function completes in under 1 second — ​the default — ​then b.N is increased and the benchmark function run again.

b.N increases in the approximate sequence, growing by roughly **20%** for each iteration.

The benchmark framework tries to be smart and if it sees small values of b.N are completing relatively quickly, it will increase the iteration count faster.

Looking at the example above, BenchmarkFib20-8 found that around `32'812 iterations of the loop` took just over a second.

From there the benchmark framework computed that the average time per operation was 36629ns.

> The `-12` suffix relates to the value of [GOMAXPROCS](https://golang.org/pkg/runtime/) that was used to run this test. This number **defaults to the number of OS-threads visible to the Go process on startup**. 
>
> You can change this value with the -cpu flag which takes a list of values to run the benchmark with.
> ```sh
> go test -bench=Fib20 -cpu=1,2,4 ./examples/fib
> BenchmarkFib20     	   33140	     35639 ns/op
> BenchmarkFib20-2   	   33775	     35561 ns/op
> BenchmarkFib20-4   	   32001	     35590 ns/op
> ```
> This shows running the benchmark with 1, 2, and 4 OS-threads.
> 
> In this case the flag has little effect on the outcome because this benchmark is entirely sequential.

### 1.2.3. Go 1.13 benchmarking changes

In Go 1.13 the rounding has been removed, which improves the accuracy of benchmarking operations in the low ns/op range, and reduces the run time of benchmarks overall as the benchmark framework arrives at the correct iteration count faster.

### 1.2.4. Improving benchmark accuracy

The fib function is a slightly contrived example—​unless you are writing a TechPower web server benchmark — ​it’s unlikely your business is going to be gated on how quickly you can compute the 20th number in the Fibonaci sequence.

But, the benchmark does provide a faithful example of a valid benchmark.

Specifically you want your benchmark to run for several tens of thousand iterations so you get a good average per operation. 

> If your benchmark runs for only 100’s or 10’s of iterations, the average of those runs may have a high standard deviation.

> If your benchmark runs for millions or billions of iterations, the average may be very accurate, but subject to the vaguaries of code layout and alignment.

To increase the number of iterations, the benchmark time can be increased with the -benchtime flag. For example:

```sh
go test -bench=Fib20 -benchtime=10s ./examples/fib/

BenchmarkFib20-12    	  330016	     35252 ns/op
PASS
ok  	examples/fib	11.935s
```

It ran the same benchmark until it reached a value of b.N that took longer than 10 seconds to return.

As we’re running for 10x longer, the total number of iterations is 10x larger.

The result hasn’t changed much, which is what we expected.

---

For times measured in `10 or single digit nanoseconds per operation` the relativistic effects of instruction reordering and code alignment will have an impact on your benchmark times.

> To address this **run benchmarks multiple times** with the `-count` flag:

```go
go test -run=^$ -bench=Fib1 -count=10 ./examples/fib

goos: linux
goarch: amd64
BenchmarkFib1-12    	774237474	         1.51 ns/op
BenchmarkFib1-12    	788348062	         1.51 ns/op
BenchmarkFib1-12    	787899610	         1.51 ns/op
BenchmarkFib1-12    	786285502	         1.52 ns/op
BenchmarkFib1-12    	792306502	         1.51 ns/op
BenchmarkFib1-12    	789509175	         1.51 ns/op
BenchmarkFib1-12    	779690390	         1.51 ns/op
BenchmarkFib1-12    	783065533	         1.51 ns/op
BenchmarkFib1-12    	773222475	         1.51 ns/op
BenchmarkFib1-12    	781180959	         1.52 ns/op
PASS
ok  	examples/fib	13.420s
```
A benchmark of Fib(1) takes around `1.5 nano seconds` with a variance of `+/- 15%`.

    b.N = 774237474 (number of for-loop iterations per second)
    op = 1.5 ns     (time used to complete the fn-call)
    fn = Fib(1)     (1st invocation)

In Go 1.12 the -benchtime flag now takes a **number of iterations**, eg. `-benchtime=20x` which will **run your code exactly benchtime times**.

> Try running the fib bench above with a -benchtime of 10x, 20x, 50x, 100x, and 300x. What do you see?

```sh
make fib-bench-fib1-iterations

BenchmarkFib1-12    	      10	        27.6  ns/op
BenchmarkFib1-12    	      20	         7.40 ns/op
BenchmarkFib1-12    	      50	         6.62 ns/op
BenchmarkFib1-12    	     100	         3.93 ns/op
BenchmarkFib1-12    	     300	         2.43 ns/op
```

> If you find that the defaults that go test applies need to be tweaked for a particular package, 
> 
> I suggest codifying those settings in a Makefile so everyone who wants to run your benchmarks can do so with the same settings. 

## 1.3 Comparing benchmarks with benchstat

I suggest running benchmarks more than once to get more data to average. This is good advice for any benchmark because of the effects of power management, background processes, and thermal management that I mentioned at the start of the chapter.

I’m going to introduce a tool by Russ Cox called [benchstat](https://godoc.org/golang.org/x/perf/cmd/benchstat).

Benchstat can take a set of benchmark runs and tell you how stable they are. 

Here is an example of Fib(20):

```sh
# tee ==> read from standard input and write to standard output and files
#
go test -bench=Fib20 -count=10 ./examples/fib/ | tee old.txt

BenchmarkFib20-12    	   33691	     35207 ns/op
BenchmarkFib20-12    	   34060	     35272 ns/op
BenchmarkFib20-12    	   34057	     35240 ns/op
BenchmarkFib20-12    	   33902	     35220 ns/op
BenchmarkFib20-12    	   33418	     35857 ns/op
BenchmarkFib20-12    	   33747	     35937 ns/op
BenchmarkFib20-12    	   33400	     35214 ns/op
BenchmarkFib20-12    	   33879	     35871 ns/op
BenchmarkFib20-12    	   34027	     35193 ns/op
BenchmarkFib20-12    	   33488	     35194 ns/op

$(go env GOPATH)/bin/benchstat old.txt
name      time/op
Fib20-12  35.4µs ± 1%
```

`benchstat` tells us the `mean` is 35.4 microseconds with a +/- 1% `variation` across the samples. This is because while the benchmark was running I didn’t touch the machine.

### 1.3.1. Improve `Fib`

Determining the performance delta between two sets of benchmarks can be tedious and error prone. 

Benchstat can help us with this.

> Saving the output from a benchmark run is useful,
but you can also save the binary that produced it.
this lets you rerun benchmark previous iterations.
>
> To do this, use the `-c flag` to save the test binary
> 
> ​I often rename this binary from .test to .golden
> ```sh
> go test -c ./examples/fib/
> mv fib.test fib.golden
> ```

The previous Fib fuction had hard coded values for the 0th and 1st numbers in the fibonaci series. After that; the code calls itself recursively.

> We’ll talk about the cost of recursion later today, but for the moment, assume it has a cost, especially as our algorithm uses exponential time.

As simple fix to this would be to hard code another number from the fibonacci series, reducing the depth of each recusive call by one.

```go
func Fib2(n int) int {
	switch n {
	case 0:
		return 0
	case 1:
		return 1
	case 2:
		return 1
	default:
		return Fib2(n-1) + Fib2(n-2)
	}
}
```

To compare our new version, we compile a new test binary and benchmark both of them and use benchstat to compare the outputs.

```sh
go test -c ./examples/fib

./fib.golden -test.bench=Fib20 -test.count=10 > old.txt
./fib.test -test.bench=Fib20 -test.count=10 > new.txt

$(go env GOPATH)/bin/benchstat old.txt new.txt
name      old time/op  new time/op  delta
Fib20-12  35.3µs ± 0%  21.7µs ± 0%  -38.39%  (p=0.000 n=8+9)
```

Another version without recursion:

```go
func Fib3(n int) int {
	a, b := 0, 1
	for i := 0; i < n; i++ {
		a, b = b, a+b
	}
	return a
}
```

To compare all three versions

```sh
go test -c ./examples/fib

./fib.golden -test.bench=Fib20 -test.count=10 > fib1.txt
./fib.fib2 -test.bench=Fib20 -test.count=10 > fib2.txt
./fib.test -test.bench=Fib20 -test.count=10 > fib3.txt

$(go env GOPATH)/bin/benchstat fib1.txt fib2.txt fib3.txt
name \ time/op  fib1.txt     fib2.txt     fib3.txt
Fib20-12        35.6µs ± 1%  21.7µs ± 0%  0.0µs ± 3%
```

There are two things to check when comparing benchmarks:

- The variance ± in the old and new times. 
  - 1-2% is good
  - 3-5% is ok
  - greater than 5% and some of your samples will be considered unreliable.
  
  Be careful when comparing benchmarks where one side has a high variance, you may not be seeing an improvement.
- Missing samples.
  - `benchstat` will report how many of the old and new samples it considered to be valid, 
  - sometimes you may find only, say, 9 reported, even though you did -count=10.
  - A 10% or lower rejection rate is ok, higher than 10% may indicate your setup is unstable and you may be comparing too few samples.

### 1.3.2. Beware the p-value

- `p-values` **lower than 0.05** are likely to be statistiaclly significant.
- `p-values` **greater than 0.05** imply the benchmark may not be statistically significant.
- [An example of questionable p-values](https://go-review.googlesource.com/c/go/+/171736)
- Further reading: [P-value](https://en.wikipedia.org/wiki/P-value) (wikipedia).

## 1.4 Avoiding benchmarking start up costs

Sometimes your benchmark has a once per run setup cost.

`b.ResetTimer()` can be used to ignore the time accrued in setup.

```go
// Resetting the benchmark timer - once per run
func BenchmarkExpensive(b *testing.B) {
        boringAndExpensiveSetup()
        b.ResetTimer()
        for n := 0; n < b.N; n++ {
                // function under test
        }
}
```
```go
// Resetting the benchmark timer - per loop iteration
func BenchmarkComplicated(b *testing.B) {
        for n := 0; n < b.N; n++ {
                b.StopTimer() // pause
                complicatedSetup()
                b.StartTimer() // resume
                // function under test
        }
}
```

## 1.5 Benchmarking allocations

Allocation count and size is strongly correlated with benchmark time.

You can tell the testing framework to record the number of allocations made by code under test.

```go
func BenchmarkRead(b *testing.B) {
        b.ReportAllocs() // allocations made by code under test
        for n := 0; n < b.N; n++ {
                // function under test
        }
}
```

Here is an example using the [bufio](https://go.googlesource.com/go/+/refs/heads/master/src/bufio/bufio_test.go) package’s benchmarks.

```sh
# Report allocs for benchmarks using ==> b.ReportAllocs() 
go test -run=^$ -bench=. bufio

goos: linux
goarch: amd64
pkg: bufio
BenchmarkReaderCopyOptimal-12       	16917717	        68.2 ns/op
BenchmarkReaderCopyUnoptimal-12     	10249344	       117 ns/op
BenchmarkReaderCopyNoWriteTo-12     	  487870	      2424 ns/op
BenchmarkReaderWriteToOptimal-12    	 4578621	       260 ns/op
BenchmarkReaderReadString-12        	13596823	       153 ns/op	     144 B/op	       1 allocs/op
BenchmarkWriterCopyOptimal-12       	16265359	        72.8 ns/op
BenchmarkWriterCopyUnoptimal-12     	12794445	       102 ns/op
BenchmarkWriterCopyNoReadFrom-12    	  489555	      2414 ns/op
BenchmarkReaderEmpty-12             	 2291383	       527 ns/op	    4224 B/op	       3 allocs/op
BenchmarkWriterEmpty-12             	 2640555	       455 ns/op	    4096 B/op	       1 allocs/op
BenchmarkWriterFlush-12             	98622181	        11.9 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	bufio	15.840s
```

You can also use the go test `-benchmem flag` to **force** the testing framework **to report allocation statistics for all** benchmarks run.

```sh
# Report allocs for ==> ALL benchmarks 
go test -run=^$ -bench=. -benchmem bufio

goos: linux
goarch: amd64
pkg: bufio
BenchmarkReaderCopyOptimal-12       	16808652	        68.7 ns/op	      16 B/op	       1 allocs/op
BenchmarkReaderCopyUnoptimal-12     	10274464	       116 ns/op	      32 B/op	       2 allocs/op
BenchmarkReaderCopyNoWriteTo-12     	  471523	      2408 ns/op	   32800 B/op	       3 allocs/op
BenchmarkReaderWriteToOptimal-12    	 4655492	       258 ns/op	      16 B/op	       1 allocs/op
BenchmarkReaderReadString-12        	14015299	       156 ns/op	     144 B/op	       1 allocs/op
BenchmarkWriterCopyOptimal-12       	16181314	        72.9 ns/op	      16 B/op	       1 allocs/op
BenchmarkWriterCopyUnoptimal-12     	12604712	        96.3 ns/op	      32 B/op	       2 allocs/op
BenchmarkWriterCopyNoReadFrom-12    	  460353	      2414 ns/op	   32800 B/op	       3 allocs/op
BenchmarkReaderEmpty-12             	 2288095	       525 ns/op	    4224 B/op	       3 allocs/op
BenchmarkWriterEmpty-12             	 2652475	       450 ns/op	    4096 B/op	       1 allocs/op
BenchmarkWriterFlush-12             	98424276	        11.9 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	bufio	16.555s
```

## 1.6 Watch out for compiler optimisations

This example comes from [issue 14813](https://github.com/golang/go/issues/14813#issue-140603392).


```go
const m1 = 0x5555555555555555
const m2 = 0x3333333333333333
const m4 = 0x0f0f0f0f0f0f0f0f
const h01 = 0x0101010101010101

func popcnt(x uint64) uint64 {
	x -= (x >> 1) & m1
	x = (x & m2) + ((x >> 2) & m2)
	x = (x + (x >> 4)) & m4
	return (x * h01) >> 56
}

func BenchmarkPopcnt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		popcnt(uint64(i))
	}
}
```

How fast do you think this function will benchmark? Let’s find out.

```sh
go test -run=^$ -bench=PopcntInline ./examples/popcnt
---
BenchmarkPopcntInline-12       1000000000            0.278 ns/op
```

0.278 of a nano second - that’s basically one clock cycle. 

Even assuming that the CPU may have a few instructions in flight per clock tick, this number seems unreasonably low.

What happened?

To understand what happened, we have to look at the function under benchmake, popcnt. popcnt is a `leaf function` — it does not call any other functions — so the **compiler can inline** it.

Because the function is inlined, the compiler now can see it has **no side effects**:

- popcnt does not affect the state of any global variable.
- the call is eliminated.

This is what the compiler sees:

```go
func BenchmarkPopcnt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// optimised away
	}
}
```

On all versions of the Go compiler that i’ve tested, the loop is still generated. But Intel CPUs are really good at optimising loops, especially empty ones.

### 1.6.1. Exercise, look at the assembly

Before we go on, lets look at the assembly to confirm what we saw.

Use gcflags="-l -S" to disable inlining, how does that affect the assembly output?

```sh
go test -gcflags=-S      # -S assembly output
go test -gcflags="-l -S" # -l disable inlining
```

>**Optimisation is a good thing**
>
> The thing to take away is the same optimisations that **make real code fast, by removing unnecessary computation**, are the same ones that remove benchmarks that have no observable side effects.
>
> This is only going to get more common as the Go compiler improves.

---

### 1.6.2. Fixing the benchmark

Disabling inlining to make the benchmark work is unrealistic; we want to build our code with optimisations on.

To fix this benchmark we must ensure that the **compiler cannot prove** that the body of BenchmarkPopcnt does not cause global state to change.

```go
// This is the recommended way to ensure the compiler cannot optimise away the body of the loop.
var Result uint64

func BenchmarkPopcntNoInline(b *testing.B) {
	var r uint64
	for i := 0; i < b.N; i++ {
		r = popcnt(uint64(i))
	}
	Result = r
}
```

This is the recommended way to ensure the compiler cannot optimise away body of the loop.

- First we use the result of calling popcnt by storing it in `r`.
- Second, because `r` is declared locally inside the scope of `BenchmarkPopcntNoInline` once the benchmark is over, the result of `r` is never visible to another part of the program, 
- as the final act we assign the value of `r` to the package public variable `Result`.

Because `Result is public` the **compiler cannot prove** that another package importing this one
will not be able to see the value of Result changing over time,
hence it cannot optimise away any of the operations leading to its assignment.

```sh
go test -run=^$ -bench=PopcntNoInline ./examples/popcnt
---
BenchmarkPopcntNoInline-12    	776304234	         1.51 ns/op
```

- What happens if we assign to `Result` directly?
- Does this affect the benchmark time?
- What about if we assign the result of popcnt to `_`?

Why can’t the compiler optimise:

```go
func BenchmarkFib20(b *testing.B) {
        var r uint64
        for i := 0; i < b.N; i++ {
                r = Fib(20)
        }
        Result = r
}
```

to simply?

```go
func BenchmarkFib20(b *testing.B) {
        var r uint64
        r = Fib(20)
        Result = r
}
```

## 1.7 Benchmark mistakes

The for loop is crucial to the operation of the benchmark.

Here are two incorrect benchmarks, can you explain what is wrong with them?

```go
func BenchmarkFibWrong(b *testing.B) {
	Fib(b.N) // b.N 
}

func BenchmarkFibWrong2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Fib(n) // n 
	}
}
```

Run these benchmarks, what do you see?


## 1.8 Benchmarking `math/rand`

Thanks to Spectre and Meltdown we all know that computers are very good a caching predictable operations.

Perhaps our Popcnt benchmark, even the correct version, is returning us a cached value — ​data that varies unpredictably might be slower than we’re expecting.

Let’s test this.

```go
var Result uint64

func BenchmarkPopcnt(b *testing.B) {
	var r uint64
	for i := 0; i < b.N; i++ {
		r = popcnt(rand.Uint64())
	}
	Result = r
}
```

```sh
go test -run=^$ -bench=PopcntRandSeed$ ./examples/popcnt
---
BenchmarkPopcntRand-12            73995884         15.10  ns/op
```

Is this result reliable? If not, what went wrong?

```go
func BenchmarkPopcntRandSeed(b *testing.B) {
	var r uint64
	for i := 0; i < b.N; i++ {
		rand.Seed(time.Now().UnixNano()) // seed
		r = popcnt(rand.Uint64())
	}
	Result = r
}
```

```sh
go test -run=^$ -bench=. ./examples/popcnt
---
BenchmarkPopcnt-12              1000000000          0.255 ns/op # inlined
BenchmarkPopcntNoInline-12       780286964          1.51  ns/op
BenchmarkPopcntRand-12            73995884         15.10  ns/op
BenchmarkPopcntRandSeed-12          149853       7853.00  ns/op # random seed
```

## 1.9 Profiling benchmarks

The testing package has **built in support** for generating CPU, memory, and block profiles.

- -cpuprofile=$FILE writes a CPU profile to $FILE.
- -memprofile=$FILE, writes a memory profile to $FILE
- -memprofilerate=N adjusts the profile rate to 1/N.
- -blockprofile=$FILE, writes a block profile to $FILE.

Using any of these flags also preserves the binary.

```sh
go test -run=^$ -bench=. -cpuprofile=cpu.profile bytes # running bytes package benchmarks
go tool pprof cpu.profile

go test -run=^$ -bench=. -memprofile=mem.profile bytes
go test -run=^$ -bench=. -memprofile=mem.profile -memprofilerate=5 bytes # N=1/5
go tool pprof mem.profile

go test -run=^$ -bench=. -blockprofile=bloc.profile bytes
go tool pprof bloc.profile
```
