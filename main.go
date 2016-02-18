package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dasfoo/bright-pi"
	"github.com/dasfoo/i2c"
	"github.com/dasfoo/lidar-lite-v2"
	"github.com/dasfoo/rover/bb"
	"github.com/dasfoo/rover/mc"
)

/*func park(b *bb.BB) {
	_ = b.ArmBasePan(90)
	_ = b.ArmBaseTilt(160)
	_ = b.ArmElbow(150)
}*/

var (
	board  *bb.BB
	motors *mc.MC
	meter  *lidar.Lidar
	lights *bpi.BrightPI
)

type sleeper interface {
	Sleep() error
}

type waker interface {
	Wake() error
}

func sleepAwake(awake bool, devices ...interface{}) (changed int) {
	for _, device := range devices {
		if awake {
			if w, ok := device.(waker); ok {
				if e := w.Wake(); e == nil {
					changed++
				}
			}
		} else {
			if s, ok := device.(sleeper); ok {
				if e := s.Sleep(); e == nil {
					changed++
				}
			}
		}
	}
	return
}

func httpStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Status:")

	fmt.Fprint(w, "\tBattery: ")
	if batteryPercentage, e := board.GetBatteryPercentage(); e != nil {
		fmt.Fprintln(w, e)
	} else {
		fmt.Fprintf(w, "%d%%\n", batteryPercentage)
	}

	fmt.Fprint(w, "\tLight (0-1023): ")
	if light, e := board.GetAmbientLight(); e != nil {
		fmt.Fprintln(w, e)
	} else {
		fmt.Fprintf(w, "%d\n", light)
	}

	fmt.Fprintln(w, "\tEncoders:")
	fmt.Fprint(w, "\t\tLeft Front: ")
	if encoder, e := motors.ReadEncoder(mc.EncoderLeftFront); e != nil {
		fmt.Fprintln(w, e)
	} else {
		fmt.Fprintln(w, encoder)
	}
	fmt.Fprint(w, "\t\tLeft Back: ")
	if encoder, e := motors.ReadEncoder(mc.EncoderLeftBack); e != nil {
		fmt.Fprintln(w, e)
	} else {
		fmt.Fprintln(w, encoder)
	}
	fmt.Fprint(w, "\t\tRight Front: ")
	if encoder, e := motors.ReadEncoder(mc.EncoderRightFront); e != nil {
		fmt.Fprintln(w, e)
	} else {
		fmt.Fprintln(w, encoder)
	}
	fmt.Fprint(w, "\t\tRight Back: ")
	if encoder, e := motors.ReadEncoder(mc.EncoderRightBack); e != nil {
		fmt.Fprintln(w, e)
	} else {
		fmt.Fprintln(w, encoder)
	}

	fmt.Fprint(w, "\tEnvironment: ")
	if t, h, e := board.GetTemperatureAndHumidity(); e != nil {
		fmt.Fprintln(w, e)
	} else {
		fmt.Fprintln(w, "temperature", t, "*C, humidity", h, "%")
	}

	fmt.Fprint(w, "\tDistance: ")
	if d, e := meter.GetDistance(); e != nil {
		fmt.Fprintln(w, e)
	} else {
		fmt.Fprintln(w, d, "meters")
	}
}

func httpMove(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = motors.Left(0)
		_ = motors.Right(0)
	}()

	if e := r.ParseForm(); e != nil {
		fmt.Fprintln(w, e)
		return
	}

	var (
		left, right, duration int
		e                     error
	)
	if left, e = strconv.Atoi(r.FormValue("left")); e != nil {
		fmt.Fprintln(w, "Left:", e)
		return
	}
	if right, e = strconv.Atoi(r.FormValue("right")); e != nil {
		fmt.Fprintln(w, "Right:", e)
		return
	}
	if duration, e = strconv.Atoi(r.FormValue("duration")); e != nil {
		fmt.Fprintln(w, "Duration:", e)
		return
	}

	if e = motors.Left(int8(left)); e != nil {
		fmt.Fprintln(w, "Left:", e)
		return
	}
	if e = motors.Right(int8(right)); e != nil {
		fmt.Fprintln(w, "Right:", e)
		return
	}
	fmt.Fprint(w, "Started... ")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	time.Sleep(time.Duration(duration) * time.Millisecond)
	fmt.Fprintln(w, "Done")
}

func httpSleep(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Setting sleep (low power) mode: ")
	// Putting motors to sleep doesn't seem perfect yet - they may go crazy
	changed := sleepAwake(false, board, meter, lights)
	fmt.Fprintln(w, changed, "changed")
}

func httpWake(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Setting wake mode: ")
	// Putting motors to sleep doesn't seem perfect yet - they may go crazy
	changed := sleepAwake(true, board, meter, motors, lights)
	fmt.Fprintln(w, changed, "changed")
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	if bus, err := i2c.NewBus(1); err != nil {
		log.Fatal(err)
	} else {
		// Silence i2c bus log
		bus.Log = func(string, ...interface{}) {}

		board = bb.NewBB(bus, bb.Address)
		motors = mc.NewMC(bus, mc.Address)
		meter = lidar.NewLidar(bus, lidar.DefaultAddress)
		lights = bpi.NewBrightPI(bus, bpi.DefaultAddress)

		if s, e := board.GetStatus(); e == nil {
			fmt.Printf("Status bits: %.16b\n", s)
		}
	}

	http.HandleFunc("/status", httpStatus)

	http.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bye!")
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		log.Fatal("/quit handler")
	})

	http.HandleFunc("/sleep", httpSleep)
	http.HandleFunc("/distance", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Version: ")
		if hw, sw, e := meter.GetVersion(); e != nil {
			fmt.Fprintln(w, e)
		} else {
			fmt.Fprintf(w, "LIDAR-Lite v2 \"Blue Label\" hw%dsw%d\n", hw, sw)
		}

		/* Continuous mode. First, prepare safe position for hand. Rotate lidar during operation
		_ = meter.SetContinuousMode(50, 100*time.Millisecond)
		_ = meter.Acquire(false)
		for i := 0; i < 50; i++ {
			fmt.Println("Round", i)
			fmt.Println(s.ReadDistance())
			fmt.Println(s.ReadVelocity())
			time.Sleep(100 * time.Millisecond)
		}*/
	})

	http.HandleFunc("/wake", httpWake)

	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		board.Reset(bb.ResetPin)
		fmt.Fprintln(w, "Done")
	})

	http.HandleFunc("/move", httpMove)

	log.Fatal(http.ListenAndServe(":8080", nil))

	/*_ = mco.Left(mc.MaxSpeed / 2)
	time.Sleep(300 * time.Millisecond)
	_ = mco.Left(0)
	time.Sleep(time.Second)
	_ = mco.Right(-mc.MaxSpeed / 2)
	time.Sleep(300 * time.Millisecond)
	_ = mco.Right(0)
	time.Sleep(time.Second)

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

	time.Sleep(time.Second)
	park(board)
	time.Sleep(time.Second)*/
}
