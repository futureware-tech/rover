package main

import (
	"fmt"
	"log"

	"github.com/dasfoo/i2c"
	"github.com/dasfoo/lidar-lite-v2"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	bus, _ := i2c.NewBus(1)
	s := lidar.NewLidar(bus, lidar.DefaultAddress)

	fmt.Println(s.GetDistance())

	if hw, sw, err := s.GetVersion(); err == nil {
		fmt.Printf("LIDAR-Lite v2 \"Blue Label\" hw%dsw%d\n", hw, sw)
	}
}
