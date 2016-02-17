#include <Wire.h>
#include <DHT.h>
#include <Servo.h>
#include <math.h>

#include "bb.h"
#include "music.h"

/*
 * BotBoarduino PINs
 * According to http://www.lynxmotion.com/images/html/build185.htm and some others.
 * Servos: PWM vs PPM: http://forum.arduino.cc/index.php/topic,14146.0.html
 */

enum {
  PinSerialRX = 0,
  PinSerialTX = 1,
  PinReserved0 = 2,
  PinReserved1 = 3,      // PWM
  PinEnvironmentSensor = 4,
  PinSpeaker = 5,        // speaker enabled by jumper, PWM/timer
  PinArmBasePan = 6,     // PWM/timer
  PinArmBaseTilt = 7,    // A LED enabled by jumper, A button
  PinArmElbow = 8,       // B LED enabled by jumper, B button
  PinArmWristRotate = 9, // C LED enabled by jumper, C button, PWM
  PinArmWristTilt = 10,  // PWM
  PinArmGrip = 11,       // PWM
  PinTilt = 12,
  PinAwakeLed = 13,      // "L" LED
};

enum {
  AnalogPinBattery = 2,
  AnalogPinLightSensor = 3,
  AnalogPinI2CSDA = 4, // i2c enabled by jumper (blue wire)
  AnalogPinI2CSCL = 5, // i2c enabled by jumper (green wire)
};

Servo Tilt,
      ArmBasePan, ArmBaseTilt,
      ArmElbow, ArmGrip,
      ArmWristRotate, ArmWristTilt;

DHT EnvironmentSensor(PinEnvironmentSensor, DHT11);
volatile byte environment_temperature, environment_humidity;

volatile byte i2cRegister = 0xff; // register to read from / write to

#define MODULE_REGISTER(module) ((Module ## module) * 0x10)

volatile uint16_t status = 0;
#define MODULE_ISREADY(module) (status & (1 << Module ## module))
#define MODULE_BUSY(module, ifnotready) { \
  if (!MODULE_ISREADY(module)) { \
    { ifnotready; } \
  } \
  (status = status & ~(1 << Module ## module)); \
}
#define MODULE_READY(module) (status |= (1 << Module ## module))

void attachArm(boolean attach) {
  if (attach) {
    ArmBasePan.attach(PinArmBasePan);
    ArmBaseTilt.attach(PinArmBaseTilt);
    ArmElbow.attach(PinArmElbow);
    ArmWristRotate.attach(PinArmWristRotate);
    ArmWristTilt.attach(PinArmWristTilt);
    ArmGrip.attach(PinArmGrip);
    MODULE_READY(Arm);
  } else {
    // do not "return" if not ready, because we can detach at any moment
    MODULE_BUSY(Arm,);
    ArmBasePan.detach();
    ArmBaseTilt.detach();
    ArmElbow.detach();
    ArmWristRotate.detach();
    ArmWristTilt.detach();
    ArmGrip.detach();
  }
}

void attachTilt(boolean attach) {
  if (attach) {
    Tilt.attach(PinTilt);
    MODULE_READY(Tilt);
  } else {
    // do not "return" if not ready, because we can detach at any moment
    MODULE_BUSY(Tilt,);
    Tilt.detach();
  }
}

void setup() {
  Wire.begin(I2CAddress); // join i2c channel
  Wire.onReceive(i2cReceive);
  MODULE_READY(Command);
  Wire.onRequest(i2cRequest);
  MODULE_READY(Board);

  EnvironmentSensor.begin();
  // not: READY(EnvironmentSensor);
  // (do not make it ready until the first measurement)

  attachTilt(true);
  attachArm(true);

  MODULE_READY(LightSensor);

  //play(PinSpeaker, melody_HappyBirthday, sizeof(melody_HappyBirthday) >> 2);
}

void loop() {
  if (!MODULE_ISREADY(EnvironmentSensor)) {
    // DHT library uses delay() internally, which can't be used in interrupts
    environment_temperature = (byte)EnvironmentSensor.readTemperature();
    environment_humidity = (byte)EnvironmentSensor.readHumidity();
    MODULE_READY(EnvironmentSensor);
  }
  delay(100);
}

void boardCommand(byte value) {
  switch (value) {
  case CommandMeasureEnvironment:
    // an indicator for the main loop()
    MODULE_BUSY(EnvironmentSensor,);
    break;
  case CommandSleep:
    attachArm(false);
    attachTilt(false);
    break;
  case CommandWake:
    attachTilt(true);
    attachArm(true);
    break;
  }
}

void i2cReceive(int count) {
  i2cRegister = Wire.read();
  if (count == 1) {
    return;
  }
  byte value8 = Wire.read();

  switch (i2cRegister) {
  case MODULE_REGISTER(Command):
    boardCommand(value8);
    break;
  case MODULE_REGISTER(Tilt):
    Tilt.write(constrain(value8, MinTilt, MaxTilt));
    break;

#define SERVO_CASE_WITH_ADDITION(module, addition, value) \
  case MODULE_REGISTER(module) + Module ## module ## addition: \
    module ## addition.write(value); \
    break;

  SERVO_CASE_WITH_ADDITION(Arm, BasePan, value8)
  SERVO_CASE_WITH_ADDITION(Arm, BaseTilt, value8)
  SERVO_CASE_WITH_ADDITION(Arm, Elbow, value8)
  SERVO_CASE_WITH_ADDITION(Arm, WristRotate, value8)
  SERVO_CASE_WITH_ADDITION(Arm, WristTilt, value8)
  SERVO_CASE_WITH_ADDITION(Arm, Grip, value8)
  }
}

void writeWord(uint16_t value) {
  value = (value << 8) | (value >> 8);
  Wire.write((byte *)(&value), sizeof(value));
}

void i2cRequest() {
  int value;
  switch (i2cRegister) {
  case MODULE_REGISTER(Board) + ModuleBoardStatus:
    writeWord(status);
    break;
  case MODULE_REGISTER(Board) + ModuleBoardBattery:
    value = analogRead(AnalogPinBattery);
    Wire.write((byte)(constrain(map(value,
                                    (2.0 / 5.0 * 1023.0), (3.5 / 5.0 * 1023.0),
                                    0, 100), 0, 100)));
    break;
  case MODULE_REGISTER(LightSensor):
    value = analogRead(AnalogPinLightSensor);
    writeWord(value);
    break;
  case MODULE_REGISTER(EnvironmentSensor) + ModuleEnvironmentSensorTemperature:
    Wire.write(environment_temperature);
    break;
  case MODULE_REGISTER(EnvironmentSensor) + ModuleEnvironmentSensorHumidity:
    Wire.write(environment_humidity);
    break;
  }
}
