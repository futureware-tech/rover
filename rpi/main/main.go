package main

import (
	"fmt"
	"time"

	"github.com/dasfoo/rover/rpi/gpio"
	"github.com/dasfoo/rover/rpi/vchi"
)

func main() {
	p := gpio.Pin(4)
	for {
		out, err := vchi.Send("measure_temp")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("temp:", out)

		out, err = vchi.Send("measure_volts")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("volts:", out)
		time.Sleep(time.Second)

		fmt.Println("Pin 4 status:", p.Read())
	}
}
