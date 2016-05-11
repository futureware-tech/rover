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

// MoveRover implements
func (s *server) MoveRover(ctx context.Context, in *pb.RoverWheelRequest) (*pb.RoverWheelResponse, error) {
	_ = motors.Left(int8(in.Left)) //TODO error check
	_ = motors.Right(int8(in.Right))
	time.Sleep(1 * time.Second)

	_ = motors.Left(0)
	_ = motors.Right(0)
	return &pb.RoverWheelResponse{Message: "Ok "}, nil
}

func (s *server) GetBoardInfo(ctx context.Context, in *pb.BoardInfoRequest) (*pb.BoardInfoResponse, error) {
	var errorResponse = &pb.BoardInfoResponse{
		Battery:     0,
		Light:       0,
		Temperature: 0,
		Humidity:    0,
	}

	var batteryPercentage, t, h byte
	var e error
	var light uint16

	if batteryPercentage, e = board.GetBatteryPercentage(); e != nil {
		return errorResponse, e
	}

	if light, e = board.GetAmbientLight(); e != nil {
		return errorResponse, e
	}

	if t, h, e = board.GetTemperatureAndHumidity(); e != nil {
		return errorResponse, e
	}
	return &pb.BoardInfoResponse{
		Battery:     int32(batteryPercentage),
		Light:       int32(light),
		Temperature: int32(t),
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
