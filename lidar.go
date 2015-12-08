package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

type Lidar struct {
	bus     embd.I2CBus
	address byte
}

//Ready - Ready status. 0 - ready for a new command, 1 - busy
//          with acquisition
//RefOverflow - Overflow detected in correlation process assotieted with
//          a reference acquisition
//SignalOverflow - Overflow detected in correlation process assotieted with a signal
//          acquisition
//SignalNotValid - Indicates that the signal correlation peak is equal to or below
//          correlation record threshold
//SecondaryReturn - Secondary return detected above correlation noise floor threshold
//Health - 1 if is good, 0 if is bad
//ErrorDetection - Process error detected/measurement invalid
//EyeSafe - This bit will go high if eye-safety protection has been activated
const (
	Ready           = 1 << iota
	RefOverflow     = 1 << iota
	SigOverflow     = 1 << iota
	SignalNotValid  = 1 << iota
	SecondatyReturn = 1 << iota
	Health          = 1 << iota
	ErrorDetection  = 1 << iota
	EyeSafe         = 1 << iota
)

func NewLidar(i2cbus, addr byte) *Lidar {

	lSensor := Lidar{bus: embd.NewI2CBus(i2cbus), address: byte(addr)}
	if e := lSensor.bus.WriteByteToReg(lSensor.address, 0x00, 0x00); e != nil {
		log.Panic("Write ", e)
	}
	time.Sleep(300 * time.Millisecond)
	return &lSensor

}

func (ls *Lidar) GetStatus() (byte, error) {

	if val, err := ls.bus.ReadByteFromReg(ls.address, 0x01); err == nil {
		fmt.Printf("status: %.8b\n", val)
		return val, nil
	} else {
		log.Panic("GetStatus ", err)
		return 0, err
	}

}

//stablizePreampFlag - true - take aquisition with DC stabilisation/correction.
//false - it will read faster, but you will need to stabilize DC every once in
//awhile(ex. 1 out of every 100 readings is typically good)
func (ls *Lidar) Distance(stablizePreampFlag bool) {

	ls.GetStatus()

	for {
		if stablizePreampFlag {
			if e := ls.bus.WriteByteToReg(ls.address, 0x00, 0x04); e != nil {
				log.Panic("Write ", e)
			}
		} else if e := ls.bus.WriteByteToReg(ls.address, 0x00, 0x03); e != nil {
			log.Panic("Write ", e)
		}

		time.Sleep(100 * time.Millisecond)

		ls.GetStatus()
		// TODO: error check
		v1, _ := ls.bus.ReadByteFromReg(ls.address, 0x10)
		v2, _ := ls.bus.ReadByteFromReg(ls.address, 0x0f)

		fmt.Println((int(v2) << 8) + int(v1))

		time.Sleep(1 * time.Second)
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
	lidar := NewLidar(1, 0x62)
	lidar.Distance(true)
	defer lidar.bus.Close()
}
