package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dasfoo/bright-pi"
	"github.com/dasfoo/i2c"
	"github.com/dasfoo/lidar-lite-v2"
)

func ledCircle(b *bpi.BrightPI, circles int) {
	leds := []byte{bpi.WhiteTopLeft, bpi.WhiteTopRight, bpi.WhiteBottomRight, bpi.WhiteBottomLeft}
	for i := 0; i < circles; i++ {
		for _, led := range leds {
			_ = b.Power(led)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	bus, err := i2c.NewBus(1)
	if err != nil {
		log.Println(err)
		return
	}

	b := bpi.NewBrightPI(bus, bpi.DefaultAddress)
	s := lidar.NewLidar(bus, lidar.DefaultAddress)
	_ = s.Reset()

	_ = b.Dim(bpi.WhiteAll, bpi.DefaultDim)
	_ = b.Gain(bpi.DefaultGain)
	go ledCircle(b, 16)

	// Simple GetDistance
	fmt.Println(s.GetDistance())

	// Continuous mode
	_ = s.SetContinuousMode(50, 100*time.Millisecond)
	_ = s.Acquire(false)
	for i := 0; i < 50; i++ {
		fmt.Println("Round", i)
		fmt.Println(s.ReadDistance())
		fmt.Println(s.ReadVelocity())
		time.Sleep(100 * time.Millisecond)
	}

	if hw, sw, err := s.GetVersion(); err == nil {
		fmt.Printf("LIDAR-Lite v2 \"Blue Label\" hw%dsw%d\n", hw, sw)
	}

	// Put devices in low power consumption mode
	_ = s.Sleep()
	_ = b.Sleep()
}
