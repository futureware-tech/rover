/*
 * reset is done from outside (RPi)

 * 0x00 [w 8 bit] command
   * 0x02 stop motors
   + 0x06 read humidity

 * 0x01 [r 16 bit] ready status
   * bit 0 reserved
   + bit 1 board
   * bit 2 motor
   + bit 3 light sensor
   * bit 4 arm
   + bit 5 pan & tilt
   + bit 6 humidity sensor
   * bit 7 speech recognition

 * 0x02 [r 16 bit] battery status

 * 0x2x [rw] motors & encoders

 + 0x30 [r 16 bit] light sensor

 * 0x4x [w] arm

 + 0x50 [w 8 bit] pan, degrees 0-180
 + 0x51 [w 8 bit] tilt, degrees 0-MAX_TILT

 + 0x60 [r 8 bit] temperature, C
 + 0x61 [r 8 bit] humidity, %

 * 0x7x [rw] speech recognition
 */

#define MAX_TILT 125

#define I2C_ADDRESS 0x42

#define MODULE_COMMAND 0
#define MODULE_BOARD 1
#define MODULE_MOTOR 2
#define MODULE_LIGHT_SENSOR 3
#define MODULE_ARM 4
#define MODULE_PAN_TILT 5
#define MODULE_ENVIRONMENT_SENSOR 6
#define MODULE_SPEECH 7
