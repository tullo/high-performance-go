package fib

// Fib computes the n'th number in the Fibonacci series.
func Fib(n int) int {
	switch n {
	case 0:
		return 0
	case 1:
		return 1
	default:
		return Fib(n-1) + Fib(n-2)
	}
}

// Fib2 computes the n'th number in the Fibonacci series.
func Fib2(n int) int {
	switch n {
	case 0:
		return 0
	case 1:
		return 1
	case 2:
		return 1
	default:
		return Fib2(n-1) + Fib2(n-2)
	}
}

// Fib3 computes the n'th number in the Fibonacci series.
func Fib3(n int) int {
	a, b := 0, 1
	for i := 0; i < n; i++ {
		a, b = b, a+b
	}
	return a
}
