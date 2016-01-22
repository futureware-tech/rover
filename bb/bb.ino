#include <Wire.h>
#include <DHT.h>
#include <Servo.h>

#include "bb.h"

Servo pan, tilt;

DHT environment_sensor(4, DHT11);
byte environment_temperature, environment_humidity;

byte light_sensor = 3;

byte battery_status = 2;

boolean halt = false;

byte i2cRegister = 0xff; // register to read from / write to

#define REGISTER(module) ((MODULE_ ## module) * 0x10)

uint16_t status = 0;

#define BUSY(module) (status = status & ~(MODULE_ ## module))
#define READY(module) (status |= (1 << MODULE_ ## module))

#ifdef DEBUG
#  define Log Serial.println
#else
#  define Log(...)
#endif

void setup() {
#ifdef DEBUG
  Serial.begin(9600);
#endif

  Wire.begin(I2C_ADDRESS); // join i2c channel
  Wire.onRequest(i2cRequest);
  Wire.onReceive(i2cReceive);
  READY(BOARD);

  environment_sensor.begin();
  // READY(ENVIRONMENT_SENSOR);
  // Do not make it free until the first measurement.

  pan.attach(13);
  tilt.attach(12);
  READY(PAN_TILT);

  READY(LIGHT_SENSOR);

  Log(F("Booted"));
}

void loop() {
  delay(100);
}

void boardCommand(byte value) {
  switch (value) {
  case MODULE_ENVIRONMENT_SENSOR:
    BUSY(ENVIRONMENT_SENSOR);
    environment_temperature = (byte)environment_sensor.readTemperature();
    environment_humidity = (byte)environment_sensor.readHumidity();
    READY(ENVIRONMENT_SENSOR);
    break;
  case MODULE_MOTOR:
    halt = true;
    break;
  default:
    Log(F("Command: unknown"));
    Log(value);
  }
}

void i2cReceive(int count) {
  i2cRegister = Wire.read();
  if (!Wire.available()) {
    return;
  }
  byte value8 = Wire.read();

  switch (i2cRegister) {
  case REGISTER(COMMAND):
    boardCommand(value8);
    break;
  case REGISTER(PAN_TILT):
    BUSY(PAN_TILT);
    pan.write(value8);
    delay(20);
    READY(PAN_TILT);
    break;
  case REGISTER(PAN_TILT) + 1:
    if (value8 < MAX_TILT) {
      BUSY(PAN_TILT);
      tilt.write(value8);
      delay(20);
      READY(PAN_TILT);
    }
    break;
  default:
    Log(F("Write: unknown register"));
    Log(i2cRegister);
  }
  while (Wire.available()) {
    Log(F("Write: extra byte:"));
    Log(Wire.read());
  }
}

void i2cRequest() {
  int value;
  switch (i2cRegister) {
  case REGISTER(BOARD):
    Wire.write((byte *)(&status), 2);
    break;
  case REGISTER(BOARD) + 1:
    value = analogRead(battery_status);
    Wire.write((byte *)(&value), 2);
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
  default:
    Log(F("Read: unknown register"));
    Log(i2cRegister);
  }
}
