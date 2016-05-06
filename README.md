# rover

[![GoDoc](https://godoc.org/github.com/dasfoo/rover?status.svg)](http://godoc.org/github.com/dasfoo/rover)
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

## configure WiFi via SD card

See [Wireless CLI](
https://www.raspberrypi.org/documentation/configuration/wireless/wireless-cli.md)

## raspi-config

In `# raspi-config`, use "Expand Filesystem" option to use the whole SD card.
Right there as well, go to "Advanced Options" and enable "I2C", and agree to
kernel module autoload.

## save SD space and package installation time

See `Documentation` section of [Reducing Disk Footprint](
https://wiki.ubuntu.com/ReducingDiskFootprint#Documentation)

## `/boot/config.txt`

* set GPU mem allocation to 128MB (required for camera): `gpu_mem=128`
* Enable i2c: `dtparam=i2c_arm=on`
* Enable audio: `dtparam=audio=on`
* Enable camera: `start_x=1`
* Disable camera LED: `disable_camera_led=1`

## `/boot/cmdline.txt`

* Remove a `console=` reference to `serial0` (this is a UART interface) to allow
  normal SIM800 module interactions

## packages

`# apt install git pptp-linux socat`

## $HOME (/home/pi)

`$ git clone https://github.com/dasfoo/rover.git`

## systemd

* Do not use UART as console: `systemctl disable serial-getty@ttyAMA0.service`
* Add services:

  ```
  # ln -sf /home/pi/rover/systemd/<svc>.service \
    /etc/systemd/system/multi-user.target.wants/<svc@optional_args>.service`
  ```
    - reverse tunnel for ssh: `reverse-tunnel@<remote_host>:22.service`.
      Make sure `ssh <remote_host>` works (doesn't ask for host key / password)
    - autoswitch sim800 and wlan0: `autoswitch-wlan-sim800@<isp>.service`.
    - camera streaming server: `capture-server@<port>.service`.
      Don't forget to set `CAPTURE_PASSWORD` in `$HOME/.config/capture-server`.
    - camera streaming reverse tunnel (underscore means it's open to the world):
      `reverse-tunnel@<remote_host>:<port>_.service`
    - rover API server: `rover@<port>.service`
    - reverse tunnel for ssh, camera streaming server, and rover API server:
      `reverse-tunnel@<remote_host>.service` and its config (3 examples):
      
      ```
      $ cat $HOME/.config/reverse-tunnel/<remote_host>
      <remote_ssh_port> 22
      :<remote_camera_server_port> <local_camera_server_port>
      :<local_and_remote_rover_API_port>
      ```

## `/etc/fstab`

Mount /tmp as tmpfs (RAM):

`tmpfs /tmp tmpfs defaults,noatime,nosuid,size=50m 0 0`

## groups

Optional, as systemd runs the script as `root` now.

`pi` must be a member of `dip` group to call for `pon`/`poff`:

`# usermod -aG dip pi`

## ppp

Replace `$ISPNAME` with what's your ISP suggests as "endpoint", e.g.:
data.umts.example.org

```
$ cat /etc/ppp/peers/$ISPNAME
# serial path
/dev/ttyAMA0
# baud rate
115200
connect '/usr/sbin/chat -v -f /etc/chatscripts/gprs -T $ISPNAME'

# do not require auth from remote side
noauth
# add default route
defaultroute
# even if there's already default route, replace it with PPP
replacedefaultroute
# ask the peer for up to 2 DNS servers
usepeerdns
# name the interface pppN, N=0
unit 0
# dial again when connection is lost
persist
# rechallenge the peer every 321 seconds
chap-interval 321
# random string to identify connection
ipparam $ISPNAME
# send echo every 20 seconds
lcp-echo-interval 20
# reconnect if not responded to 3 echoes in a row
lcp-echo-failure 3
# do not try to guess own IP address (only receive it from ISP)
noipdefault
# do not try to guess remote peer IP address (only receive it from ISP)
noremoteip
# detach from controlling terminal only once connection is established
updetach

# (optionally) disable compression
#nopcomp
#novjccomp
#nobsdcomp
#nodeflate
#noaccomp
```

To test: power on SIM800 and `# pon $ISPNAME`.

## sshd

It might be useful to add `PasswordAuthentication no` to your
`/etc/ssh/sshd_config` if logging in to Pi with a key.

Removing password from `pi` user might make debugging network issues
(by plugging in a keyboard) a bit harder.

## backup

Once everything is configured, power Pi off (`# poweroff`) and unplug the SD
card. Then plug it into your computer, find out dev node for it (e.g.
`/dev/mmcblk0` or `/dev/rdisk2`) and make a backup:

`# dd if=/dev/rdisk2 of=rover_backup.img bs=1m`
