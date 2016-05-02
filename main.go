package main

import (
	"log"
	"time"

	"github.com/dasfoo/i2c"
	"github.com/dasfoo/rover/mc"
)

func main() {

	if bus, err := i2c.NewBus(1); err != nil {
		log.Fatal(err)
	} else {
		// Silence i2c bus log
		//bus.Log = func(string, ...interface{}) {}
		motors := mc.NewMC(bus, mc.Address)
		_ = motors.Left(30)
		_ = motors.Right(30)
		time.Sleep(1 * time.Second)

		_ = motors.Left(0)
		_ = motors.Right(0)

	}

}
