package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

type LidarSensor struct {
	bus     embd.I2CBus
	address byte
}

func NewLidarSensor(i2cbus, addr byte) *LidarSensor {

	lSensor := LidarSensor{bus: embd.NewI2CBus(i2cbus), address: byte(addr)}
	if e := lSensor.bus.WriteByteToReg(lSensor.address, 0x00, 0x00); e != nil {
		log.Panic("Write ", e)
	}
	time.Sleep(300 * time.Millisecond)
	return &lSensor

}

func (ls *LidarSensor) Distance() {

	if v, e := ls.bus.ReadByteFromReg(ls.address, 0x01); e != nil {
		log.Panic("ReadByteFromReg ", e)
	} else {
		fmt.Printf("status: %.8b\n", v)
	}

	for {
		if e := ls.bus.WriteByteToReg(ls.address, 0x00, 0x04); e != nil {
			log.Panic("Write ", e)
		}

		time.Sleep(100 * time.Millisecond)

		if v, e := ls.bus.ReadByteFromReg(ls.address, 0x01); e != nil {
			log.Panic("ReadByteFromReg ", e)
		} else {
			fmt.Printf("status: %.8b\n", v)
		}

		// TODO: error check
		v1, _ := ls.bus.ReadByteFromReg(ls.address, 0x10)
		v2, _ := ls.bus.ReadByteFromReg(ls.address, 0x0f)

		fmt.Println((int(v2) << 8) + int(v1))

		time.Sleep(1 * time.Second)
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
	lidarSensor := NewLidarSensor(1, 0x62)
	lidarSensor.Distance()
	defer lidarSensor.bus.Close()
}
