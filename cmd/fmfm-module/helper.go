package main

/*
static void _set_longlong_array(long long* p, long long offset, long long value) {
	p[offset] = value;
}
static void _set_uchar_array(unsigned char* p, long long offset, unsigned char value) {
	p[offset] = value;
}
*/
import "C"

import (
	"sort"
)

func collectInts(fn func(chan<- int)) []int {
	found := map[int]struct{}{}
	ch := make(chan int, 100)
	go func() {
		fn(ch)
		close(ch)
	}()
	for v := range ch {
		found[v] = struct{}{}
	}

	result := []int{}
	for v := range found {
		result = append(result, v)
	}
	sort.Ints(result)
	return result
}

func writeInts(out *C.longlong, a []int) C.longlong {
	for i, v := range a {
		C._set_longlong_array(out, C.longlong(i), C.longlong(v))
	}
	return C.longlong(len(a))
}

func writeBytes(out *C.uchar, a []byte) C.longlong {
	for i, v := range a {
		C._set_uchar_array(out, C.longlong(i), C.uchar(v))
	}
	return C.longlong(len(a))
}
