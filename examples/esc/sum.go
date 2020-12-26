package esc

import "fmt"

// Sum returns the sum of the numbers 1 to 100.
func Sum() int {
	const count = 100 // Values for "count => 8192" | 2^13 -> escape to heap
	numbers := make([]int, count)
	for i := range numbers {
		numbers[i] = i + 1
	}

	var sum int
	for _, i := range numbers {
		sum += i
	}
	return sum
}

func main() {
	answer := Sum()
	fmt.Println(answer)
}
