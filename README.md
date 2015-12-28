# rover

[![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)
[![Build Status](https://travis-ci.org/dasfoo/rover.svg?branch=master)](https://travis-ci.org/dasfoo/rover)
[![Coverage Status](https://coveralls.io/repos/dasfoo/rover/badge.svg?branch=master&service=github)](https://coveralls.io/github/dasfoo/rover?branch=master)

# golang cross-compiling with CGo

* https://medium.com/@rakyll/go-1-5-cross-compilation-488092ba44ec
* https://wiki.debian.org/RaspberryPi (from here we learn that Pi is "armel")
* https://wiki.debian.org/CrossToolchains

# Raspberry PI (A+ V1.1) preparations

## OS

[Raspbian](https://www.raspbian.org/)

## `/boot/config.txt`

* Drop GPU mem allocation: `gpu_mem=16`
* Enable i2c: `dtparam=i2c_arm=on`
* Enable audio: `dtparam=audio=on`
* Enable camera: `start_x=1`
* Disable camera LED: `disable_camera_led=1`

## `/boot/cmdline.txt`

* Remove a `console=` reference to `ttyAMA0` (this is a UART interface) to allow normal SIM800 module interactions

## systemd

* Do not use UART as console: `systemctl disable serial-getty@ttyAMA0.service`

## `/etc/fstab`

Mount /tmp as tmpfs (RAM):

`tmpfs /tmp tmpfs defaults,noatime,nosuid,size=50m 0 0`

## $HOME

`git clone https://github.com/dasfoo/rover.git`

## crontab

There is a number of ways to run commands at startup, crontab seems most portable:

`@reboot sh -c 'REMOTE_PORT=XXXXXX REMOTE_HOST=YYYYYY $HOME/rover/bin/reverse-tunnel >/dev/null 2>&1 &'`
