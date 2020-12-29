package p

import "sync"

// A Pool's purpose is to cache allocated but unused items
// for later reuse, relieving pressure on the garbage collector.
var pool = sync.Pool{
	// New specifies a function to generate a value
	// when Get would otherwise return nil
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

func fn() {
	buf := pool.Get().([]byte) // takes from pool or calls New
	// do work
	pool.Put(buf) // returns buf to the pool
}
