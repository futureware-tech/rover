package main

import (
	"log"
	"net"
	"time"

	"github.com/dasfoo/i2c"
	"github.com/dasfoo/rover/mc"
	"golang.org/x/net/context"

	"google.golang.org/grpc"

	pb "github.com/dasfoo/rover/proto"
)

const (
	port = ":50051"
)

// server is used to implement roverserver.RoverServiceServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) MoveRover(ctx context.Context, in *pb.RoverWheelRequest) (*pb.RoverWheelResponse, error) {
	if bus, err := i2c.NewBus(1); err != nil {
		log.Fatal(err)
	} else {
		// Silence i2c bus log
		//bus.Log = func(string, ...interface{}) {}
		motors := mc.NewMC(bus, mc.Address)
		_ = motors.Left(30)
		_ = motors.Right(30)
		time.Sleep(1 * time.Second)

		_ = motors.Left(0)
		_ = motors.Right(0)
	}
	return &pb.RoverWheelResponse{Message: "Ok "}, nil
}

func main() {
	log.Println("Server started")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRoverServiceServer(s, &server{})
	s.Serve(lis)
}
