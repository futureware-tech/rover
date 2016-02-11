/* A header file shared between C (.ino) software and Go interface
 * for motors controller, which rules the motors and makes some
 * time-critical checks, like counting rotary encoder flips.
 * More info about the board peripherals may be found in C (.ino)
 * file.
 */

enum {
  I2CAddress = 0x43,
};

enum {
  // Use encoder statuses to detect movement and block it with motors
  CommandBrake,
  // Release the brake mode
  CommandReleaseBrake,
  // Detach from all motors (low power mode)
  CommandSleep,
  // Re-attach motors
  CommandWake,
};

// These values are not for use in external Go interface,
// it's just a reference for constants below.
enum {
  MinEncoderPin = 2,
  PinEncoderLeftFront1 = MinEncoderPin,
  PinEncoderLeftFront2,
  PinEncoderLeftBack1,
  PinEncoderLeftBack2,
  PinEncoderRightFront1,
  PinEncoderRightFront2,
  PinEncoderRightBack1,
  PinEncoderRightBack2,
  EncoderPins,
};

enum {
  RegisterCommand = 0,

  // 4-byte registers to read encoder values from
  RegisterEncoderLeftFront = ((PinEncoderLeftFront1 - MinEncoderPin) >> 1) + 1,
  RegisterEncoderLeftBack = ((PinEncoderLeftBack1 - MinEncoderPin) >> 1) + 1,
  RegisterEncoderRightFront = ((PinEncoderRightFront1 - MinEncoderPin) >> 1) + 1,
  RegisterEncoderRightBack = ((PinEncoderRightBack1 - MinEncoderPin) >> 1) + 1,

  RegisterMotorLeft,
  RegisterMotorRight,
};
