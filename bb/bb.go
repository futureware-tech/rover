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

const (
	moduleCommand           = C.MODULE_COMMAND
	moduleBoard             = C.MODULE_BOARD
	moduleMotor             = C.MODULE_MOTOR
	moduleLightSensor       = C.MODULE_LIGHT_SENSOR
	modulePan               = C.MODULE_PAN
	moduleTilt              = C.MODULE_TILT
	moduleEnvironmentSensor = C.MODULE_ENVIRONMENT_SENSOR

	motorLeft                    = C.MOTOR_LEFT
	motorRight                   = C.MOTOR_RIGHT
	boardStatus                  = C.BOARD_STATUS
	boardBattery                 = C.BOARD_BATTERY
	environmentSensorTemperature = C.ENVIRONMENT_SENSOR_TEMPERATURE
	environmentSensorHumidity    = C.ENVIRONMENT_SENSOR_HUMIDITY

	commandMeasureEnvironment = C.COMMAND_MEASURE_ENVIRONMENT
	commandSleep              = C.COMMAND_SLEEP
	commandWake               = C.COMMAND_WAKE
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

// GetStatus returns the readiness status bitmask, which can be checked with module* bit numbers
func (bb *BB) GetStatus() (uint16, error) {
	return bb.bus.ReadWordFromReg(bb.address, register(moduleBoard)+boardStatus)
}

// Pan the LIDAR (or anything else attached to Pan/Tilt) for angle degrees (0-180)
func (bb *BB) Pan(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(modulePan), angle)
}

// Tilt the LIDAR (or anything else attached to Pan/Tilt) for angle degrees (0-MaxTilt)
func (bb *BB) Tilt(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(moduleTilt), angle)
}

// MotorLeft changes left motor speed, range -MaxMotorSpeed .. MaxMotorSpeed
func (bb *BB) MotorLeft(speed int8) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(moduleMotor)+motorLeft,
		byte(int(speed)+MaxMotorSpeed))
}

// MotorRight changes right motor speed, range -MaxMotorSpeed .. MaxMotorSpeed
func (bb *BB) MotorRight(speed int8) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(moduleMotor)+motorRight,
		byte(int(speed)+MaxMotorSpeed))
}

// GetBatteryPercentage returns estimated battery charge, in percent
func (bb *BB) GetBatteryPercentage() (byte, error) {
	// TODO: check status
	return bb.bus.ReadByteFromReg(bb.address, register(moduleBoard)+boardBattery)
}

// GetBrightness returns ambient brightness in range 0..1023
func (bb *BB) GetBrightness() (uint16, error) {
	// TODO: check status
	return bb.bus.ReadWordFromReg(bb.address, register(moduleLightSensor))
}

// GetTemperatureAndHumidity gets ambient temperature in Celcius and relative humidity in %
func (bb *BB) GetTemperatureAndHumidity() (t byte, h byte, e error) {
	if e = bb.bus.WriteByteToReg(bb.address,
		register(moduleCommand), commandMeasureEnvironment); e != nil {
		return
	}
	// TODO: check status instead of Sleep()
	time.Sleep(500 * time.Millisecond)
	if t, e = bb.bus.ReadByteFromReg(bb.address,
		register(moduleEnvironmentSensor)+environmentSensorTemperature); e != nil {
		return
	}
	h, e = bb.bus.ReadByteFromReg(bb.address,
		register(moduleEnvironmentSensor)+environmentSensorHumidity)
	return
}

// Sleep reduces power usage of the module (and some hardware)
func (bb *BB) Sleep() error {
	return bb.bus.WriteByteToReg(bb.address, register(moduleCommand), commandSleep)
}

// Wake is necessary to re-enable hardware disabled by Sleep()
func (bb *BB) Wake() error {
	return bb.bus.WriteByteToReg(bb.address, register(moduleCommand), commandWake)
}
