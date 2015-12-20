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
// Model LL-905-PIN-02
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
const (
	Ready = 1 << iota

// RefOverflow     = 1 << iota
// SigOverflow     = 1 << iota
// SignalNotValid  = 1 << iota
// SecondatyReturn = 1 << iota
// Health          = 1 << iota
// ErrorDetection  = 1 << iota
// EyeSafe         = 1 << iota
)

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

//CloseLidar closes releases the resources associated with the bus
func (ls *Lidar) CloseLidar() {
	// Reset FPGA. All registers return to default values
	if e := ls.bus.WriteByteToReg(ls.address, 0x00, 0x00); e != nil {
		log.Panic("Write ", e)
	}
	if err := ls.bus.Close(); err != nil {
		log.Println(err)
	}
}

// GetStatus gets Mode/Status of sensor
func (ls *Lidar) GetStatus() (byte, error) {

	val, err := ls.bus.ReadByteFromReg(ls.address, 0x01)
	if err == nil {
		log.Printf("Status: %.8b\n", val)
	} else {
		log.Println("GetStatus", err)
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
			log.Println("Write ", wErr)
			return 0, wErr
		}

		time.Sleep(100 * time.Millisecond)
	} else {
		log.Println(errSt)
		return 0, errSt
	}

	_, errSt := ls.GetStatus()
	if errSt == nil {
		v1, rErr := ls.bus.ReadByteFromReg(ls.address, 0x10)
		if rErr != nil {
			log.Println("Read ", rErr)
			return 0, rErr
		}
		v2, rErr := ls.bus.ReadByteFromReg(ls.address, 0x0f)
		if rErr != nil {
			log.Println("Read", rErr)
			return 0, rErr
		}

		return ((int(v2) << 8) + int(v1)), nil

	}
	log.Println(errSt)
	return 0, errSt
}

// Velocity is measured by observing the change in distance over a fixed time
// of perion
// TODO 0x04 Check Mode Control
func (ls *Lidar) Velocity() (int, error) {
	// Write 0xa0 to 0x04 to switch on velocity mode
	if wErr := ls.bus.WriteByteToReg(ls.address, 0x04, 0xa0); wErr != nil {
		log.Println("Write ", wErr)
		return -1, wErr
	}
	// Write 0x04 to register 0x00 to start getting distance readings
	if wErr := ls.bus.WriteByteToReg(ls.address, 0x00, 0x04); wErr != nil {
		log.Println("Write ", wErr)
		return -1, wErr
	}
	//Read 1 byte from register 0x09 to get velocity measurement
	for {
		// Sensor is ready for reading
		stVal, err := ls.GetStatus()
		if err != nil {
			log.Println(err)
			return -1, err
		}
		if (stVal & Ready) == 0 {
			val, e := ls.bus.ReadByteFromReg(ls.address, 0x09)
			if e != nil {
				log.Println(e)
				return -1, e
			}
			return int(val), nil
		}
		time.Sleep(300 * time.Microsecond)

	}
}

// BeginContinuous allows to tell the sensor to take a certain number (or
// infinite) readings allowing you to read from it at a continuous rate.
// modePinLow tells the mode pin to go low when a new reading is available.
// TODO search more about interval
func (ls *Lidar) BeginContinuous(modePinLow bool, interval, numberOfReadings byte) error {

	// Register 0x45 sets the time between measurements. Min val os 0x02
	// for proper operations.
	if wErr := ls.bus.WriteByteToReg(ls.address, 0x45, interval); wErr != nil {
		log.Println(wErr)
		return wErr
	}

	if modePinLow {
		if wErr := ls.bus.WriteByteToReg(ls.address, 0x04, 0x21); wErr != nil {
			log.Print(wErr)
			return wErr
		}

	} else {
		// Set register 0x04 to 0x20 to look at "NON-default" value of velocity
		if wErr := ls.bus.WriteByteToReg(ls.address, 0x04, 0x20); wErr != nil {
			log.Print(wErr)
			return wErr
		}
	}
	// Set the number of readings, 0xfe = 254 readings, 0x01 = 1 reading and
	// 0xff = continuous readings
	if wErr := ls.bus.WriteByteToReg(ls.address, 0x11, numberOfReadings); wErr != nil {
		log.Println(wErr)
		return wErr
	}

	// Initiate reading distance
	if wErr := ls.bus.WriteByteToReg(ls.address, 0x00, 0x04); wErr != nil {
		log.Println(wErr)
		return wErr
	}
	return nil
}

//DistanceContinuous reads in continuous mode
func (ls *Lidar) DistanceContinuous() (int, error) {

	val, rErr := ls.bus.ReadWordFromReg(ls.address, 0x8f)
	if rErr != nil {
		log.Println("Read", rErr)
		return -1, rErr
	}

	return int(val), nil
}

func main() {
	log.SetFlags(log.Lshortfile)
	lidar := NewLidar(1, 0x62) // 0x62 the default LidarLite address
	defer lidar.CloseLidar()
	/*for {
		if val, err := lidar.Distance(true); err == nil {
			fmt.Println(val)
			time.Sleep(1 * time.Second)
		}
		if val, err := lidar.Velocity(); err == nil {
			fmt.Println(val)
		}
	}*/

	if err := lidar.BeginContinuous(true, 0x08, 0xff); err == nil {
		for {
			val, e := lidar.DistanceContinuous()

			if e == nil {
				fmt.Println(val)
			}
		}
	}
}
