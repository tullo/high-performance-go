package p

import "sync"

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
