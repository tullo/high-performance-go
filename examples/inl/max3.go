// +build none

package main

func F() {
	const a, b = 100, 20
	var result int
	if true {
		result = a
	} else {
		result = b
	}
	if result == b {
		panic(b)
	}
}

func main() {
	F()
}
