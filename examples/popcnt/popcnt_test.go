package popcnt

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkPopcntDiscarded(b *testing.B) {
	for i := 0; i < b.N; i++ {
		popcnt(uint64(i)) // compiler inlines and discards leaf function
	}
}

func BenchmarkPopcntNotDiscarded(b *testing.B) {
	var r uint64
	for i := 0; i < b.N; i++ {
		r = popcnt(uint64(i))
	}
	Result = r // asignment to package public variable
}

func BenchmarkPopcntRand(b *testing.B) {
	var r uint64
	for i := 0; i < b.N; i++ {
		r = popcnt(rand.Uint64())
	}
	Result = r
}

func BenchmarkPopcntRandSeed(b *testing.B) {
	var r uint64
	for i := 0; i < b.N; i++ {
		rand.Seed(time.Now().UnixNano()) // seed
		r = popcnt(rand.Uint64())        // pseudo-random value
	}
	Result = r
}
