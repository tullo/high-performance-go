package main

import (
	"sync"
	"testing"
)

type AtomicVariable struct {
	mu  sync.Mutex
	val uint64
}

func (av *AtomicVariable) Inc() {
	av.mu.Lock()
	av.val++
	av.mu.Unlock()
}

// Mutex contention increases with the number of goroutines.
func BenchmarkInc(b *testing.B) {
	var av AtomicVariable

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			av.Inc()
		}
	})
}

/*
go test -bench=. -cpu=1,2,4,8,16 ./examples/mutex/

BenchmarkInc       	89743338	        12.6 ns/op
BenchmarkInc-2     	72014619	        16.4 ns/op
BenchmarkInc-4     	30865150	        41.0 ns/op
BenchmarkInc-8     	20037595	        59.5 ns/op
BenchmarkInc-16    	18283101	        67.8 ns/op
PASS
ok  	high-performance-go/examples/mutex	6.209s
*/
