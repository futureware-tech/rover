package sim800

import (
	"time"

	"github.com/dasfoo/rpi-gpio"
)

// Functions for Itead SIM800 stackable.
// http://wiki.iteadstudio.com/RPI_SIM800_GSM/GPRS_ADD-ON_V2.0

// Reset Itead SIM800 stackable. Only works when power is on.
func Reset() {
	pin := gpio.Pin(18)
	pin.SetMode(gpio.OUTPUT)
	pin.Write(true)
	time.Sleep(100 * time.Millisecond)
	pin.SetMode(gpio.INPUT)
}

// TogglePower changes power state of the Itead SIM800 stackable.
// The result can be verified with AT commands, e.g.:
//   screen /dev/ttyAMA0 115200
//   AT+CPIN? # ask if PIN is necessary for the SIM card
//   AT+CSPN? # ask SIM service provider
//   AT+CREG? # ask network registration status
// Alternatively SIM800 can be shut down with AT commands:
//   AT+CPOWD=1 # or =0 for urgent shutdown
// Note: power cycling or resetting does not restore all factory settings.
// For factory settings, use the "ATZ" command.
func TogglePower() {
	pin := gpio.Pin(17)
	pin.SetMode(gpio.OUTPUT)
	pin.Write(true)
	time.Sleep(2 * time.Second)
	pin.SetMode(gpio.INPUT)
}
