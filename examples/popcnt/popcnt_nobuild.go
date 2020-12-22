// +build none

package popcnt

import "testing"

func BenchmarkPopcnt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// optimised away
	}
}
