SHELL = /bin/bash -o pipefail

# The GOMAXPROCS variable limits the number of operating system threads
# that can execute user-level Go code simultaneously.
# https://golang.org/pkg/runtime/
#
# cat /proc/cpuinfo
#
# escape $ in makefile with $$

# =============================================================================
# === BENCHMARKS ==============================================================
# =============================================================================

fib-bench-all: # run all benchmarks
	go test -run=^$$ -bench=. ./examples/fib

fib-bench-fib20-cpu124: # run with (1, 2 and 4) OS-threads visible to the runtime - GOMAXPROCS
	go test -run=^$$ -bench=Fib20 -cpu=1,2,4 ./examples/fib

fib-bench-fib20-10sec: # run until b.N takes longer than 10 seconds to return | -benchtime
	go test -run=^$$ -bench=Fib20 -benchtime=10s ./examples/fib

fib-bench-fib1-run10x: # run benchmarks multiple times | -count
	go test -run=^$$ -bench=Fib1 -count=10 ./examples/fib

fib-bench-fib1-iterations: # run your code exactly benchtime times | -benchtime
	go test -run=^$$ -bench=Fib1 -benchtime=10x ./examples/fib
	go test -run=^$$ -bench=Fib1 -benchtime=20x ./examples/fib
	go test -run=^$$ -bench=Fib1 -benchtime=50x ./examples/fib
	go test -run=^$$ -bench=Fib1 -benchtime=100x ./examples/fib
	go test -run=^$$ -bench=Fib1 -benchtime=300x ./examples/fib

fib-bench-fib20-benchstat: # run benchmarks 10x and compare with benchstat 
	go test -run=^$$ -bench=Fib20 -count=10 ./examples/fib/ | tee fib1.txt
	$$(go env GOPATH)/bin/benchstat fib1.txt
	# save the test binary
	go test -c ./examples/fib/
	@mv fib.test fib.golden
	@ls -lh | awk '{print $$5,$$9}' | grep 'fib.'

fib-bench-fib20-benchstat-comp: # run benchmarks 10x and compare with benchstat 
	# save the test binary
	go test -c ./examples/fib
	./fib.golden -test.bench=Fib20 -test.count=10 > fib1.txt
	./fib.test -test.bench=Fib20 -test.count=10 > fib2.txt
	$$(go env GOPATH)/bin/benchstat fib1.txt fib2.txt
	@mv fib.test fib.fib2

fib-bench-fib20-benchstat-comp3: # run benchmarks 10x and compare with benchstat
	# save the test binary
	go test -c ./examples/fib
	./fib.golden -test.bench=Fib20 -test.count=10 > fib1.txt
	./fib.fib2 -test.bench=Fib20 -test.count=10 > fib2.txt
	./fib.test -test.bench=Fib20 -test.count=10 > fib3.txt
	$$(go env GOPATH)/bin/benchstat fib1.txt fib2.txt
	$$(go env GOPATH)/bin/benchstat fib1.txt fib3.txt
	$$(go env GOPATH)/bin/benchstat fib2.txt fib3.txt
	$$(go env GOPATH)/bin/benchstat fib1.txt fib2.txt fib3.txt

benchmarks-report-allocs:	# reports allocs for benchmarks using ==> b.ReportAllocs()
	go test -run=^$$ -bench=. bufio

benchmarks-report-benchmem:	# reports allocs for ==> ALL benchmarks
	go test -run=^$$ -bench=. -benchmem bufio

compiler-optimisation-inline:	# compiler inlines leaf function
# 	-gcflags=-S      # -S assembly output
# 	-gcflags="-l -S" # -l disable inlining
	@go test -run=^$$ -gcflags="-S -m=2 -d=ssa/prove/debug=on" -bench=PopcntDiscarded ./examples/popcnt 2>popcnt.txt
#	@grep -v PCDATA popcnt.txt | grep -v FUNCDATA | grep -C 3 "inlining call to popcnt"
	@echo
	@echo "1. compiler inlined popcnt function body"
	@echo "2. compiler discarded func body"
	@echo
	@grep -A 9 "BenchmarkPopcntDiscarded(SB)" popcnt.txt

compiler-optimisation-defense:
#	go build -gcflags="-S -m=2 -d=ssa/prove/debug=on" ./examples/popcnt 2>popcnt.txt
	go test -run=^$$ -gcflags="-S -m=2 -d=ssa/prove/debug=on" -bench=PopcntNotDiscarded ./examples/popcnt 2>popcnt.txt
#	grep -v PCDATA popcnt.txt | grep -v FUNCDATA | grep -C 3 "inlining call to popcnt"
	@echo
	@echo "1. compiler inlined popcnt function body"
	@echo "19 assembly instructions"
	@echo
	@grep -A 32 "BenchmarkPopcntNotDiscarded(SB)" popcnt.txt

benchmarking-with-random-data:
	go test -run=^$$ -bench=PopcntRand ./examples/popcnt

# =============================================================================
# === TRACES ==================================================================
# =============================================================================

words:
#	go build -o words1 ./examples/words/main.go
#	go build -o words2 ./examples/words/main.go
	go build -o words3 ./examples/words/main.go

time-words:
#	@time ./words1 ./examples/words/moby.txt
#	@time ./words2 ./examples/words/moby.txt
#	@time ./words3 ./examples/words/moby.txt
	time wc -w ./examples/words/moby.txt
	cd examples/words/; time go run main.go moby.txt

inuse-allocs:
	go run examples/inuseallocs/main.go

block-profiling:
	cd examples/block ; go build && ./block
	go tool pprof -http=:8080 ./examples/block/block.pprof
#	go test -run=XXX -bench=ClientServer$$ -blockprofile=/tmp/block.p net/http
#	go tool pprof -http=:8080 /tmp/block.p



mutex-profiling:
	go test -bench=. -cpu=1,2,4,8,16 ./examples/mutex

bench-startstop-reset:
	go test -bench=. -benchtime=100000x -cpuprofile=c.p ./examples/benchstartstop

# =============================================================================
# === COMPILER OPTIMISATION ===================================================
# =============================================================================

escape-analysis-sum:	# -m = escape analysis decisions
	go build -gcflags=-m ./examples/esc/sum.go
#	examples/esc/sum.go:22:13: inlining call to fmt.Println
#	examples/esc/sum.go:8:17: make([]int, count) does not escape
#	examples/esc/sum.go:22:13: answer escapes to heap
#	examples/esc/sum.go:22:13: []interface {} literal does not escape
	@echo
	go build -gcflags='-m -m' ./examples/esc/sum.go 2>&1 | grep sum.go:22
	@echo
	go build -gcflags=-m ./examples/esc/center.go
	@echo
	go test -run=none -bench=Sum ./examples/esc/

leaf-func-inlining: 	# -m = escape analysis decisions
	go build -gcflags=-m=2 examples/inl/max.go
	@echo
	@echo
	go build -gcflags=-S examples/inl/max.go 2>&1 | grep -A5 '"".F STEXT'
	@echo
	@echo
	bash asm.sh ./examples/inl/max.go

prove-pass:				# -gcflags=-d=ssa/prove/debug=on
	go build -gcflags=-d=ssa/prove/debug=on examples/prove/foo.go
#	examples/prove/foo.go:5:10: Proved Less32

intrinsic-functions:	# code replacement with architecture native instructions
	go test -bench=.  ./examples/popcnt-intrinsic/
	@echo
	bash asm.sh ./examples/counter/counter.go

bounds-check-elimination:	# arrays & slices
	go test -gcflags=-S -bench=BoundsCheck  ./examples/bounds/bounds_test.go 2>bounds-check.txt
	@echo
	@echo "BenchmarkBoundsCheckInOrder:"
	@grep -v PCDATA bounds-check.txt | grep -v FUNCDATA \
		| grep -A 99 "BenchmarkBoundsCheckInOrder(SB)" | grep "runtime.panicIndex(SB)"
	@echo
	@echo "BenchmarkBoundsCheckOutOfOrder:"
	@grep -v PCDATA bounds-check.txt | grep -v FUNCDATA \
		| grep -A 50 "BenchmarkBoundsCheckOutOfOrder(SB)" | grep "runtime.panicIndex(SB)"

# =============================================================================
# === EXECUTION TRACERS =======================================================
# =============================================================================

mandelbrot:
	cd examples/mandelbrot ; go build && ./mandelbrot

mandelbrot-timer:
	cd examples/mandelbrot ; time ./mandelbrot

mandelbrot-runtime-pprof:
	cd examples/mandelbrot-runtime-pprof ; go run mandelbrot.go > cpu.pprof

mandelbrot-pkg-profile:
	cd examples/mandelbrot-pkg-profile ; go run mandelbrot.go
	go tool pprof -http=:8080 ./examples/mandelbrot-pkg-profile/cpu.pprof

mandelbrot-trace:				# sequential execution
	cd examples/mandelbrot-trace ; go build
	cd examples/mandelbrot-trace ; time ./mandelbrot-trace
	go tool trace ./examples/mandelbrot-trace/trace.out
#	47K trace.out				Gs: 1

mandelbrot-trace-mode-px:		# parallel execution
	cd examples/mandelbrot-trace ; go build
	cd examples/mandelbrot-trace ; time ./mandelbrot-trace -mode px
	go tool trace ./examples/mandelbrot-trace/trace.out
#	40M trace.out !				Gs: 1<<20 (1024×1024)

mandelbrot-trace-mode-row:		# parallel execution
	cd examples/mandelbrot-trace ; go build
	cd examples/mandelbrot-trace ; time ./mandelbrot-trace -mode row
	go tool trace ./examples/mandelbrot-trace/trace.out
#	64K trace.out				Gs: 1<<10 (1024)

mandelbrot-trace-mode-workers:	# parallel execution
	cd examples/mandelbrot-trace ; go build
	cd examples/mandelbrot-trace ; time ./mandelbrot-trace -mode workers -workers 4
	go tool trace ./examples/mandelbrot-trace/trace.out
#	48K trace.out				Gs: 4, channel buffer size: 1<<10 (1024)

mandelbrot-buffered-mode-workers:	# parallel execution
	cd examples/mandelbrot-buffered ; go build
	cd examples/mandelbrot-buffered ; time ./mandelbrot-buffered -mode workers -workers 4
	go tool trace ./examples/mandelbrot-buffered/trace.out
#	59K trace.out				Gs: 4, channel buffer size: 1<<20 (1024×1024)

mandelbrot-buffered-mode-workers-per-row:	# parallel execution
	cd examples/mandelbrot-buffered/exercise ; go build
	cd examples/mandelbrot-buffered/exercise ; time ./exercise -mode workers -workers 4
	go tool trace ./examples/mandelbrot-buffered/exercise/trace.out
#	48K trace.out				Gs: 4, channel buffer size: 1<<10 (1024)

mandelweb:
	go run examples/mandelweb/mandelweb.go

mandelweb-five-second-trace:	# grab a five second trace from mandelweb
	curl -o examples/mandelweb/trace.out http://127.0.0.1:8080/debug/pprof/trace?seconds=5
	go tool trace ./examples/mandelweb/trace.out

mandelweb-load-generator:
	@echo "Let's start with one request per second | 1 worker | 1000 requests."
	$$(go env GOPATH)/bin/hey -c 1 -n 1000 -q 1 http://127.0.0.1:8080/mandelbrot
#	GO111MODULE=on go get -u github.com/rakyll/hey

mandelweb-overload-simulator:
	@echo "Let's increase the rate to 5 requests per second | 5 workers | 1000 requests."
	$$(go env GOPATH)/bin/hey -c 5 -n 1000 -q 5 http://127.0.0.1:8080/mandelbrot

concurrent-prime-sieve:
	cd examples/sieve ; go build
	cd examples/sieve ; time ./sieve
	go tool trace ./examples/sieve/trace.out


# =============================================================================
# === MEMORY and GARBAGE COLLECTOR ============================================
# =============================================================================

benchmap-key-conversion:		# using []byte as a map key
	go test -run=^$$ -bench=. -benchmem ./examples/benchmap/

bytes-equality-testing:			# []byte to string conversions
	go test -run=^$$ -bench=. -benchmem ./examples/byteseq/

string-concatenation:			# avoid string concatenation
	go test -run=^$$ -bench=. -benchmem ./examples/concat/

slice-grow-with-append:			# append() is convenient, but wasteful
	go run ./examples/grow

padding-and-alignment:			# fields padding and alignment
	go run ./examples/fields/

range-vs-for:					# range versus for loop
	go test -gcflags=-S -run=^$$ -bench=. -benchmem ./examples/range 2>bounds-check.txt
	@echo "BenchmarkRangeValue:"
	@grep -v PCDATA bounds-check.txt | grep -v FUNCDATA | grep -A 58 "BenchmarkRangeValue(SB)"
	@echo
	@echo "BenchmarkRangeIndex:"
	@grep -v PCDATA bounds-check.txt | grep -v FUNCDATA | grep -A 22 "BenchmarkRangeIndex(SB)"
	@echo
	@echo "BenchmarkFor:"
	@grep -v PCDATA bounds-check.txt | grep -v FUNCDATA | grep -A 22 "BenchmarkFor(SB)"
