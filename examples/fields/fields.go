package main

import (
	"fmt"
	"unsafe"
)

type S struct {
	a bool
	b float64
	c int32
}

/*
type S struct {
	a bool
	_ [7]byte // padding <1>
	b float64
	c int32
	_ [4]byte // padding <2>
}
*/

func main() {
	var s S
	fmt.Println(unsafe.Sizeof(s)) // <1>
}
