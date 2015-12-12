package gpio

/*
#include "gpio.h"
*/
import "C"

func init() {
	C.gpioInitialise()
}

// Pin type to operate single GPIO pin state, mode and value
type Pin byte

// Mode represents pin mode (see options below)
type Mode byte

// Pin operating mode
const (
	// INPUT (available for read) mode
	INPUT  Mode = C.PI_INPUT
	OUTPUT      = C.PI_OUTPUT
	ALT0        = C.PI_ALT0
	ALT1        = C.PI_ALT1
	ALT2        = C.PI_ALT2
	ALT3        = C.PI_ALT3
	ALT4        = C.PI_ALT4
	ALT5        = C.PI_ALT5
)

// PullState is a pin pull-up/down state
type PullState byte

// Pull states
const (
	// Pull off
	OFF  PullState = C.PI_PUD_OFF
	DOWN           = C.PI_PUD_DOWN
	UP             = C.PI_PUD_UP
)

// SetMode sets pin operating mode
func (pin Pin) SetMode(mode Mode) {
	C.gpioSetMode(C.uint(pin), C.uint(mode))
}

// GetMode gets pin operating mode
func (pin Pin) GetMode() Mode {
	return Mode(C.gpioGetMode(C.uint(pin)))
}

func (pin Pin) Read() bool {
	return C.gpioRead(C.uint(pin)) > 0
}

func (pin Pin) Write(value bool) {
	var intValue C.uint
	if value {
		intValue = 1
	}
	C.gpioWrite(C.uint(pin), intValue)
}

/*
void gpioSetPullUpDown(unsigned gpio, unsigned pud)
void gpioTrigger(unsigned gpio, unsigned pulseLen, unsigned level)
*/
