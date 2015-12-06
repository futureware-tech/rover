package drivers

import (
	"os"
	"strconv"
	"syscall"
)

// Unexport GPIO pin to make it usable via EMBD.
// Call once at start for each GPIO pin you're using directly (not via i2c etc).
func ResetGPIOPin(pin byte) error {
	if unexport, e := os.OpenFile("/sys/class/gpio/unexport", os.O_WRONLY, 0); e != nil {
		return e
	} else {
		defer unexport.Close()
		if _, e := unexport.WriteString(strconv.Itoa(int(pin))); e != nil {
			if ose, ok := e.(*os.PathError); !ok || ose.Err != syscall.EINVAL {
				return e
			}
		}
		return nil
	}
}
