package cgo

/*
int f() {
	return 42;
}
*/
import "C"
import "fmt"

// F 42
func F() int {
	x := C.f()
	fmt.Printf("%[1]T: %#[1]v\n", x)
	return int(x)
}
