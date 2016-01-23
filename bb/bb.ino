#include <Wire.h>
#include <DHT.h>
#include <Servo.h>

#include "bb.h"
#include "music.h"

Servo pan, tilt;
Servo left, right;

DHT environment_sensor(4, DHT11);
byte environment_temperature, environment_humidity;

byte light_sensor = 3;

byte battery_status = 2;

byte i2cRegister = 0xff; // register to read from / write to

#define REGISTER(module) ((MODULE_ ## module) * 0x10)

uint16_t status = 0;

#define BUSY(module, ifnotready) { \
  if (!(status & (MODULE_ ## module))) { \
    { ifnotready; } \
  } \
  (status = status & ~(MODULE_ ## module)); \
}
#define READY(module) (status |= (1 << MODULE_ ## module))

void setupMotor() {
  left.attach(2, 1000, 2000);
  right.attach(3, 1000, 2000);
  READY(MOTOR);
}

void setupPanTilt() {
  pan.attach(13);
  READY(PAN);
  tilt.attach(12);
  READY(TILT);
}

void setup() {
  Wire.begin(I2C_ADDRESS); // join i2c channel
  Wire.onRequest(i2cRequest);
  Wire.onReceive(i2cReceive);
  READY(BOARD);

  environment_sensor.begin();
  // READY(ENVIRONMENT_SENSOR);
  // Do not make it ready until the first measurement.

  setupPanTilt();
  setupMotor();

  READY(LIGHT_SENSOR);

  play(melody_HappyBirthday, sizeof(melody_HappyBirthday) >> 2);
}

void loop() {
  delay(100);
}

void boardCommand(byte value) {
  switch (value) {
  case COMMAND_HALT:
    BUSY(MOTOR,);
    left.write(90);
    right.write(90);
    READY(MOTOR);
    break;
  case COMMAND_MEASURE_ENVIRONMENT:
    BUSY(ENVIRONMENT_SENSOR,); // do not "return" if not ready, because it's initial status
    environment_temperature = (byte)environment_sensor.readTemperature();
    environment_humidity = (byte)environment_sensor.readHumidity();
    READY(ENVIRONMENT_SENSOR);
    break;
  case COMMAND_SLEEP:
    BUSY(PAN,);
    pan.detach();
    BUSY(TILT,);
    tilt.detach();
    BUSY(MOTOR,);
    left.detach();
    right.detach();
    break;
  case COMMAND_WAKE:
    setupMotor();
    setupPanTilt();
    break;
  }
}

void i2cReceive(int count) {
  (void)count;

  // TODO: see if i2cReceive is called when count is exhausted
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
      delay(200);
      READY(PAN);
      break;
    case REGISTER(TILT):
      if (value8 <= MAX_TILT) {
        BUSY(TILT, return);
        tilt.write(value8);
        delay(200);
        READY(TILT);
      }
      break;
    case REGISTER(MOTOR):
      BUSY(MOTOR, return);
      left.write(value8);
      READY(MOTOR);
      break;
    case REGISTER(MOTOR) + 1:
      BUSY(MOTOR, return);
      right.write(value8);
      READY(MOTOR);
      break;
    }
  }
}

void i2cRequest() {
  int value;
  switch (i2cRegister) {
  case REGISTER(BOARD):
    Wire.write((byte *)(&status), 2);
    break;
  case REGISTER(BOARD) + 1:
    // V = analog / 56.88
    // Range = 10.6 .. 12.6
    value = analogRead(battery_status);
    Wire.write((byte)(((float)value / 56.88 - 10.6) * 50));
    break;
  case REGISTER(LIGHT_SENSOR):
    value = analogRead(light_sensor);
    Wire.write((byte *)(&value), 2);
    break;
  case REGISTER(ENVIRONMENT_SENSOR):
    Wire.write(environment_temperature);
    break;
  case REGISTER(ENVIRONMENT_SENSOR) + 1:
    Wire.write(environment_humidity);
    break;
  }
}
