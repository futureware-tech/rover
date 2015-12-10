package rande_test

import (
	"testing"
	"time"

	"github.com/dasfoo/rover/rande"
)

func TestRandomSameForSeed(t *testing.T) {
	rande.Seed(0)
	a := rande.Random()
	rande.Seed(0)
	if rande.Random() != a {
		t.Fail()
	}
}

func TestRandomDifferentForSeed(t *testing.T) {
	rande.Seed(int(time.Now().Unix()))
	a := rande.Random()
	time.Sleep(2 * time.Second)
	rande.Seed(int(time.Now().Unix()))
	if rande.Random() == a {
		t.Fail()
	}
}
