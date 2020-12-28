// Adapted from https://golang.org/doc/play/sieve.go
// Copywrite the Go authors, 2009

// A concurrent prime sieve
package main

import (
	"fmt"

	"github.com/pkg/profile"
)

/*
	The generate goroutine emits integers, starting from 2,
	each new goroutine filters out only specific prime multiples (2, 3, 5, 7...),
	sending first found prime to main.

	All numbers being sent from goroutines to main are prime numbers.
*/

// Generate emits integers, starting from 2
func Generate(ch chan<- int) {
	for i := 2; ; i++ {
		ch <- i
	}
}

// Filter filters out specific prime multiples.
func Filter(in <-chan int, out chan<- int, prime int) {
	for {
		i := <-in
		if i%prime != 0 { // e.g. 3/2 = 1.5
			out <- i
		}
	}
}

// The prime sieve.
func main() {
	defer profile.Start(profile.TraceProfile, profile.ProfilePath(".")).Stop()
	ch := make(chan int)      // unbuffered channel
	go Generate(ch)           // launch 1 producer emiting integers; one at a time
	for i := 0; i < 10; i++ { // launch 10 workers filtering primes
		prime := <-ch             // pull prime of the channel; 2 being the first
		fmt.Println(prime)        // print it
		ch1 := make(chan int)     // new channel for this prime filter
		go Filter(ch, ch1, prime) // filter with the pulled prime
		ch = ch1                  // reassign ch
	}
}
