package popcnt

const m1 = 0x5555555555555555
const m2 = 0x3333333333333333
const m4 = 0x0f0f0f0f0f0f0f0f
const h01 = 0x0101010101010101

func popcnt(x uint64) uint64 {
	x -= (x >> 1) & m1
	x = (x & m2) + ((x >> 2) & m2)
	x = (x + (x >> 4)) & m4
	return (x * h01) >> 56
}

// ========================================================
// This is the recommended way to ensure the compiler
// cannot optimise away the body of the loop.

// Result is public, so the compiler cannot prove that
// another package importing this one will not be able
// to see the value of Result changing over time, hence
// it cannot optimise away any of the operations leading
// to its assignment.
var Result uint64
