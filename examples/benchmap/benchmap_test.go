package main

import "testing"

// use go test -bench=. -benchmem

var capitals = map[string]string{
	"Algeria":                "Algiers",
	"Argentina":              "Buenos Aires",
	"Australia":              "Canberra",
	"Austria":                "Vienna",
	"Bahamas":                "Nassau",
	"Belarus":                "Minsk",
	"Bosnia and Herzegovina": "Sarajevo",
	"Brazil":                 "Brasilia",
	"Bulgaria":               "Sofia",
	"Canada":                 "Ottawa",
	"China":                  "Beijing",
	"Croatia":                "Zagreb",
	"Cuba":                   "Havana",
	"Egypt":                  "Cairo",
	"France":                 "Paris",
	"Germany":                "Berlin",
	"Indonesia":              "Jakarta",
	"Ireland":                "Dublin",
	"Jamaica":                "Kingston",
	"Japan":                  "Tokyo",
	"Luxembourg":             "Luxembourg",
}

var sink string

func BenchmarkMapLookup(b *testing.B) {
	var key = []byte{'F', 'r', 'a', 'n', 'c', 'e'}
	var r string
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		r = capitals[string(key)] // The compiler implements a specific optimisation for this case
	}
	sink = r
}

func BenchmarkMapLookup2(b *testing.B) {
	var key = []byte{'F', 'r', 'a', 'n', 'c', 'e'}
	var r string
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		k := string(key) // No compiler optimization
		r = capitals[k]
	}
	sink = r
}
