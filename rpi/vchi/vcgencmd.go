package vchi

/*
#cgo CFLAGS: -I ${SRCDIR}/userland
#cgo CFLAGS: -I ${SRCDIR}/userland/interface/vcos/pthreads
#cgo CFLAGS: -I ${SRCDIR}/userland/interface/vmcs_host/linux
// A hack. To add bcm_host to DT_NEEDED tags in the ELF output, we need to put
// -lbcm_host on the ld commandline. ld will then unconditionally search for it.
// However, the library is not present on a desktop system that is used for
// cross-compiling. Thus we use runtime linking and create a fake .so file:
// $ arm-linux-gnueabi-gcc -shared -x c /dev/null -o libbcm_host.so
#cgo LDFLAGS: -Wl,--unresolved-symbols=ignore-in-object-files
#cgo arm LDFLAGS: -Wl,--no-as-needed -L${SRCDIR} -lbcm_host
// Alternatively, if in future we'll be building userland and getting real
// libraries from it, '#cgo amd64 CFLAGS -I <directory_with_fake_h_file>'
// will help to avoid issues with 'go get'.

#include "vcgencmd.h"
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

// Send VCHI command
func Send(command string) (string, error) {
	cCommand := C.CString(command)
	defer C.free(unsafe.Pointer(cCommand))

	output, err := C.Send(cCommand)
	if err != nil {
		return "", err
	}
	return C.GoString(output), nil
}
