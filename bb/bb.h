enum {
  MaxTilt = 124,
  I2CAddress = 0x42,
};

enum {
  // Stop all movement
  CommandHalt,
  // Invoke environment sensor
  CommandMeasureEnvironment,
  // Detach from all servos / motors (release, low power mode)
  CommandSleep,
  // Re-attach servos / motors
  CommandWake,
  // Use encoder statuses to detect movement and block it with motors
  CommandBrake,
  // Release the brake mode
  CommandReleaseBrake,
};

enum {
  // Command register
  ModuleCommand,
  // Board "readiness" bits (bit numbers are module IDs)
  ModuleBoard,
  ModuleMotor,
  ModuleLightSensor,
  ModuleArm,
  ModulePanTilt,
  ModuleEnvironmentSensor,
  ModuleSpeech,
};

// Additions for Motor register
enum {
  ModuleMotorLeft,
  ModuleMotorRight,
  ModuleMotorEncoderLeftFront,
  ModuleMotorEncoderLeftBack,
  ModuleMotorEncoderRightFront,
  ModuleMotorEncoderRightBack,
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

// Additions for PanTilt register
enum {
  ModulePanTiltPan,
  ModulePanTiltTilt,
};
