#include <host_applications/linux/libs/bcm_host/include/bcm_host.h>
#include <stdlib.h>
#include <errno.h>

#include "vcgencmd.h"

static VCHI_INSTANCE_T vchi;
static VCHI_CONNECTION_T *vchi_connections;

static char response[4096];

void Start() {
	bcm_host_init();
	vcos_init();
	if ((errno = vchi_initialise(&vchi)) != 0) {
		return;
	}

	if (errno = vchi_connect(NULL, 0, vchi) != 0) {
		return;
	}

	vc_vchi_gencmd_init(vchi, &vchi_connections, 1);
	errno = 0;
}

void Stop() {
	vc_gencmd_stop();
	if ((errno = vchi_disconnect(vchi)) != 0) {
		return;
	}
	errno = 0;
}

const char *Send(const char *command) {
	if ((errno = vc_gencmd(response, sizeof(response), command)) != 0) {
		return NULL;
	}
	return response;
}
