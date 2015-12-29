package i2c

/*
#include <linux/i2c-dev.h>
*/
import "C"

import (
	"fmt"
	"os"
	"sync"
	"syscall"
)

// Bus is a type to access I2C bus
type Bus struct {
	file          *os.File
	opLock        sync.Mutex
	remoteAddress byte
}

// NewBus opens a Linux i2c bus file
func NewBus(id byte) (*Bus, error) {
	file, err := os.OpenFile(
		fmt.Sprintf("/dev/i2c-%d", id),
		os.O_RDWR,
		os.ModeExclusive)
	if err != nil {
		return nil, err
	}
	return &Bus{file: file}, nil
}

func (b *Bus) setRemoteAddress(addr byte) error {
	if addr != b.remoteAddress {
		if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, b.file.Fd(),
			C.I2C_SLAVE, uintptr(addr)); errno != 0 {
			return syscall.Errno(errno)
		}
		b.remoteAddress = addr
	}
	return nil
}

// ReadByteFromReg reads 1 byte from a register of a slave device
func (b *Bus) ReadByteFromReg(addr, reg byte) (byte, error) {
	b.opLock.Lock()
	defer b.opLock.Unlock()
	if err := b.setRemoteAddress(addr); err != nil {
		return 0, err
	}
	value, err := C.i2c_smbus_read_byte_data(C.int(b.file.Fd()), C.__u8(reg))
	return byte(value), err
}

// ReadWordFromReg reads 2 bytes from a register of a slave device
func (b *Bus) ReadWordFromReg(addr, reg byte) (uint16, error) {
	b.opLock.Lock()
	defer b.opLock.Unlock()
	if err := b.setRemoteAddress(addr); err != nil {
		return 0, err
	}
	value, err := C.i2c_smbus_read_word_data(C.int(b.file.Fd()), C.__u8(reg))
	return uint16(value), err
}

// WriteByteToReg writes 1 byte to a register of a slave device
func (b *Bus) WriteByteToReg(addr, reg, value byte) error {
	b.opLock.Lock()
	defer b.opLock.Unlock()
	if err := b.setRemoteAddress(addr); err != nil {
		return err
	}
	_, err := C.i2c_smbus_write_byte_data(C.int(b.file.Fd()), C.__u8(reg),
		C.__u8(value))
	return err
}

// WriteWordToReg writes 2 bytes to a register of a slave device
func (b *Bus) WriteWordToReg(addr, reg byte, value uint16) error {
	b.opLock.Lock()
	defer b.opLock.Unlock()
	if err := b.setRemoteAddress(addr); err != nil {
		return err
	}
	_, err := C.i2c_smbus_write_word_data(C.int(b.file.Fd()), C.__u8(reg),
		C.__u16(value))
	return err
}

// Close frees any resources allocated for the bus
func (b *Bus) Close() error {
	b.opLock.Lock()
	defer b.opLock.Unlock()
	return b.file.Close()
}
