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
	// defer av.mu.Unlock() // cost: closure capture
	av.val++
	av.mu.Unlock()
}

// Mutex contention increases with the number of goroutines.
func BenchmarkInc(b *testing.B) {
	var av AtomicVariable

	b.RunParallel(func(pb *testing.PB) {
		// Each goroutine can use own resources here - buffers/etc.
		for pb.Next() {
			// The loop body is executed b.N times total across all goroutines.
			av.Inc()
		}
	})
}

/*
go test -bench=. -cpu=1,2,4,8,12 ./examples/mutex/

=> Mutex contention:

BenchmarkInc       	91627281	        12.6 ns/op
BenchmarkInc-2     	77747353	        15.0 ns/op
BenchmarkInc-4     	34756356	        39.3 ns/op
BenchmarkInc-8     	22753944	        55.5 ns/op
BenchmarkInc-12    	17719736	        63.1 ns/op
*/

/*
go test -bench=. ./examples/mutex/

BenchmarkInc-12    	19281381	        64.8 ns/op	=> av.mu.Unlock()
BenchmarkInc-12    	19105226	        67.8 ns/op	=> defer av.mu.Unlock()
*/
