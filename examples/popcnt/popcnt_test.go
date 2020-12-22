package popcnt

import "testing"

func BenchmarkPopcntInline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		popcnt(uint64(i))
	}
}

func BenchmarkPopcntNoInline(b *testing.B) {
	var r uint64
	for i := 0; i < b.N; i++ {
		r = popcnt(uint64(i))
	}
	Result = r
}
