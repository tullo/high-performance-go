package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"unicode"
)

func readbyte(r io.Reader, buf *[1]byte) (rune, error) {
	//var buf [1]byte
	_, err := r.Read(buf[:])
	return rune(buf[0]), err
}

type bytereader struct {
	buf [1]byte
	r   io.Reader
}

func (b *bytereader) next() (rune, error) {
	_, err := b.r.Read(b.buf[:])
	return rune(b.buf[0]), err
}

func main() {
	// defer profile.Start().Stop() // Add CPU profiling
	// defer profile.Start(profile.MemProfile).Stop() // Add Memory profiling
	// defer profile.Start(profile.MemProfile, profile.MemProfileRate(1)).Stop() // account all allocs

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("could not open file %q: %v", os.Args[1], err)
	}

	br := bytereader{
		r: bufio.NewReader(f),
	}
	words := 0
	inword := false
	for {
		r, err := br.next()
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
