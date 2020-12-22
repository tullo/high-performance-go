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
