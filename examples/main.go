package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dasfoo/bright-pi"
	"github.com/dasfoo/i2c"
	"github.com/dasfoo/lidar-lite-v2"
	"github.com/dasfoo/rover/bb"
	"github.com/dasfoo/rover/mc"
)

func ledCircle(b *bpi.BrightPI, circles int) {
	leds := []byte{bpi.WhiteTopLeft, bpi.WhiteTopRight, bpi.WhiteBottomRight, bpi.WhiteBottomLeft}
	for i := 0; i < circles; i++ {
		for _, led := range leds {
			_ = b.Power(led)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func park(b *bb.BB) {
	_ = b.ArmBasePan(90)
	_ = b.ArmBaseTilt(160)
	_ = b.ArmElbow(150)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	bus, err := i2c.NewBus(1)
	if err != nil {
		log.Println(err)
		return
	}

	bus.Log = func(string, ...interface{}) {}

	board := bb.NewBB(bus, bb.Address)
	_ = board.Wake()

	if s, e := board.GetStatus(); e == nil {
		fmt.Printf("Status bits: %.16b\n", s)
	}

	for i := 0; i < 10; i++ {
		if p, e := board.GetBatteryPercentage(); e != nil {
			board.Reset(bb.ResetPin)
		} else {
			fmt.Println("Battery status (estimated):", p, "%")
		}
		time.Sleep(time.Second)
	}

	if l, e := board.GetAmbientLight(); e == nil {
		fmt.Println("Brightness (0-1023):", l)
	}

	mco := mc.NewMC(bus, mc.Address)
	_ = mco.Wake()

	fmt.Println("Encoders (LF LB RF RB):")
	fmt.Println(mco.ReadEncoder(mc.EncoderLeftFront))
	fmt.Println(mco.ReadEncoder(mc.EncoderLeftBack))
	fmt.Println(mco.ReadEncoder(mc.EncoderRightFront))
	fmt.Println(mco.ReadEncoder(mc.EncoderRightBack))

	_ = mco.Left(mc.MaxSpeed / 2)
	time.Sleep(300 * time.Millisecond)
	_ = mco.Left(0)
	time.Sleep(time.Second)
	_ = mco.Right(-mc.MaxSpeed / 2)
	time.Sleep(300 * time.Millisecond)
	_ = mco.Right(0)
	time.Sleep(time.Second)

	fmt.Println("Encoders (LF LB RF RB) (after move: LEFT FWD, RIGHT REV):")
	fmt.Println(mco.ReadEncoder(mc.EncoderLeftFront))
	fmt.Println(mco.ReadEncoder(mc.EncoderLeftBack))
	fmt.Println(mco.ReadEncoder(mc.EncoderRightFront))
	fmt.Println(mco.ReadEncoder(mc.EncoderRightBack))

	_ = mco.Right(mc.MaxSpeed / 2)
	time.Sleep(300 * time.Millisecond)
	_ = mco.Right(0)
	time.Sleep(time.Second)
	_ = mco.Left(-mc.MaxSpeed / 2)
	time.Sleep(300 * time.Millisecond)
	_ = mco.Left(0)

	fmt.Println("Encoders (LF LB RF RB) (after move: LEFT REV, RIGHT FWD):")
	fmt.Println(mco.ReadEncoder(mc.EncoderLeftFront))
	fmt.Println(mco.ReadEncoder(mc.EncoderLeftBack))
	fmt.Println(mco.ReadEncoder(mc.EncoderRightFront))
	fmt.Println(mco.ReadEncoder(mc.EncoderRightBack))

	_ = board.ArmWristRotate(90)
	_ = board.ArmGrip(0)
	time.Sleep(time.Second)
	for i := 0; i < 3; i++ {
		_ = board.ArmBaseTilt(90)
		_ = board.ArmElbow(90)
		_ = board.ArmBasePan(45)
		time.Sleep(time.Second)

		_ = board.Tilt(45)
		time.Sleep(time.Second)
		_ = board.Tilt(90)

		_ = board.ArmBaseTilt(45)
		_ = board.ArmElbow(90)
		_ = board.ArmWristTilt(45)
		time.Sleep(time.Second)

		_ = board.ArmGrip(120)
		time.Sleep(time.Second)

		_ = board.ArmBaseTilt(90)
		_ = board.ArmElbow(90)
		time.Sleep(time.Second)

		_ = board.ArmBasePan(90)
		time.Sleep(time.Second)

		_ = board.ArmBaseTilt(45)
		_ = board.ArmElbow(90)
		_ = board.ArmWristTilt(45)
		_ = board.ArmGrip(0)
	}

	if t, h, e := board.GetTemperatureAndHumidity(); e == nil {
		fmt.Println("Temperature", t, "*C, humidity", h, "%")
	}

	b := bpi.NewBrightPI(bus, bpi.DefaultAddress)
	s := lidar.NewLidar(bus, lidar.DefaultAddress)
	_ = s.Reset()

	_ = b.Dim(bpi.WhiteAll, bpi.DefaultDim)
	_ = b.Gain(bpi.DefaultGain)
	go ledCircle(b, 2)

	// Simple GetDistance
	fmt.Println(s.GetDistance())

	// Continuous mode
	/*_ = s.SetContinuousMode(50, 100*time.Millisecond)
	_ = s.Acquire(false)
	for i := 0; i < 50; i++ {
		fmt.Println("Round", i)
		fmt.Println(s.ReadDistance())
		fmt.Println(s.ReadVelocity())
		time.Sleep(100 * time.Millisecond)
	}*/

	if hw, sw, err := s.GetVersion(); err == nil {
		fmt.Printf("LIDAR-Lite v2 \"Blue Label\" hw%dsw%d\n", hw, sw)
	}

	time.Sleep(time.Second)
	park(board)
	time.Sleep(time.Second)

	// Put devices in low power consumption mode
	_ = s.Sleep()
	_ = b.Sleep()
	_ = mco.Sleep()
	_ = board.Sleep()
}
