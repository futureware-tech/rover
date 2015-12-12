package main

import (
	"fmt"
	"time"

	"github.com/dasfoo/rover/rpi"
)

func main() {
	for {
		out, err := rpi.SendVCHICommand("measure_temp")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("temp:", out)

		out, err = rpi.SendVCHICommand("measure_volts")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("volts:", out)
		time.Sleep(time.Second)
	}
}
