package rpi

/*
#cgo amd64 CFLAGS: -I ${SRCDIR}/amd64

#cgo arm CFLAGS: -I ${SRCDIR}/arm/opt/vc/include
#cgo arm CFLAGS: -I ${SRCDIR}/arm/opt/vc/include/interface/vcos/pthreads
#cgo arm CFLAGS: -I ${SRCDIR}/arm/opt/vc/include/interface/vmcs_host/linux
#cgo arm LDFLAGS: -L ${SRCDIR}/arm/opt/vc/lib
#cgo arm LDFLAGS: -lbcm_host -lvcos -lvchiq_arm

#include <bcm_host.h>
#include <stdlib.h>
#include <string.h>

char *f(char *s) {
	VCHI_INSTANCE_T vchi;
	VCHI_CONNECTION_T *vchi_connections;

	bcm_host_init();
	vcos_init();
	if (vchi_initialise(&vchi) != 0) {
		return NULL;
	}

	if (vchi_connect(NULL, 0, vchi) != 0) {
		return NULL;
	}

	vc_vchi_gencmd_init(vchi, &vchi_connections, 1);

	char *r = malloc(200);
	strcpy(r, "test!");
	vc_gencmd(r, 200, s);

	vc_gencmd_stop();
	vchi_disconnect(vchi); // TODO: check error

	return r;
}
*/
import "C"
import "fmt"

// F 42
func F() int {
	// TODO: omg, C.free the resources please!
	x := C.GoString(C.f(C.CString("measure_temp")))
	fmt.Printf("%[1]T: %#[1]v\n", x)
	x = C.GoString(C.f(C.CString("measure_volts")))
	fmt.Printf("%[1]T: %#[1]v\n", x)
	return 24
}
