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

## `/boot/config.txt`

* ~~Drop GPU mem allocation to 16MB: `gpu_mem=16`~~ enabling camera requires at least 128MB GPU
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

## groups

`pi` must be a member of `dip` group to call for `pon`/`poff`:

`sudo usermod -aG dip pi`

## ppp

Replace `$ISPNAME` with what's your ISP suggests as "endpoint", e.g.: data.umts.example.org

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

## ssh

If planning to use `reverse-tunnel` or friends, make sure `ssh $REMOTE_HOST` works (e.g. doesn't
ask for host key or password).
