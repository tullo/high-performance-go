package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"unicode"

	"github.com/pkg/profile"
)

func readbyte(r io.Reader, buf *[1]byte) (rune, error) {
	//var buf [1]byte
	_, err := r.Read(buf[:])
	return rune(buf[0]), err
}

func main() {
	// defer profile.Start().Stop() // Add CPU profiling
	// defer profile.Start(profile.MemProfile).Stop() // Add Memory profiling
	defer profile.Start(profile.MemProfile, profile.MemProfileRate(1)).Stop() // account all allocs

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("could not open file %q: %v", os.Args[1], err)
	}

	var buf [1]byte
	b := bufio.NewReader(f) // Default buffer size = 4096
	words := 0
	inword := false
	for {
		r, err := readbyte(b, &buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("could not read file %q: %v", os.Args[1], err)
		}
		if unicode.IsSpace(r) && inword {
			words++
			inword = false
		}
		inword = unicode.IsLetter(r)
	}
	fmt.Printf("%q: %d words\n", os.Args[1], words)
}
