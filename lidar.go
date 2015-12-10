package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

// Lidar is structure to access basic functions of LidarLite
// LidarLite_V2 blue label
// Documentation on http://lidarlite.com/docs/v2/specs_and_hardware
type Lidar struct {
	bus     embd.I2CBus
	address byte
}

// Ready - Ready status. 0 - ready for a new command, 1 - busy
//          with acquisition
// RefOverflow - Overflow detected in correlation process assotieted with
//          a reference acquisition
// SignalOverflow - Overflow detected in correlation process assotieted with a signal
//          acquisition
// SignalNotValid - Indicates that the signal correlation peak is equal to or below
//          correlation record threshold
// SecondaryReturn - Secondary return detected above correlation noise floor threshold
// Health - 1 if is good, 0 if is bad
// ErrorDetection - Process error detected/measurement invalid
// EyeSafe - This bit will go high if eye-safety protection has been activated
//const (
// Ready           = 1 << iota
// RefOverflow     = 1 << iota
// SigOverflow     = 1 << iota
// SignalNotValid  = 1 << iota
// SecondatyReturn = 1 << iota
// Health          = 1 << iota
// ErrorDetection  = 1 << iota
// EyeSafe         = 1 << iota
//)

// NewLidar sets the configuration for the sensor
// Write 0x00 to Register 0x00 reset FPGA. Re-loads FPGA from internal Flash
// memory: all registers return to default values
func NewLidar(i2cbus, addr byte) *Lidar {

	lSensor := Lidar{bus: embd.NewI2CBus(i2cbus), address: byte(addr)}
	if e := lSensor.bus.WriteByteToReg(lSensor.address, 0x00, 0x00); e != nil {
		log.Panic("Write ", e)
	}
	time.Sleep(300 * time.Millisecond)
	return &lSensor

}

// GetStatus gets Mode/Status of sensor
func (ls *Lidar) GetStatus() (byte, error) {

	val, err := ls.bus.ReadByteFromReg(ls.address, 0x01)
	if err == nil {
		log.Println("Status ", val)
	} else {
		log.Panic("GetStatus ", err)
		return 0, err
	}
	return val, nil

}

// Distance reads the distance from LidarLite
// stablizePreampFlag - true - take aquisition with DC stabilisation/correction.
// false - it will read faster, but you will need to stabilize DC every once in
// awhile(ex. 1 out of every 100 readings is typically good)
func (ls *Lidar) Distance(stablizePreampFlag bool) (int, error) {

	var wErr error // Write and Read errors

	if _, errSt := ls.GetStatus(); errSt == nil {
		if stablizePreampFlag {
			wErr = ls.bus.WriteByteToReg(ls.address, 0x00, 0x04)
		} else {
			wErr = ls.bus.WriteByteToReg(ls.address, 0x00, 0x03)
		}

		if wErr != nil {
			log.Panic("Write ", wErr)
			return 0, wErr
		}

		time.Sleep(100 * time.Millisecond)
	} else {
		log.Panic(errSt)
		return 0, errSt
	}

	_, errSt := ls.GetStatus()
	if errSt == nil {
		v1, rErr := ls.bus.ReadByteFromReg(ls.address, 0x10)
		if rErr != nil {
			log.Panic("Read ", rErr)
			return 0, rErr
		}
		v2, rErr := ls.bus.ReadByteFromReg(ls.address, 0x0f)
		if rErr != nil {
			log.Panic("Read", rErr)
			return 0, rErr
		}

		return ((int(v2) << 8) + int(v1)), nil

	}
	log.Panic(errSt)
	return 0, errSt

}

func main() {
	log.SetFlags(log.Lshortfile)
	lidar := NewLidar(1, 0x62) // 0x62 the default LidarLite address
	defer lidar.bus.Close()
	for {
		if val, err := lidar.Distance(true); err == nil {
			fmt.Println(val)
		}
	}

}
