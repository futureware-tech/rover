package rande

/*
#include <stdlib.h>
*/
import "C"

// Random from stdlib
func Random() int {
	return int(C.random())
}

// Seed (srandom) from stdlib
func Seed(i int) {
	C.srandom(C.uint(i))
}
