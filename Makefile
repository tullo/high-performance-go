SHELL = /bin/bash -o pipefail

fib-bench-all: # escape $ in make with $$
	go test -run=^$$ -bench=. ./examples/fib/
