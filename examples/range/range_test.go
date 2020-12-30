package main

import (
	"testing"
)

var X [1 << 15]struct {
	val int
	_   [4096]byte
}

var Result int

func BenchmarkRangeValue(b *testing.B) {
	var r int
	for n := 0; n < b.N; n++ {
		for _, x := range X {
			r += x.val
		}
	}
	Result = r
}

func BenchmarkRangeIndex(b *testing.B) {
	var r int
	for n := 0; n < b.N; n++ {
		for i := range X {
			x := &X[i]
			r += x.val
		}
	}
	Result = r
}

func BenchmarkFor(b *testing.B) {
	var r int
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(X); i++ {
			x := &X[i]
			r += x.val
		}
	}
	Result = r
}

/*
go test -run=^$ -bench=. ./examples/range

BenchmarkRangeValue-12            75	  15603563 ns/op
BenchmarkRangeIndex-12         10497	    115826 ns/op
BenchmarkFor-12                10422	    114531 ns/op
*/
