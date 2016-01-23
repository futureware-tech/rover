#define MAX_TILT 124

#define I2C_ADDRESS 0x42

// Stop all movement
#define COMMAND_HALT                0x00
// Invoke environment sensor
#define COMMAND_MEASURE_ENVIRONMENT 0x01
// Detach from all servos / motors (release, low power mode)
#define COMMAND_SLEEP               0x02
// Re-attach servos / motors
#define COMMAND_WAKE                0x03

// Command register
#define MODULE_COMMAND            0x00
// Board "readiness" bits (bit numbers are module IDs)
#define MODULE_BOARD              0x01
#define MODULE_MOTOR              0x02
#define MODULE_LIGHT_SENSOR       0x03
#define MODULE_ARM_BASE_ROTATE    0x04
#define MODULE_ARM_RESERVED0      0x05
#define MODULE_ARM_RESERVED1      0x06
#define MODULE_ARM_RESERVED2      0x07
#define MODULE_ARM_WRIST_ROTATE   0x08
#define MODULE_ARM_WRIST          0x09
#define MODULE_PAN                0x0a
#define MODULE_TILT               0x0b
#define MODULE_ENVIRONMENT_SENSOR 0x0c
#define MODULE_SPEECH             0x0d
#define MODULE_RESERVED0          0x0e
#define MODULE_RESERVED1          0x0f
