package mc

// #include "mc.h"
import "C"
import (
	"bytes"
	"encoding/binary"

	"github.com/dasfoo/i2c"
)

// Address of BotBoarduino on i2c bus
const Address = C.I2CAddress

// MaxSpeed is a max speed (positive or negative) for motors
const MaxSpeed = 90

// MC is a control interface for Motor Controller part of the project
type MC struct {
	bus     i2c.Bus
	address byte
}

// NewMC creates a new instance of BotBoarduino to use
func NewMC(bus i2c.Bus, addr byte) *MC {
	return &MC{
		bus:     bus,
		address: addr,
	}
}

const (
	registerCommand = C.RegisterCommand

	registerMotorLeft  = C.RegisterMotorLeft
	registerMotorRight = C.RegisterMotorRight
)

// Encoder registers for ReadEncoder function
const (
	EncoderLeftFront  = C.RegisterEncoderLeftFront
	EncoderLeftBack   = C.RegisterEncoderLeftBack
	EncoderRightFront = C.RegisterEncoderRightFront
	EncoderRightBack  = C.RegisterEncoderRightBack
)

const (
	commandBrake        = C.CommandBrake
	commandReleaseBrake = C.CommandReleaseBrake
	commandSleep        = C.CommandSleep
	commandWake         = C.CommandWake
)

// Sleep reduces power usage of the module (and some hardware)
func (mc *MC) Sleep() error {
	return mc.bus.WriteByteToReg(mc.address, registerCommand, commandSleep)
}

// Wake is necessary to re-enable hardware disabled by Sleep()
func (mc *MC) Wake() error {
	return mc.bus.WriteByteToReg(mc.address, registerCommand, commandWake)
}

// Brake enables or disables an algorithm which tries to keep motor encoder deltas to zero.
// It basically means that whenever an encoder detects wheel movement, power is applied to
// the motors to revert to the previous encoder position.
// This is a somewhat similar to a car brake.
func (mc *MC) Brake(brake bool) error {
	command := byte(commandBrake)
	if !brake {
		command = commandReleaseBrake
	}
	return mc.bus.WriteByteToReg(mc.address, registerCommand, command)
}

// Right motor start, speed in rage -MaxSpeed .. MaxSpeed
func (mc *MC) Right(speed int8) error {
	return mc.bus.WriteByteToReg(mc.address, registerMotorRight, byte(int(speed)+MaxSpeed))
}

// Left motor start, speed in rage -MaxSpeed .. MaxSpeed
func (mc *MC) Left(speed int8) error {
	return mc.bus.WriteByteToReg(mc.address, registerMotorLeft, byte(int(speed)+MaxSpeed))
}

// ReadEncoder reads encoder value, in absolute steps, Encoder{Left,Right}{Front,Back}
func (mc *MC) ReadEncoder(encoder byte) (value int32, e error) {
	data := make([]byte, 4)
	if _, e = mc.bus.ReadSliceFromReg(mc.address, encoder, data); e != nil {
		return
	}
	if e = binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &value); e != nil {
		return
	}
	return
}
