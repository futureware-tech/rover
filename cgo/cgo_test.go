package cgo

import "testing"

func TestF(t *testing.T) {
	if F() != 42 {
		t.Fail()
	}
}
