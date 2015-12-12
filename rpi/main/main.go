package main

import (
	"fmt"
	"time"

	"github.com/dasfoo/rover/rpi/gpio"
	"github.com/dasfoo/rover/rpi/vchi"
)

func main() {
	dht11 := gpio.Pin(4)
	for {
		var h, t byte
		for h+t == 0 {
			h, t = dht11.DHT11()
			time.Sleep(500 * time.Microsecond)
		}
		fmt.Printf("humidity = %d, temperature = %d\n", h, t)

		out, err := vchi.Send("measure_temp")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(out)

		out, err = vchi.Send("measure_volts")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(out)
		time.Sleep(time.Second)
	}
}
