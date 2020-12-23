
```
High Performance Go Workshop

1. Benchmarking
2. Performance measurement and profiling

    2.1. pprof
    2.2. Types of profiles
    2.3. One profile at at time
    2.4. Collecting a profile
    2.5. Analysing a profile with pprof
    2.6. Discussion

3. Compiler optimisations
4. Execution Tracer
5. Memory and Garbage Collector
6. Tips and trips

Dave Cheney dave@cheney.net (v379996b, 2019-07-24) 
```

# 2. Performance measurement and profiling

In the previous section we looked at benchmarking individual functions which is useful when you know ahead of time where the bottlekneck is.

However, often you will find yourself in the position of asking:

> Why is this program taking so long to run?

Profiling **whole** programs is useful for answering high level questions.

In this section we’ll use profiling tools built into Go to investigate the operation of the program from the inside.

----

## 2.1. [pprof](https://dave.cheney.net/high-performance-go-workshop/gophercon-2019.html#pprof)

 - `pprof` descends from the https://github.com/gperftools/gperftools
 - `pprof` is a tool for **visualization and analysis of profiling data**
 - https://github.com/google/pprof



`pprof` consists of **two parts**:
- `runtime/pprof` package **built into every Go program**
- `go tool pprof` for investigating profiles.

----

### 2.2.1. CPU profiling

CPU profiling is the most common type of profile, and the most obvious.

When CPU profiling is enabled the runtime will **interrupt itself every 10ms and record the stack trace of the currently running goroutines**.

Once the profile is complete we can analyse it to determine the hottest code paths.

The more times a **function appears** in the profile, the more time that code path is taking as a percentage of the total runtime.

----

### 2.2.2. Memory profiling

Memory profiling records the stack trace when a `heap allocation` is made.

**Stack allocations** are assumed to be free and are `not_tracked` in the memory profile.

`Memory` profiling, like `CPU` profiling is **sample based**,
by default memory profiling `samples 1 in every 1000 allocations`. This rate can be changed.

Personal Opinion: I do not find memory profiling useful for finding memory leaks.

> There are better ways to determine **how much memory** your application is using.

We will discuss these later in the presentation.

----

### 2.2.3. Block profiling

Block profiling is quite unique to Go.

A `block profile` records the amount of time a goroutine spent **waiting for a shared resource**.

This can be useful for determining **concurrency bottlenecks** in your application.

`Block profiling` can show you when a large number of **goroutines could make progress, but were blocked**.

Blocking includes:

- **Sending or receiving** on a `unbuffered channel`.
- **Sending** to a `full channel`.
- **Receiving** from an `empty channel`.
- **Trying to Lock** a `sync.Mutex` that is **locked by another goroutine**.

> Block profiling is a **very specialised tool**,
>
> it **should not be used until** you believe you have
>
> **eliminated all your CPU and memory usage bottlenecks**.

----

### 2.2.4. Mutex profiling

`Mutex profiling` is focused exclusively on **operations that lead to delays** caused by **mutex contention**.

Just like blocking profile, it says **how much time** was spent **waiting for a resource**.

Said another way, the `mutex profile` **reports how much time could been saved** if the lock contention was removed.

----

## 2.3. One profile at at time

Profiling is **not free**.

Profiling has a moderate, but measurable impact on programs performance — especially if you increase the **memory profile sample rate**.

Do not enable more than one kind of profile at a time.

> If you enable multiple profile’s at the same time, they will observe their own interactions and throw off your results.

----

## 2.4. Collecting a profile

The Go runtime’s profiling interface lives in the `runtime/pprof` package.

`runtime/pprof` is a very low level tool, and for historic reasons the interfaces to the different kinds of profile are not uniform.

As we saw in the previous section, pprof profiling is built into the testing package, but sometimes its **inconvenient, or difficult,
to place the code you want to profile in the context of at testing.B benchmark and must use the `runtime/pprof` API directly**.

A few years ago I wrote a small package, to make it **easier to profile an existing application**.

https://github.com/pkg/profile

```go
import "github.com/pkg/profile"

func main() {
	defer profile.Start().Stop()
	// ...
}
```

We'll use the profile package throughout this section.

Later in the day we’ll touch on using the runtime/pprof interface directly.

----

## 2.5. Analysing a profile with pprof

The analysis is driven by the `go pprof` subcommand:

`go tool pprof /path/to/your/profile`

This tool provides several different representations of the profiling data; textual, graphical, even flame graphs.

Since Go 1.9 the **profile file contains all the information** needed to render the profile.

> You do no longer need the binary which produced the profile. 🎉

----

### 2.5.1. Further reading

- Profiling Go Programs (2013) https://blog.golang.org/pprof
- Debugging performance issues in Go* programs https://software.intel.com/content/www/us/en/develop/blogs/debugging-performance-issues-in-go-programs.html
  - Note: the formats described here are based on Go1.3 release.

----

### 2.5.2. CPU profiling (exercise)

Let’s write a program to count words:

```go
func readbyte(r io.Reader) (rune, error) {
	var buf [1]byte
	_, err := r.Read(buf[:])
	return rune(buf[0]), err
}

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("could not open file %q: %v", os.Args[1], err)
	}

	words := 0
	inword := false
	for {
		r, err := readbyte(f)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("could not read file %q: %v", os.Args[1], err)
		}
		if unicode.IsSpace(r) && inword {
			words++
			inword = false
		}
		inword = unicode.IsLetter(r)
	}
	fmt.Printf("%q: %d words\n", os.Args[1], words)
}
```

Let’s see how many words there are in Herman Melville’s classic [Moby Dick](https://www.gutenberg.org/ebooks/2701) (sourced from Project Gutenberg)

```sh
go build -o words1 ./examples/words/main.go
time ./words1 ./examples/words/moby.txt
"./examples/words/moby.txt": 181275 words

real	0m0,439s
user	0m0,174s
sys	0m0,270s
```

```sh
time wc -w ./examples/words/moby.txt
215829 ./examples/words/moby.txt

real	0m0,017s
user	0m0,013s
sys	0m0,005s
```

So the numbers aren't the same. wc is about 19% higher because what it considers a word is different to what my simple program does. That’s not important — ​**both programs** take the whole file as input and in a single pass **count the number of transitions from word to non word**.

Let's investigate why these programs have different run times using `pprof`.

----

### 2.5.3. Add CPU profiling

```go
func readbyte(r io.Reader) (rune, error) {
	var buf [1]byte
	_, err := r.Read(buf[:])
	return rune(buf[0]), err
}

func main() {
    defer profile.Start(profile.MemProfile).Stop()
}
```

Now we have the profile we can analyse it with go tool pprof

```sh
go tool pprof ./cpu.pprof 
File: words
Type: cpu
Time: Dec 18, 2020 at 2:06pm (CET)
Duration: 200.84ms, Total samples = 20ms ( 9.96%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof)
(pprof) top
Showing nodes accounting for 20ms, 100% of 20ms total
      flat  flat%   sum%        cum   cum%
      10ms 50.00% 50.00%       10ms 50.00%  bufio.(*Reader).Read
      10ms 50.00%   100%       10ms 50.00%  runtime.mallocgc
         0     0%   100%       20ms   100%  main.main
         0     0%   100%       20ms   100%  main.readbyte (inline)
         0     0%   100%       20ms   100%  runtime.main
         0     0%   100%       10ms 50.00%  runtime.newobject

# However, in Go 1.10 Go ships with a version of pprof that natively supports a http sever.
# It will open a web browser;
#   Graph mode          http://localhost:8080/ui/?si=cpu
#   Flame graph mode    http://localhost:8080/ui/flamegraph?si=cpu
#   Top mode            http://localhost:8080/ui/top?si=cpu
go tool pprof -http=:8080 ./cpu.pprof
```

On the graph the box that consumes the most CPU time is the largest

— we see sys `call.Syscall` at 99.3% of the total time spent in the program.

The size of the arrow represents how much time was spent in children of a box, 

— we see that from `main.readbyte` onwards they account for near 0 of the 1.41 second spent in this arm of the graph.

Each call to readbyte results in a syscall.Read with a buffer size of 1. So the number of syscalls executed by our program is equal to the size of the input. We can see that in the pprof graph that reading the input dominates everything else.

```go
func readbyte(r io.Reader) (rune, error) {
	var buf [1]byte // allocation is here
	_, err := r.Read(buf[:])
	return rune(buf[0]), err
}

func main() {
	defer profile.Start(profile.MemProfile, profile.MemProfileRate(1)).Stop()
}
```

----

## 2.5.5. Memory profiling

The new words profile suggests that something is allocating inside the readbyte function. We can use pprof to investigate.

`defer profile.Start(profile.MemProfile).Stop()`

`go run words.go moby.txt`

`go tool pprof -http=:8080 /tmp/profile018586125/mem.pprof`

As we suspected the allocation was coming from readbyte — this wasn’t that complicated, readbyte is three lines long:

> Use pprof to determine where the allocation is coming from.

```go
func readbyte(r io.Reader) (rune, error) {
	var buf [1]byte // allocation is here
	_, err := r.Read(buf[:])
	return rune(buf[0]), err
}
```

What we see is every call to readbyte is allocating a new one byte long array and that array is being allocated on the heap.

> What are some ways we can avoid this? Try them and use CPU and memory profiling to prove it.

----

## 2.5.6. Alloc objects vs. inuse objects

Memory profiles come in two varieties, named after their go tool pprof flags
- -alloc_objects reports the call site where each allocation was made.
- -inuse_objects reports the call site where an allocation was made if it was reachable at the end of the profile.

To demonstrate this, here is a contrived program which will allocate a bunch of memory in a controlled manner.

```go
const count = 100000

var y []byte

func main() {
	defer profile.Start(profile.MemProfile, profile.MemProfileRate(1)).Stop()
	y = allocate()
	runtime.GC()
}

// allocate allocates count byte slices and returns the first slice allocated.
func allocate() []byte {
	var x [][]byte
	for i := 0; i < count; i++ {
		x = append(x, makeByteSlice())
	}
	return x[0]
}

// makeByteSlice returns a byte slice of a random length in the range [0, 16384).
func makeByteSlice() []byte {
	return make([]byte, rand.Intn(2^14))
}
```

The graph of `allocated objects` shows the call graphs that lead to the allocation of every object during the profile.

```sh
cd ../memory/
go run memory.go 
2020/12/18 15:27:07 profile: memory profiling enabled (rate 1), /tmp/profile925718186/mem.pprof
2020/12/18 15:27:07 profile: memory profiling disabled, /tmp/profile925718186/mem.pprof

go tool pprof -http=:8080 /tmp/profile925718186/mem.pprof
`http://localhost:8080/ui/?si=alloc_objects`
# Not surprisingly more than 99% of the allocations were inside makeByteSlice.

# Now lets look at the same profile using -inuse_objects
`http://localhost:8080/ui/?si=inuse_objects`
```

What we see is **not** the objects that were allocated during the profile, but the **objects that remain in use**, at the time the profile was taken — this ignores the stack trace for objects which have been reclaimed by the garbage collector.

----

## 2.5.7. Block profiling

The last profile type we’ll look at is block profiling. We’ll use the ClientServer benchmark from the `net/http` package

```sh
go test -run=XXX -bench=ClientServer$ -blockprofile=/tmp/block.p net/http

pkg: net/http
BenchmarkClientServer-12    	   22063	     54877 ns/op	    4996 B/op	      59 allocs/
```

---

### 2.5.8. Mutex profiling

Mutex contention increases with the number of goroutines.

```sh
go test -bench=. -cpu=1,2,4,8,16 ./examples/mutex/
goos: linux
goarch: amd64
pkg: high-performance-go-workshop/examples/mutex
BenchmarkInc       	91271556	        12.6 ns/op
BenchmarkInc-2     	77997690	        16.6 ns/op
BenchmarkInc-4     	31951504	        39.9 ns/op
BenchmarkInc-8     	21071281	        55.7 ns/op
BenchmarkInc-16    	17673577	        65.4 ns/op
```

----

### 2.5.9. Framepointers

The frame pointer is a register that always points to the top of the current stack frame.

Framepointers enable tools like gdb(1), and perf(1) to understand the Go call stack.

- [Seven ways to profile a Go program](https://talks.godoc.org/github.com/davecheney/presentations/seven.slide) (slides) 2016
- [Golang UK Conference 2016 - Dave Cheney - Seven ways to Profile Go Applications](https://www.youtube.com/watch?v=2h_NFBFrciI)


```sh
# The testing package has built in support for generating CPU, memory, and block profiles.
#   -cpuprofile=$FILE writes a CPU profile to $FILE. 
#   -memprofile=$FILE, writes a memory profile to $FILE, 
#   -memprofilerate=N adjusts the profile rate to 1/N.
#   -blockprofile=$FILE writes a block profile to $FILE.
go test -run=XXX -bench=IndexByte -cpuprofile=/tmp/c.p bytes
go tool pprof bytes.test /tmp/c.p

#==============================================================================
# We can visualise memory profiles in the same way.
go build -gcflags='-memprofile=/tmp/m.p'
go tool pprof --alloc_objects -svg $(go tool -n compile) /tmp/m.p > alloc_objects.svg
go tool pprof --inuse_objects -svg $(go tool -n compile) /tmp/m.p > inuse_objects.svg
# go tool [-n] command [args...]
...
compile
doc
pprof
...


# A simple way to obtain a general idea of how hard the garbage collector is working
# is to enable the output of GC logging. 
env GODEBUG=gctrace=1 godoc -http=:8080

# native time => more poweful than bash builtin
/usr/bin/time -v go fmt std

Command exited with non-zero status 1
	Command being timed: "go fmt std"
	User time (seconds): 11.25
	System time (seconds): 1.52
	Percent of CPU this job got: 279%
	Elapsed (wall clock) time (h:mm:ss or m:ss): 0:04.57
	Average shared text size (kbytes): 0
	Average unshared data size (kbytes): 0
	Average stack size (kbytes): 0
	Average total size (kbytes): 0
	Maximum resident set size (kbytes): 58448
	Average resident set size (kbytes): 0
	Major (requiring I/O) page faults: 0
	Minor (reclaiming a frame) page faults: 983168
	Voluntary context switches: 93906
	Involuntary context switches: 13751
	Swaps: 0
	File system inputs: 1318
	File system outputs: 0
	Socket messages sent: 0
	Socket messages received: 0
	Signals delivered: 0
	Page size (bytes): 4096
	Exit status: 1

#==============================================================================
# /debug/pprof
#
#  If your program runs a webserver you can enable debugging over http.
import _ "net/http/pprof"

func main() {
    log.Println(http.ListenAndServe("localhost:3999", nil))
}
# Then use the pprof tool to look at a 30-second CPU profile:
go tool pprof http://localhost:3999/debug/pprof/profile
# Or to look at the heap profile:
go tool pprof http://localhost:3999/debug/pprof/heap
#  Or to look at the goroutine blocking profile: 
go tool pprof http://localhost:3999/debug/pprof/block

# pprof should always be invoked with two arguments.
go tool pprof /path/to/your/binary /path/to/your/profile
# The binary argument must be the binary that produced this profile.
# The profile argument must be the profile generated by this binary.
go tool pprof $BINARY /tmp/c.p # det passer vist ikke mere
(pprof) top
(pprof) top10
(pprof) web

go tool pprof --help
...
pprof <format> [options] [binary] <source> ...
pprof [options] [binary] <source> ...
pprof -http [host]:[port] [options] [binary] <source> ...
...
-top             Outputs top entries in text form
-topN
-traces          Outputs all profile samples in text form
-tree            Outputs a text rendering of call graph
-web             Visualize graph through web browser
-weblist         Display annotated source in a web browser
...
-cum             Sort entries based on cumulative weight
-flat            Sort entries based on own weight
...
-seconds         Duration for time-based profile collection
-timeout         Timeout in seconds for profile collection
...
Legacy convenience options:
-inuse_space           Same as -sample_index=inuse_space
-inuse_objects         Same as -sample_index=inuse_objects
-alloc_space           Same as -sample_index=alloc_space
-alloc_objects         Same as -sample_index=alloc_objects
-total_delay           Same as -sample_index=delay
-contentions           Same as -sample_index=contentions
-mean_delay            Same as -mean -sample_index=delay
...
Environment Variables:
PPROF_TMPDIR       (default $HOME/pprof)
PPROF_BINARY_PATH  default: $HOME/pprof/binaries

(pprof) top10
Total: 2525 samples
#    (1)  (2)    (3)        (4)  (5)
     298  11.8%  11.8%      345  13.7% runtime.mapaccess1_fast64
     268  10.6%  22.4%     2124  84.1% main.FindLoops
     251   9.9%  32.4%      451  17.9% scanblock
# (1) number of samples in which the function was running [raw count] 
# (2) number of samples in which the function was running [percentage of total samples]
# (3) running total during the listing
# (4) number of samples in which the function appeared (running | waiting for a called function to return) [raw count]
# (5) number of samples in which the function appeared (running | waiting for a called function to return) [percentage of total samples]
# (1+2) The `runtime.mapaccess1_fast64` function was running during 298 samples, or 11.8%.
# (3)   The first three rows account for 32.4% of the samples.
# (4+5) The `main.FindLoops` function was running in 10.6% of the samples, 
#       but it was on the call stack (it or functions it called were running)
#       in 84.1% of the samples.
# When CPU profiling is enabled, the Go program stops about 100 times per second and records a sample consisting of the program counters on the currently executing goroutine's stack.
#
# To sort by the (4) and (5) columns, use the -cum (for cumulative) flag:
(pprof) top5 -cum
Total: 2525 samples
#     (1)  (2)    (3)       (4)  (5)
       0   0.0%   0.0%     2144  84.9% gosched0
       0   0.0%   0.0%     2144  84.9% main.main
       0   0.0%   0.0%     2144  84.9% runtime.main
       0   0.0%   0.0%     2124  84.1% main.FindHavlakLoops
     268  10.6%  10.6%     2124  84.1% main.FindLoops
(pprof) web
(pprof) web mapaccess1  # use only samples that include a specific function [runtime.mapaccess1_fast64]
(pprof) list DFS        # zoom in on a particular function [main.DFS]
Total: 2525 samples
     7    354  247:                     lastid = DFS(target, nodes, number, last, lastid+1)
# Since we already know that the time is going into map lookups implemented by the hash runtime functions, we care most about the second column.
# A large fraction of time is spent in recursive calls to DFS (line 247)
# It looks like the time is going into the accesses to the number map
# We can use a []int, a slice indexed by the block number.
# There's no reason to use a map when an array or slice will do.
# Changing number from a map to a slice requires editing seven lines in the program
# and cut its run time by nearly a factor of two:
# https://github.com/rsc/benchgraffiti/commit/58ac27bcac3ffb553c29d0b3fb64745c91c95948
go tool pprof havlak2 havlak2.prof
(pprof) top5
Total: 1652 samples
     197  11.9%  11.9%      382  23.1% scanblock
     189  11.4%  23.4%     1549  93.8% main.FindLoops
     130   7.9%  31.2%      152   9.2% sweepspan
#    104   6.3%  37.5%      896  54.2% runtime.mallocgc
      98   5.9%  43.5%      100   6.1% flushptrbuf
# We can confirm that main.DFS is no longer a significant part of the run time.
# Now the program is spending most of its time allocating memory and garbage collecting
# (runtime.mallocgc, which both allocates and runs periodic garbage collections, 
# accounts for 54.2% of the time).
# To find out why the garbage collector is running so much, 
# we have to find out what is allocating memory.
# One way is to add memory profiling to the program.

# Now the samples are memory allocations, not clock ticks.
go tool pprof havlak3 havlak3.mprof
(pprof) top5
Total: 82.4 MB
    56.3  68.4%  68.4%     56.3  68.4% main.FindLoops
    17.6  21.3%  89.7%     17.6  21.3% main.(*CFG).CreateNode
     8.0   9.7%  99.4%     25.6  31.0% main.NewBasicBlockEdge
     0.5   0.6% 100.0%      0.5   0.6% itab
     0.0   0.0% 100.0%      0.5   0.6% fmt.init
# FindLoops has allocated approximately 56.3 of the 82.4 MB in use.
# To find the memory allocations, we can list those functions.
(pprof) list FindLoops
# The current bottleneck is the same as the last one
# FindLoops is allocating about 29.5 MB of maps.
# make([][]int, size) https://github.com/rsc/benchgraffiti/commit/245d899f7b1a33b0c8148a4cd147cb3de5228c8a
# ANOTHER VIEW: go tool pprof --inuse_objects havlak3 havlak3.mprof
#
# We're now at 2.11x faster than when we started. Let's look at a CPU profile again.
go tool pprof havlak4 havlak4.prof
(pprof) top10
Total: 1173 samples
     205  17.5%  17.5%     1083  92.3% main.FindLoops
     138  11.8%  29.2%      215  18.3% scanblock
      88   7.5%  36.7%       96   8.2% sweepspan
      76   6.5%  43.2%      597  50.9% runtime.mallocgc
      75   6.4%  49.6%       78   6.6% runtime.settype_flush
      74   6.3%  55.9%       75   6.4% flushptrbuf
      64   5.5%  61.4%       64   5.5% runtime.memmove
      63   5.4%  66.8%      524  44.7% runtime.growslice
      51   4.3%  71.1%       51   4.3% main.DFS
      50   4.3%  75.4%      146  12.4% runtime.MCache_Alloc
(pprof)
# Now memory allocation and the consequent garbage collection 
# (runtime.mallocgc) accounts for 50.9% of our run time.
# Another way to look at why the system is garbage collecting 
# is to look at the allocations that are causing the collections, 
# the ones that spend most of the time in mallocgc:
(pprof) web mallocgc
# It's hard to tell what's going on in that graph, 
# because there are many nodes with small sample numbers obscuring the big ones.
# We can tell go tool pprof to ignore nodes that don't account for 
# at least 10% of the samples:
go tool pprof --nodefraction=0.1 havlak4 havlak4.prof
(pprof) web mallocgc
# We can follow the thick arrows easily now, to see that FindLoops is triggering most of the garbage collection.
# If we list FindLoops we can see that much of it is right at the beginning:
(pprof) list FindLoops
...
# Every time FindLoops is called, it allocates some sizable bookkeeping structures.
# Since the benchmark calls FindLoops 50 times, 
# these add up to a significant amount of garbage,
# so a significant amount of work for the garbage collector.

# Having a garbage-collected language doesn't mean you can ignore memory allocation issues.
# In this case, a simple solution is to introduce a cache 
# so that each call to FindLoops reuses the previous call's storage when possible.
# We'll add a global cache structure:
# https://github.com/rsc/benchgraffiti/commit/2d41d6d16286b8146a3f697dd4074deac60d12a4
# Such a global variable is bad engineering practice, of course: 
# it means that concurrent calls to FindLoops are now unsafe.
# For now, we are making the minimal possible changes in order to understand
# what is important for the performance of our program; 
# The final version of the Go program will use a separate LoopFinder instance to track this memory, restoring the possibility of concurrent use.
go build havlak5.go
# There's more we can do to clean up the program and make it faster,
# but none of it requires profiling techniques that we haven't already shown.

# The final version is written using idiomatic Go style, using data structures and methods.
# The final version runs in 2.29 seconds and uses 351 MB of memory:
make havlak6
go build havlak6.go
./xtime ./havlak6
2.26u 0.02s 2.29r 360224kB
# https://github.com/rsc/benchgraffiti/blob/master/havlak/havlak6.go
# https://github.com/rsc/benchgraffiti/blob/master/havlak/havlak6.cc

go tool pprof http://localhost:6060/debug/pprof/profile   # 30-second CPU profile
go tool pprof http://localhost:6060/debug/pprof/heap      # heap profile
go tool pprof http://localhost:6060/debug/pprof/block     # goroutine blocking profile



# =============================================================================
# perf
# 
# Now we have frame pointers, perf can profile Go applications.
sudo go build -toolexec="perf stat" cmd/compile/internal/gc

# =============================================================================
# perf stat
sudo perf stat sleep 1
sudo go build -toolexec="perf stat" cmd/compile/internal/gc

# Performance counter stats for '/snap/go/6745/pkg/tool/linux_amd64/compile ..... :

#           4.337,07 msec task-clock                #    3,433 CPUs utilized          
#              5.336      context-switches          #    0,001 M/sec                  
#                 56      cpu-migrations            #    0,013 K/sec                  
#             90.416      page-faults               #    0,021 M/sec                  
#     16.827.969.773      cycles                    #    3,880 GHz                    
#     24.631.461.225      instructions              #    1,46  insn per cycle         
#      5.236.084.622      branches                  # 1207,287 M/sec                  
#        114.230.895      branch-misses             #    2,18% of all branches        

#        1,263451576 seconds time elapsed

#        4,262601000 seconds user
#        0,120412000 seconds sys


# =============================================================================
# perf record
sudo go build -toolexec="perf record -g -o /tmp/perf.data" cmd/compile/internal/gc
sudo perf report -i /tmp/perf.data

sudo strace -f -e open perf report -i /tmp/perf.data 2>&1 | grep tips.txt
sudo strace -f -e file perf report -i /tmp/perf.data 2>&1 | grep tips.txt
access("/usr/share/doc/perf-tip/tips.txt", F_OK) = -1 ENOENT (No such file or directory)
access("/build/linux-iUcj8A/linux-5.8.0/debian/build/tools-perarch/tools/perf/Documentation/tips.txt", F_OK) = -1 ENOENT (No such file or directory)
# (Cannot load tips.txt file, please install perf!)



# =============================================================================
# Flame graph

=> x axis: alphabetical stack sort, to maximise merging.
=> y axis: stack depth.

# Each rectangle represents a stack frame.
# The wider a frame is is, the more often it was present in the stacks.
# The top edge shows what is on-CPU, and beneath it is its ancestry.
# The colors are usually not significant, picked randomly to differentiate frames.
#
# https://www.slideshare.net/brendangregg/java-performance-analysis-on-linux-with-flame-graphs
#
# Flame graphs can consume data from many sources, including pprof (and perf(1)).

go build -gcflags=-cpuprofile=/tmp/c.p .
go tool pprof -http=:8080 /tmp/c.p

# =============================================================================
# go tool trace
#
# Gives insight into dynamic execution of a program.
#
# Captures with nanosecond precision:
#
#     goroutine creation/start/end
#     goroutine blocking/unblocking
#     network blocking
#     system calls
#     GC events
# https://www.dotconferences.com/2016/10/rhys-hiltner-go-execution-tracer
# https://making.pusher.com/go-tool-trace/
# https://github.com/campoy/go-tooling-workshop/blob/master/3-dynamic-analysis/4-tracing/1-tracing.md
# https://github.com/guevara/read-it-later/issues/4515

go tool trace -help
go test -trace=trace.out path/to/package
go tool trace [flags] pkg.test trace.out

# Installs a handler under the /debug/pprof/trace URL to download a live trace: 
import _ "net/http/pprof"

# https://golang.org/src/runtime/trace/trace.go
trace.WithRegion(ctx, "makeCappuccino", func() {

   // orderID allows to identify a specific order
   // among many cappuccino order region records.
   trace.Log(ctx, "orderID", orderID)

   trace.WithRegion(ctx, "steamMilk", steamMilk)
   trace.WithRegion(ctx, "extractCoffee", extractCoffee)
   trace.WithRegion(ctx, "mixMilkCoffee", mixMilkCoffee)
})

go build -gcflags=-traceprofile=/tmp/t.p cmd/compile/internal/gc
go tool trace /tmp/t.p
# open in chrome or firefox
```
