package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

func main() {
	bus := embd.NewI2CBus(1)
	defer bus.Close()

	lidar := byte(0x62)

	if v, e := bus.ReadByteFromReg(lidar, 0x01); e != nil {
		log.Panic("ReadByteFromReg ", e)
	} else {
		fmt.Printf("status: %.8b\n", v)
	}

	for {
		if e := bus.WriteByteToReg(lidar, 0x00, 0x04); e != nil {
			log.Panic("Write ", e)
		}

		time.Sleep(100 * time.Millisecond)

		if v, e := bus.ReadByteFromReg(lidar, 0x01); e != nil {
			log.Panic("ReadByteFromReg ", e)
		} else {
			fmt.Printf("status: %.8b\n", v)
		}

		// TODO: error check
		v1, _ := bus.ReadByteFromReg(lidar, 0x10)
		v2, _ := bus.ReadByteFromReg(lidar, 0x0f)

		fmt.Println((v2 << 8) + v1)

		time.Sleep(1 * time.Second)
	}
}
