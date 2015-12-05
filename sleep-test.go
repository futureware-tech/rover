package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	for i := 0; i < 1000; i++ {
		time.Sleep(1 * time.Nanosecond)
	}
	fmt.Printf("resolution: %f nanosec\n",
		float64(time.Since(start).Nanoseconds())/1000.0)
}
