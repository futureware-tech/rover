package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

func ReadDHT11(pin embd.DigitalPin) (byte, byte, error) {
	if e := pin.SetDirection(embd.Out); e != nil {
		return 0, 0, e
	}
	if e := pin.Write(embd.High); e != nil {
		return 0, 0, e
	}
	time.Sleep(500 * time.Millisecond)
	if e := pin.Write(embd.Low); e != nil {
		return 0, 0, e
	}
	time.Sleep(18 * time.Millisecond)
	if e := pin.SetDirection(embd.In); e != nil {
		return 0, 0, e
	}

	var values [5]byte

	pulses := make([]byte, (len(values)*8+1)*10)
	//started_at := time.Now()
	for i := 0; i < len(pulses); i++ {
		if pulse, e := pin.Read(); e != nil {
			return 0, 0, e
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
		return values[4], values[2], nil
	}
	return 0, 0, fmt.Errorf("invalid checksum: expected %d, got %d",
		values[0], checksum)
}

func main() {
	embd.InitGPIO()
	defer embd.CloseGPIO()

	if pin, e := embd.NewDigitalPin(4); e != nil {
		log.Panic("NewDigitalPin ", e)
	} else {
		defer pin.Close()
		for {
			if h, t, e := ReadDHT11(pin); e != nil {
				fmt.Println(e)
			} else {
				fmt.Printf("%d*C, %d%%\n", t, h)
			}
			time.Sleep(1 * time.Second)
		}
	}
}
