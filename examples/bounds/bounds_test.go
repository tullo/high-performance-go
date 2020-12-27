package main

import "testing"

var v = make([]int, 9)

var A, B, C, D, E, F, G, H, I int

func BenchmarkBoundsCheckInOrder(b *testing.B) {
	var a, _b, c, d, e, f, g, h, i int
	for n := 0; n < b.N; n++ {
		a = v[0]
		_b = v[1]
		c = v[2]
		d = v[3]
		e = v[4]
		f = v[5]
		g = v[6]
		h = v[7]
		i = v[8]
	}
	A, B, C, D, E, F, G, H, I = a, _b, c, d, e, f, g, h, i
}

func BenchmarkBoundsCheckOutOfOrder(b *testing.B) {
	var a, _b, c, d, e, f, g, h, i int
	for n := 0; n < b.N; n++ {
		i = v[8]
		a = v[0]
		_b = v[1]
		c = v[2]
		d = v[3]
		e = v[4]
		f = v[5]
		g = v[6]
		h = v[7]
	}
	A, B, C, D, E, F, G, H, I = a, _b, c, d, e, f, g, h, i
}

/*
BenchmarkBoundsCheckInOrder:
	grep -A 99 "BenchmarkBoundsCheckInOrder(SB)" bounds-check.txt | grep "runtime.panicIndex(SB)"

	0x0150 00336 (bounds_test.go:20)	CALL	runtime.panicIndex(SB)
	0x015a 00346 (bounds_test.go:19)	CALL	runtime.panicIndex(SB)
	0x0164 00356 (bounds_test.go:18)	CALL	runtime.panicIndex(SB)
	0x016e 00366 (bounds_test.go:17)	CALL	runtime.panicIndex(SB)
	0x0178 00376 (bounds_test.go:16)	CALL	runtime.panicIndex(SB)
	0x0182 00386 (bounds_test.go:15)	CALL	runtime.panicIndex(SB)
	0x018c 00396 (bounds_test.go:14)	CALL	runtime.panicIndex(SB)
	0x0196 00406 (bounds_test.go:13)	CALL	runtime.panicIndex(SB)
	0x01a0 00416 (bounds_test.go:12)	CALL	runtime.panicIndex(SB)
*/

/*
BenchmarkBoundsCheckOutOfOrder:
	grep -A 50 "BenchmarkBoundsCheckOutOfOrder(SB)" bounds-check.txt | grep "runtime.panicIndex(SB)"

	0x00c1 00193 (bounds_test.go:28)	CALL	runtime.panicIndex(SB)
*/
