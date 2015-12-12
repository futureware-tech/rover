package rpi

/*
#cgo CFLAGS: -I ${SRCDIR}

#cgo amd64 CFLAGS: -I ${SRCDIR}/amd64

#cgo arm CFLAGS: -I ${SRCDIR}/arm/opt/vc/include
#cgo arm CFLAGS: -I ${SRCDIR}/arm/opt/vc/include/interface/vcos/pthreads
#cgo arm CFLAGS: -I ${SRCDIR}/arm/opt/vc/include/interface/vmcs_host/linux
#cgo arm LDFLAGS: -L ${SRCDIR}/arm/opt/vc/lib
#cgo arm LDFLAGS: -lbcm_host -lvcos -lvchiq_arm

#include <vcgencmd.c>
#include <stdlib.h>
*/
import "C"
import (
	"log"
	"unsafe"
)

func init() {
	if _, err := C.Start(); err != nil {
		log.Println("failed to initialize vchi:", err)
	}
	// TODO: call C.Stop() at the end.
}

// SendVCHICommand sends a command to VCHI
func SendVCHICommand(command string) (string, error) {
	cCommand := C.CString(command)
	defer C.free(unsafe.Pointer(cCommand))

	output, err := C.Send(cCommand)
	if err != nil {
		return "", err
	}
	return C.GoString(output), nil
}
