#include <Wire.h>
#include <DHT.h>
#include <Servo.h>
#include <math.h>

#include "bb.h"
#include "music.h"

Servo pan, tilt;
Servo left, right;

/*
 * BotBoarduino PINs
 * According to http://www.lynxmotion.com/images/html/build185.htm and some others.
 * Servos: PWM vs PPM: http://forum.arduino.cc/index.php/topic,14146.0.html
 */

#define PIN_SERIAL_RX           0
#define PIN_SERIAL_TX           1
#define PIN_MOTOR_LEFT          2
#define PIN_MOTOR_RIGHT         3 // PWM
#define PIN_ENVIRONMENT_SENSOR  4
#define PIN_SPEAKER             5 // speaker enabled by jumper, PWM/timer
#define PIN_RESERVED0           6 // PWM/timer
#define PIN_RESERVED1           7 // A LED enabled by jumper, A button
#define PIN_RESERVED2           8 // B LED enabled by jumper, B button
#define PIN_RESERVED3           9 // C LED enabled by jumper, C button, PWM
#define PIN_RESERVED4          10 // PWM
#define PIN_RESERVED5          11 // PWM
#define PIN_TILT               12
#define PIN_PAN                13 // "L" LED

#define PIN_ANALOG_RESERVED0    0
#define PIN_ANALOG_RESERVED1    1
#define PIN_ANALOG_BATTERY      2
#define PIN_ANALOG_LIGHT_SENSOR 3
#define PIN_ANALOG_I2C_SDA      4 // i2c enabled by jumper
#define PIN_ANALOG_I2C_SCL      5 // i2c enabled by jumper

// TODO: read these from Arduino Nano (protocol = ?)
long encoder_left_front = 0, encoder_left_back = 0,
     encoder_right_front = 0, encoder_right_back = 0;

DHT environment_sensor(PIN_ENVIRONMENT_SENSOR, DHT11);
byte environment_temperature, environment_humidity;

byte i2cRegister = 0xff; // register to read from / write to

#define REGISTER(module) ((MODULE_ ## module) * 0x10)

uint16_t status = 0;

#define ISREADY(module) (status & (1 << MODULE_ ## module))
#define BUSY(module, ifnotready) { \
  if (!ISREADY(module)) { \
    { ifnotready; } \
  } \
  (status = status & ~(1 << MODULE_ ## module)); \
}
#define READY(module) (status |= (1 << MODULE_ ## module))

void attachMotor(boolean attach) {
  if (attach) {
    left.attach(PIN_MOTOR_LEFT, 1000, 2000);
    right.attach(PIN_MOTOR_RIGHT, 1000, 2000);
    READY(MOTOR);
  } else {
    // do not "return" if not ready, because we can detach at any moment
    BUSY(MOTOR,);
    left.detach();
    right.detach();
  }
}

void attachArm(boolean attach) {
  (void)attach;
  // TODO: arm
}

void attachPanTilt(boolean attach) {
  if (attach) {
    pan.attach(PIN_PAN);
    READY(PAN);
    tilt.attach(PIN_TILT);
    READY(TILT);
  } else {
    // do not "return" if not ready, because we can detach at any moment
    BUSY(PAN,);
    pan.detach();
    BUSY(TILT,);
    tilt.detach();
  }
}

void setup() {
  Wire.begin(I2C_ADDRESS); // join i2c channel
  Wire.onReceive(i2cReceive);
  READY(COMMAND);
  Wire.onRequest(i2cRequest);
  READY(BOARD);

  environment_sensor.begin();
  // not: READY(ENVIRONMENT_SENSOR);
  // (do not make it ready until the first measurement)

  attachPanTilt(true);
  attachMotor(true);
  attachArm(true);

  READY(LIGHT_SENSOR);

  play(PIN_SPEAKER, melody_HappyBirthday, sizeof(melody_HappyBirthday) >> 2);
}

void loop() {
  if (!ISREADY(ENVIRONMENT_SENSOR)) {
    // DHT library uses delay() internally, which can't be used in interrupts
    environment_temperature = (byte)environment_sensor.readTemperature();
    environment_humidity = (byte)environment_sensor.readHumidity();
    READY(ENVIRONMENT_SENSOR);
  }
  delay(100);
}

void boardCommand(byte value) {
  switch (value) {
  case COMMAND_HALT:
    BUSY(MOTOR,);
    left.write(90);
    right.write(90);
    READY(MOTOR);

    BUSY(PAN,);
    pan.write(90);
    READY(PAN);

    BUSY(TILT,);
    tilt.write(90);
    READY(TILT);
    // TODO: halt arm servos
    break;
  case COMMAND_MEASURE_ENVIRONMENT:
    // an indicator for the main loop()
    BUSY(ENVIRONMENT_SENSOR,);
    break;
  case COMMAND_SLEEP:
    attachMotor(false);
    attachArm(false);
    attachPanTilt(false);
    break;
  case COMMAND_WAKE:
    attachPanTilt(true);
    attachArm(true);
    attachMotor(true);
    break;
  case COMMAND_BRAKE:
    // TODO: implementation
    break;
  case COMMAND_RELEASE_BRAKE:
    // TODO: implementation
    break;
  }
}

void i2cReceive(int count) {
  (void)count;

  // TODO: see if i2cReceive is called again when count is exhausted
  while (Wire.available()) {
    i2cRegister = Wire.read();
    if (!Wire.available()) {
      return;
    }
    byte value8 = Wire.read();

    switch (i2cRegister) {
    case REGISTER(COMMAND):
      boardCommand(value8);
      break;
    case REGISTER(PAN):
      BUSY(PAN, return);
      pan.write(value8);
      READY(PAN);
      break;
    case REGISTER(TILT):
      BUSY(TILT, return);
      tilt.write(constrain(value8, 0, MAX_TILT));
      READY(TILT);
      break;
    case REGISTER(MOTOR) + MOTOR_LEFT:
      BUSY(MOTOR, return);
      left.write(value8);
      READY(MOTOR);
      break;
    case REGISTER(MOTOR) + MOTOR_RIGHT:
      BUSY(MOTOR, return)
      right.write(value8);
      READY(MOTOR);
      break;
    case REGISTER(MOTOR) + MOTOR_ENCODER_LEFT_FRONT:
      // TODO: implementation
      break;
    // TODO: REGISTER(MOTOR) + MOTOR_ENCODER_{LEFT,RIGHT}_{BACK,FRONT}
    }
  }
}

void writeWord(uint16_t value) {
  value = (value << 8) | (value >> 8);
  Wire.write((byte *)(&value), sizeof(value));
}

void i2cRequest() {
  int value;
  switch (i2cRegister) {
  case REGISTER(BOARD) + BOARD_STATUS:
    writeWord(status);
    break;
  case REGISTER(BOARD) + BOARD_BATTERY:
    // centiV = analog * 1.7581
    // Range = 10.6V .. 12.6V
    value = analogRead(PIN_ANALOG_BATTERY);
    Wire.write((byte)(value * 1.7581 / 2));
    break;
  case REGISTER(LIGHT_SENSOR):
    value = analogRead(PIN_ANALOG_LIGHT_SENSOR);
    writeWord(value);
    break;
  case REGISTER(ENVIRONMENT_SENSOR) + ENVIRONMENT_SENSOR_TEMPERATURE:
    Wire.write(environment_temperature);
    break;
  case REGISTER(ENVIRONMENT_SENSOR) + ENVIRONMENT_SENSOR_HUMIDITY:
    Wire.write(environment_humidity);
    break;
  // TODO: send big endian over the network?
  case REGISTER(MOTOR) + MOTOR_ENCODER_LEFT_FRONT:
    Wire.write((byte *)(&encoder_left_front), sizeof(encoder_left_front));
    break;
  case REGISTER(MOTOR) + MOTOR_ENCODER_LEFT_BACK:
    Wire.write((byte *)(&encoder_left_back), sizeof(encoder_left_back));
    break;
  case REGISTER(MOTOR) + MOTOR_ENCODER_RIGHT_FRONT:
    Wire.write((byte *)(&encoder_right_front), sizeof(encoder_right_front));
    break;
  case REGISTER(MOTOR) + MOTOR_ENCODER_RIGHT_BACK:
    Wire.write((byte *)(&encoder_right_back), sizeof(encoder_right_back));
    break;
  }
}
