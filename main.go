package main

import (
	"flag"
	"log"
	"net"
	"time"

	"github.com/dasfoo/i2c"
	"github.com/dasfoo/rover/bb"
	"github.com/dasfoo/rover/mc"
	"golang.org/x/net/context"

	"google.golang.org/grpc"

	pb "github.com/dasfoo/rover/proto"
)

// server is used to implement roverserver.RoverServiceServer.
type server struct{}

var (
	board  *bb.BB
	motors *mc.MC
)

func getStatus(e error) *pb.Status {
	return &pb.Status{
		Code:    pb.StatusCode_HARDWARE_FAILURE,
		Message: e.Error(),
	}
}

// MoveRover implements
func (s *server) MoveRover(ctx context.Context,
	in *pb.RoverWheelRequest) (*pb.RoverWheelResponse, error) {
	_ = motors.Left(int8(in.Left)) //TODO error check
	_ = motors.Right(int8(in.Right))
	time.Sleep(1 * time.Second)

	_ = motors.Left(0)
	_ = motors.Right(0)
	return &pb.RoverWheelResponse{
		Status: &pb.Status{
			Code: pb.StatusCode_OK,
		},
	}, nil
}

func (s *server) GetBatteryPercentage(ctx context.Context,
	in *pb.BatteryPercentageRequest) (*pb.BatteryPercentageResponse, error) {
	var batteryPercentage byte
	var e error
	if batteryPercentage, e = board.GetBatteryPercentage(); e != nil {
		return &pb.BatteryPercentageResponse{
			Status: getStatus(e),
		}, e
	}
	return &pb.BatteryPercentageResponse{
		Status:  &pb.Status{},
		Battery: int32(batteryPercentage),
	}, e
}

func (s *server) GetAmbientLight(ctx context.Context,
	in *pb.AmbientLightRequest) (*pb.AmbientLightResponse, error) {
	var light uint16
	var e error
	if light, e = board.GetAmbientLight(); e != nil {
		return &pb.AmbientLightResponse{
			Status: getStatus(e),
		}, e
	}
	return &pb.AmbientLightResponse{
		Status: &pb.Status{},
		Light:  int32(light),
	}, e
}

func (s *server) GetTemperatureAndHumidity(ctx context.Context,
	in *pb.TemperatureAndHumidityRequest) (*pb.TemperatureAndHumidityResponse, error) {
	var t, h byte
	var e error
	if t, h, e = board.GetTemperatureAndHumidity(); e != nil {
		return &pb.TemperatureAndHumidityResponse{
			Status: getStatus(e),
		}, e

	}
	return &pb.TemperatureAndHumidityResponse{
		Status:      &pb.Status{},
		Temperature: int32(t), // TODO: check byte in proto
		Humidity:    int32(h),
	}, e
}

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	var laddr = flag.String("laddr", "", "laddr")
	var test = flag.Bool("test", false, "Flag for startup script")
	flag.Parse()
	log.Println("Properties from command line:", *laddr)
	log.Println("Flag for startup script", *test)
	if bus, err := i2c.NewBus(1); err != nil {
		log.Fatal(err)
	} else {
		// Silence i2c bus log
		//bus.Log = func(string, ...interface{}) {}

		board = bb.NewBB(bus, bb.Address)
		motors = mc.NewMC(bus, mc.Address)
	}
	lis, err := net.Listen("tcp", *laddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Server started")
	s := grpc.NewServer()
	pb.RegisterRoverServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
