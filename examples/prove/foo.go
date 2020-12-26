package main

func foo(x int32) bool {
	if x > 5 { // at this point the compiler knows that x is greater than 5
		if x > 3 { // therefore x is also greater than 3, the branch is always taken.
			return true
		}
		panic("x less than 3")
	}
	return false
}

func main() {
	foo(-1)
}

// examples/prove/foo.go:5:10: Proved Less32
