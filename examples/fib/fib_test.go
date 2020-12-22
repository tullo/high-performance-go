package fib

import "testing"

func BenchmarkFib1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Fib(1) // 1.52ns (op) to complete and return
	}
}

func BenchmarkFib20(b *testing.B) {
	for n := 0; n < b.N; n++ {
		// Fib(20)  // 35318ns (op) to complete recursive calls and return
		// Fib2(20) // 21978ns
		Fib3(20) // 9.86ns
	}
}

func TestFib(t *testing.T) {
	fibs := []int{0, 1, 1, 2, 3, 5, 8, 13, 21}
	for n, want := range fibs {
		got := Fib(n)
		if want != got {
			t.Errorf("Fib(%d): want %d, got %d", n, want, got)
		}
	}
}

func TestFib2(t *testing.T) {
	fibs := []int{0, 1, 1, 2, 3, 5, 8, 13, 21}
	for n, want := range fibs {
		got := Fib2(n)
		if want != got {
			t.Errorf("Fib2(%d): want %d, got %d", n, want, got)
		}
	}
}

func TestFib3(t *testing.T) {
	fibs := []int{0, 1, 1, 2, 3, 5, 8, 13, 21}
	for n, want := range fibs {
		got := Fib3(n)
		if want != got {
			t.Errorf("Fib2(%d): want %d, got %d", n, want, got)
		}
	}
}

func TestFib20(t *testing.T) {
	fibs := [20]int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377, 610, 987, 1597, 2584, 4181}
	want := fibs[18] + fibs[19] //6765
	got := Fib3(20)
	if want != got {
		t.Errorf("Fib3(%d): want %d, got %d", 20, want, got)
	}

}
