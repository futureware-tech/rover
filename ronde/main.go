package main

import (
	"fmt"
	"time"

	"github.com/dasfoo/rover/rande"
)

func main() {
	rande.Seed(int(time.Now().Unix()))
	fmt.Println("random number:", rande.Random())
}
