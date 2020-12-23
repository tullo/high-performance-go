SHELL = /bin/bash -o pipefail

# The GOMAXPROCS variable limits the number of operating system threads
# that can execute user-level Go code simultaneously.
# https://golang.org/pkg/runtime/
#
# cat /proc/cpuinfo
#
# escape $ in makefile with $$

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

# ========  ========================================

words:
#	go build -o words1 ./examples/words/main.go
	go build -o words2 ./examples/words/main.go
#	go build -o words3 ./examples/words/main3.go

time-words:
	@time ./words1 ./examples/words/moby.txt
	@time ./words2 ./examples/words/moby.txt
