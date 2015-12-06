package drivers

import (
	"bytes"
	"fmt"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

type DHT11 struct {
	pin         embd.DigitalPin
	Temperature byte
	Humidity    byte
}

func NewDHT11(pin embd.DigitalPin) *DHT11 {
	return &DHT11{pin: pin}
}

func (s *DHT11) readOnce() error {
	if e := s.pin.SetDirection(embd.Out); e != nil {
		return e
	}
	if e := s.pin.Write(embd.High); e != nil {
		return e
	}
	time.Sleep(500 * time.Millisecond)
	if e := s.pin.Write(embd.Low); e != nil {
		return e
	}
	time.Sleep(18 * time.Millisecond)
	if e := s.pin.SetDirection(embd.In); e != nil {
		return e
	}

	var values [5]byte

	pulses := make([]byte, (len(values)*8+1)*10)
	//started_at := time.Now()
	for i := 0; i < len(pulses); i++ {
		if pulse, e := s.pin.Read(); e != nil {
			return e
		} else {
			pulses[i] = byte(pulse)
		}
	}
	//read_time := time.Since(started_at).Seconds() * 1000000 / float64(len(pulses))

	pulse_duration := 0
	for i, j := bytes.LastIndexByte(pulses, 0), 0; i >= 0 && j < 8*len(values); i-- {
		if pulses[i] == 1 {
			pulse_duration++
		} else {
			if pulse_duration > 0 {
				value := 0
				if pulse_duration > 2 /* *read_time > 70 */ {
					value = 1
				}
				pulse_duration = 0
				values[j/8] += byte(value << byte(j%8))
				j++
			}
		}
	}

	checksum := values[4] + values[3] + values[2] + values[1]
	if checksum == values[0] {
		s.Temperature = values[2]
		s.Humidity = values[4]
		return nil
	}
	return fmt.Errorf("invalid checksum: expected %d, got %d",
		values[0], checksum)
}

func (s *DHT11) Read(max_retries int) (retries int, e error) {
	for retries = 0; retries <= max_retries; retries++ {
		if e = s.readOnce(); e == nil {
			return
		}
	}
	return
}
