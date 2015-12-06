package main

import (
	"fmt"
	"log"

	"github.com/dasfoo/rover/drivers"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

const (
	DHT11_SENSOR = 4
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if e := drivers.ResetGPIOPin(DHT11_SENSOR); e != nil {
		log.Panic(e)
	}

	embd.InitGPIO()
	defer embd.CloseGPIO()

	if pin, e := embd.NewDigitalPin(DHT11_SENSOR); e != nil {
		log.Panic(e)
	} else {
		defer pin.Close()

		s := drivers.NewDHT11(pin)
		if retries, e := s.Read(10); e != nil {
			log.Panic(e)
		} else {
			fmt.Printf("%d*C, %d%% after %d retries\n", s.Temperature,
				s.Humidity, retries)
		}
	}
}
