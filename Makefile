SHELL = /bin/bash -o pipefail

# The GOMAXPROCS variable limits the number of operating system threads
# that can execute user-level Go code simultaneously.
# https://golang.org/pkg/runtime/
#
# cat /proc/cpuinfo

fib-bench-all: # escape $ in make with $$
	go test -run=^$$ -bench=. ./examples/fib

fib-bench-fib20-cpu124: # run with (1, 2 and 4) OS-threads visible to the runtime - GOMAXPROCS
	go test -bench=Fib20 -cpu=1,2,4 ./examples/fib
