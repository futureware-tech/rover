/* A header file shared between C (.ino) software and Go interface
 * for BotBoarduino controller, which takes care of a few parts of
 * the Rover project. Notably, time-critical pulses (which are
 * required control servos and read DHT data) and analog reads
 * (ambient light sensor and battery).
 * More info about the board peripherals may be found in C (.ino)
 * file.
 */

enum {
  MinTilt = 30,
  MaxTilt = 150,
  I2CAddress = 0x42,
};

enum {
  // Invoke environment sensor
  CommandMeasureEnvironment,
  // Detach from all servos (release, low power mode)
  CommandSleep,
  // Re-attach servos
  CommandWake,
};

enum {
  // Command register
  ModuleCommand,
  // Board "readiness" bits (bit numbers are module IDs)
  ModuleBoard,
  ModuleLightSensor,
  ModuleArm,
  ModuleTilt,
  ModuleEnvironmentSensor,
  ModuleSpeech,
};

// Additions for EnvironmentSensor register
enum {
  ModuleEnvironmentSensorTemperature,
  ModuleEnvironmentSensorHumidity,
};

// Additions for Board register
enum {
  ModuleBoardStatus,
  ModuleBoardBattery,
};

// Additions for Arm register
enum {
  ModuleArmBasePan,
  ModuleArmBaseTilt,
  ModuleArmElbow,
  ModuleArmWristRotate,
  ModuleArmWristTilt,
  ModuleArmGrip,
};
