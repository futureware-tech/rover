/*
   tiny_gpio.c
   2015-09-12
   Public Domain

   Slightly modified tiny_gpio from http://abyz.co.uk/rpi/pigpio/examples.html
*/

#include "gpio.h"

#include <stdio.h>
#include <unistd.h>
#include <string.h>
#include <fcntl.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <time.h>

#define GPSET0 7
#define GPCLR0 10
#define GPLEV0 13

#define GPPUD     37
#define GPPUDCLK0 38
#define GPPUDCLK1 39

static volatile uint32_t  *gpioReg = MAP_FAILED;

#define PI_BANK (gpio>>5)
#define PI_BIT  (1<<(gpio&0x1F))

void gpioSetMode(unsigned gpio, unsigned mode)
{
   int reg, shift;

   reg   =  gpio/10;
   shift = (gpio%10) * 3;

   gpioReg[reg] = (gpioReg[reg] & ~(7<<shift)) | (mode<<shift);
}

int gpioGetMode(unsigned gpio)
{
   int reg, shift;

   reg   =  gpio/10;
   shift = (gpio%10) * 3;

   return (*(gpioReg + reg) >> shift) & 7;
}

void gpioSetPullUpDown(unsigned gpio, unsigned pud)
{
   *(gpioReg + GPPUD) = pud;

   usleep(20);

   *(gpioReg + GPPUDCLK0 + PI_BANK) = PI_BIT;

   usleep(20);

   *(gpioReg + GPPUD) = 0;

   *(gpioReg + GPPUDCLK0 + PI_BANK) = 0;
}

int gpioRead(unsigned gpio)
{
   if ((*(gpioReg + GPLEV0 + PI_BANK) & PI_BIT) != 0) return 1;
   else                                         return 0;
}

void gpioWrite(unsigned gpio, unsigned level)
{
   if (level == 0) *(gpioReg + GPCLR0 + PI_BANK) = PI_BIT;
   else            *(gpioReg + GPSET0 + PI_BANK) = PI_BIT;
}

void gpioTrigger(unsigned gpio, unsigned pulse_usec, unsigned level)
{
   if (level == 0) *(gpioReg + GPCLR0 + PI_BANK) = PI_BIT;
   else            *(gpioReg + GPSET0 + PI_BANK) = PI_BIT;

   usleep(pulse_usec);

   if (level != 0) *(gpioReg + GPCLR0 + PI_BANK) = PI_BIT;
   else            *(gpioReg + GPSET0 + PI_BANK) = PI_BIT;
}

#define CLOCK_KIND CLOCK_MONOTONIC
//#define CLOCK_KIND CLOCK_REALTIME

#define time_nsec(t) ((uint64_t)((t).tv_sec)*1000*1000*1000 + (t).tv_nsec)

uint64_t gpioReadPulse(unsigned gpio, uint64_t timeout_usec, int value) {
	struct timespec start_t, now_t;
	uint64_t timeout_nsec = timeout_usec * 1000, pulse_duration = 0;

	clock_gettime(CLOCK_KIND, &start_t);
	while (gpioRead(gpio) == value) {
		clock_gettime(CLOCK_KIND, &now_t);
		pulse_duration = time_nsec(now_t) - time_nsec(start_t);
		if (pulse_duration >= timeout_nsec) {
			break;
		}
	}
	return pulse_duration / 1000;
}

int gpioReadPulses(unsigned gpio, uint64_t timeout_usec, unsigned count, uint8_t *pulses) {
	unsigned offset, bit, i;
	uint64_t pivot, pulse;
	for (i = 0; i < count; ++i) {
		if ((pivot = gpioReadPulse(gpio, timeout_usec, 0)) >= timeout_usec) {
			return 0;
		}
		if ((pulse = gpioReadPulse(gpio, timeout_usec, 1)) >= timeout_usec) {
			return 0;
		}
		offset = i >> 3;
		bit = 7 - (i & 7);
		pulses[offset] |= (pulse > pivot) << bit;
	}
	return 1;
}

int gpioInitialise(void)
{
   int fd;

   fd = open("/dev/gpiomem", O_RDWR | O_SYNC) ;

   if (fd < 0)
   {
      // TODO: return proper error code
      fprintf(stderr, "failed to open /dev/gpiomem\n");
      return -1;
   }

   gpioReg = (uint32_t *)mmap(NULL, 0xB4, PROT_READ|PROT_WRITE, MAP_SHARED, fd, 0);

   close(fd);

   if (gpioReg == MAP_FAILED)
   {
      fprintf(stderr, "Bad, mmap failed\n");
      return -1;
   }
   return 0;
}
