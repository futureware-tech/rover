package main

import (
	"fmt"
	"time"

	"github.com/dasfoo/rover/rpi/vchi"
)

func main() {
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
	}
}
