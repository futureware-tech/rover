package bb

// #include "bb.h"
import "C"
import (
	"fmt"
	"time"

	"github.com/dasfoo/i2c"
	"github.com/dasfoo/rpi-gpio"
)

// Address of BotBoarduino on i2c bus
const Address = C.I2CAddress

// ResetPin is a Raspberry Pi pin that is connected to the Arduino Reset pin
const ResetPin = 4

// Min/Max allowed value for tilt (degrees)
const (
	MinTilt = C.MinTilt
	MaxTilt = C.MaxTilt
)

// StatusError is returned when the status returned by BB is not compatible with the command
type StatusError struct {
	Status uint16
}

func (se *StatusError) Error() string {
	return fmt.Sprintf("The module requested is not ready (status bits %.16b)", se.Status)
}

// BB is a control interface for BotBoarduino part of the project
type BB struct {
	bus     i2c.Bus
	address byte
}

// NewBB creates a new instance of BotBoarduino to use
func NewBB(bus i2c.Bus, addr byte) *BB {
	return &BB{
		bus:     bus,
		address: addr,
	}
}

// Reset BotBoarduino given that "pin" is wired to board's Reset
func (bb *BB) Reset(pin byte) {
	resetPin := gpio.Pin(pin)
	resetPin.SetMode(gpio.OUTPUT)
	resetPin.Trigger(2*time.Microsecond, false)
	resetPin.SetMode(gpio.INPUT)
}

func register(module int) byte {
	return byte(module << 4)
}

///////////////////////////////////////////////////////////////////////////////////////////////////

// Sending commands/actions to Arduino
const (
	ModuleCommand             = C.ModuleCommand
	commandMeasureEnvironment = C.CommandMeasureEnvironment
	commandSleep              = C.CommandSleep
	commandWake               = C.CommandWake
)

// Sleep reduces power usage of the module (and some hardware)
func (bb *BB) Sleep() error {
	return bb.bus.WriteByteToReg(bb.address, register(ModuleCommand), commandSleep)
}

// Wake is necessary to re-enable hardware disabled by Sleep()
func (bb *BB) Wake() error {
	return bb.bus.WriteByteToReg(bb.address, register(ModuleCommand), commandWake)
}

///////////////////////////////////////////////////////////////////////////////////////////////////

// Reading Arduino board state
const (
	ModuleBoard        = C.ModuleBoard
	moduleBoardStatus  = C.ModuleBoardStatus
	moduleBoardBattery = C.ModuleBoardBattery
)

// GetStatus returns the readiness status bitmask, which can be checked with Module* bit numbers
// Example:
//   if ((GetStatus() & (1 << ModuleEnvironmentSensor)) != 0) {
//     // environment sensor is ready
//   }
func (bb *BB) GetStatus() (uint16, error) {
	return bb.bus.ReadWordFromReg(bb.address, register(ModuleBoard)+moduleBoardStatus)
}

// GetBatteryPercentage returns estimated battery charge, in percent
func (bb *BB) GetBatteryPercentage() (byte, error) {
	// TODO: check status
	v, e := bb.bus.ReadWordFromReg(bb.address, register(ModuleBoard)+moduleBoardBattery)
	// TODO: calibration
	return byte(v >> 2), e
}

///////////////////////////////////////////////////////////////////////////////////////////////////

// Ambient light sensor installed on the robot
const (
	ModuleLightSensor = C.ModuleLightSensor
)

// GetAmbientLight returns ambient light brightness in range 0..1023
func (bb *BB) GetAmbientLight() (uint16, error) {
	// TODO: check status
	return bb.bus.ReadWordFromReg(bb.address, register(ModuleLightSensor))
}

///////////////////////////////////////////////////////////////////////////////////////////////////

// Environment sensor installed on the robot
const (
	ModuleEnvironmentSensor            = C.ModuleEnvironmentSensor
	moduleEnvironmentSensorTemperature = C.ModuleEnvironmentSensorTemperature
	moduleEnvironmentSensorHumidity    = C.ModuleEnvironmentSensorHumidity
)

// GetTemperatureAndHumidity gets ambient temperature in Celsius and relative humidity in %
func (bb *BB) GetTemperatureAndHumidity() (t byte, h byte, e error) {
	// TODO: check status
	if e = bb.bus.WriteByteToReg(bb.address,
		register(ModuleCommand), commandMeasureEnvironment); e != nil {
		return
	}
	// TODO: check status instead of Sleep()
	time.Sleep(500 * time.Millisecond)
	if t, e = bb.bus.ReadByteFromReg(bb.address,
		register(ModuleEnvironmentSensor)+moduleEnvironmentSensorTemperature); e != nil {
		return
	}
	h, e = bb.bus.ReadByteFromReg(bb.address,
		register(ModuleEnvironmentSensor)+moduleEnvironmentSensorHumidity)
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////

// Pan & Tilt installed on the robot, normally has LIDAR attached
const (
	ModuleTilt = C.ModuleTilt
)

// Tilt the LIDAR (or anything else attached to Tilt) for angle degrees (MinTilt-MaxTilt)
func (bb *BB) Tilt(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(ModuleTilt), angle)
}

///////////////////////////////////////////////////////////////////////////////////////////////////

// Robotic ARM controls
const (
	ModuleArm            = C.ModuleArm
	moduleArmBasePan     = C.ModuleArmBasePan
	moduleArmBaseTilt    = C.ModuleArmBaseTilt
	moduleArmElbow       = C.ModuleArmElbow
	moduleArmWristRotate = C.ModuleArmWristRotate
	moduleArmWristTilt   = C.ModuleArmWristTilt
	moduleArmGrip        = C.ModuleArmGrip
)

// ArmBasePan commands BB to rotate robotic arm base, 0-180 degrees
func (bb *BB) ArmBasePan(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(ModuleArm)+moduleArmBasePan, angle)
}

// ArmBaseTilt commands BB to tilt robotic arm, 0-180 degrees
func (bb *BB) ArmBaseTilt(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(ModuleArm)+moduleArmBaseTilt, angle)
}

// ArmElbow commands BB to bend robotic arm's elbow, 0-180 degrees
func (bb *BB) ArmElbow(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(ModuleArm)+moduleArmElbow, angle)
}

// ArmWristRotate commands BB to rotate the wrist, 0-180 degrees
func (bb *BB) ArmWristRotate(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(ModuleArm)+moduleArmWristRotate, angle)
}

// ArmWristTilt commands BB to tilt the wrist, 0-180 degrees
func (bb *BB) ArmWristTilt(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(ModuleArm)+moduleArmWristTilt, angle)
}

// ArmGrip commands BB to change the grip position and width, 0-180
func (bb *BB) ArmGrip(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(ModuleArm)+moduleArmGrip, angle)
}
