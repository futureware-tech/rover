package bb

// #include "bb.h"
import "C"
import (
	"time"

	"github.com/dasfoo/i2c"
	"github.com/dasfoo/rpi-gpio"
)

// Address of BotBoarduino on i2c bus
const Address = C.I2C_ADDRESS

// ResetPin is a default Reset Pin to use
const ResetPin = 4

// MaxTilt is a maximum allowed value for tilt (degrees)
const MaxTilt = C.MAX_TILT

// MaxMotorSpeed is a max speed (positive or negative) for motors
const MaxMotorSpeed = 90

// Module IDs
const (
	ModuleCommand           = C.MODULE_COMMAND
	ModuleBoard             = C.MODULE_BOARD
	ModuleMotor             = C.MODULE_MOTOR
	ModuleLightSensor       = C.MODULE_LIGHT_SENSOR
	ModulePan               = C.MODULE_PAN
	ModuleTilt              = C.MODULE_TILT
	ModuleEnvironmentSensor = C.MODULE_ENVIRONMENT_SENSOR
)

// BB is a control interface for BotBoarduino part of the project
type BB struct {
	bus     *i2c.Bus
	address byte
}

// NewBB creates a new instance of BotBoarduino to use
func NewBB(bus *i2c.Bus, addr byte) *BB {
	return &BB{
		bus:     bus,
		address: addr,
	}
}

func register(module int) byte {
	return byte(module << 4)
}

// Reset BotBoarduino given that "pin" is wired to board's Reset
func (bb *BB) Reset(pin byte) {
	resetPin := gpio.Pin(pin)
	resetPin.SetMode(gpio.OUTPUT)
	resetPin.Trigger(2*time.Microsecond, false)
	resetPin.SetMode(gpio.INPUT)
}

// GetStatus returns the readiness status bitmask, which can be checked with Module* bit numbers
func (bb *BB) GetStatus() (uint16, error) {
	return bb.bus.ReadWordFromReg(bb.address, register(ModuleBoard))
}

// Pan the LIDAR (or anything else attached to Pan/Tilt) for angle degrees (0-180)
func (bb *BB) Pan(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(ModulePan), angle)
}

// Tilt the LIDAR (or anything else attached to Pan/Tilt) for angle degrees (0-MaxTilt)
func (bb *BB) Tilt(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(ModuleTilt), angle)
}

// MotorLeft changes left motor speed, range -MaxMotorSpeed .. MaxMotorSpeed
func (bb *BB) MotorLeft(speed int8) error {
	return bb.bus.WriteByteToReg(bb.address, register(ModuleMotor),
		byte(int(speed)+MaxMotorSpeed))
	// TODO: check status
}

// MotorRight changes right motor speed, range -MaxMotorSpeed .. MaxMotorSpeed
func (bb *BB) MotorRight(speed int8) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(ModuleMotor)+1,
		byte(int(speed)+MaxMotorSpeed))
}

// GetBatteryPercentage returns estimated battery charge, in percent
func (bb *BB) GetBatteryPercentage() (byte, error) {
	// TODO: check status
	return bb.bus.ReadByteFromReg(bb.address, register(ModuleBoard)+1)
}
