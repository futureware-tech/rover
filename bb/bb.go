package bb

// #include "bb.h"
import "C"
import "github.com/dasfoo/i2c"

// Address of BotBoarduino on i2c bus
const Address = C.I2C_ADDRESS

// MaxTilt is a maximum allowed value for tilt (degrees)
const MaxTilt = C.MAX_TILT

// Module IDs
const (
	ModuleCommand           = C.MODULE_COMMAND
	ModuleBoard             = C.MODULE_BOARD
	ModuleMotor             = C.MODULE_MOTOR
	ModuleLightSensor       = C.MODULE_LIGHT_SENSOR
	ModuleArm               = C.MODULE_ARM
	ModulePanTilt           = C.MODULE_PAN_TILT
	ModuleEnvironmentSensor = C.MODULE_ENVIRONMENT_SENSOR
	ModuleSpeech            = C.MODULE_SPEECH
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

// GetStatus returns the readiness status bitmask, which can be checked with Module* bit numbers
func (bb *BB) GetStatus() (uint16, error) {
	return bb.bus.ReadWordFromReg(bb.address, register(ModuleBoard))
}

// Pan the LIDAR (or anything else attached to Pan/Tilt) for angle degrees (0-180)
func (bb *BB) Pan(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(ModulePanTilt), angle)
}

// Tilt the LIDAR (or anything else attached to Pan/Tilt) for angle degrees (0-MaxTilt)
func (bb *BB) Tilt(angle byte) error {
	// TODO: check status
	return bb.bus.WriteByteToReg(bb.address, register(ModulePanTilt)+1, angle)
}
