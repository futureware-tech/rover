package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

func main() {
	embd.InitGPIO()
	defer embd.CloseGPIO()

	if pin, e := embd.NewDigitalPin(4); e != nil {
		log.Panic("NewDigitalPin ", e)
	} else {
		defer pin.Close()

		if e := pin.SetDirection(embd.Out); e != nil {
			log.Panic("SetDirection OUT ", e)
		}
		if e := pin.Write(embd.High); e != nil {
			log.Panic("DigitalWrite HIGH ", e)
		}
		time.Sleep(500 * time.Millisecond)
		if e := pin.Write(embd.Low); e != nil {
			log.Panic("DigitalWrite LOW ", e)
		}
		time.Sleep(18 * time.Millisecond)
		if e := pin.SetDirection(embd.In); e != nil {
			log.Panic("SetDirection IN ", e)
		}

		started_at := time.Now()
		for i := 0; i < 60; i++ {
			if r, e := pin.Read(); e != nil {
				log.Panic("Read ", e)
			} else {
				fmt.Printf("after %f microseconds: %d\n",
					time.Since(started_at).Seconds()*1000000, r)
			}
		}
	}

	fmt.Println("Success!")
}
