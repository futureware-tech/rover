#include <stdint.h>

/* gpio modes. */

#define PI_INPUT  0
#define PI_OUTPUT 1
#define PI_ALT0   4
#define PI_ALT1   5
#define PI_ALT2   6
#define PI_ALT3   7
#define PI_ALT4   3
#define PI_ALT5   2

void gpioSetMode(unsigned, unsigned);
int gpioGetMode(unsigned);

/* Values for pull-ups/downs off, pull-down and pull-up. */

#define PI_PUD_OFF  0
#define PI_PUD_DOWN 1
#define PI_PUD_UP   2

void gpioSetPullUpDown(unsigned, unsigned);
int gpioRead(unsigned);
void gpioWrite(unsigned, unsigned);
void gpioTrigger(unsigned, unsigned, unsigned);

uint64_t gpioReadPulse(unsigned, uint64_t, int);
int gpioReadPulses(unsigned, uint64_t, unsigned, uint8_t *);

int gpioInitialise(void);
