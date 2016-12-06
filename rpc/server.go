package rpc

import (
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/dasfoo/rover/auth"
	"github.com/dasfoo/rover/bb"
	"github.com/dasfoo/rover/mc"
	pb "github.com/dasfoo/rover/proto"
)

// Server is an implementation of roverserver.RoverServiceServer.
type Server struct {
	AM     *auth.Manager
	Motors *mc.MC
	Board  *bb.BB
}

// Error definitions
var (
	ErrMotorsSoftwareBlocked = errors.New("Motors controller is software blocked")
	ErrBoardSoftwareBlocked  = errors.New("Board controller is software blocked")
)

// getGRPCError translates hardware errors into GRPC errors.
func (s *Server) getGRPCError(err error) error {
	// TODO(dotdoom): support more error conditions
	return grpc.Errorf(codes.Unavailable, "%s", err.Error())
}

const (
	authUserKey  = "auth-user"
	authTokenKey = "auth-token"
)

func getFirstValue(md metadata.MD, name string) (string, error) {
	values := md[name]
	if len(values) != 1 {
		return "", fmt.Errorf("Metadata key not found: %s", name)
	}
	return values[0], nil
}

func getUserAndToken(ctx context.Context) (string, string, error) {
	md, success := metadata.FromContext(ctx)
	if !success {
		return "", "", errors.New("No metadata found in the request")
	}
	user, err := getFirstValue(md, authUserKey)
	var token string
	if err == nil {
		token, err = getFirstValue(md, authTokenKey)
	}
	return user, token, err
}

func (s *Server) checkAccess(ctx context.Context) error {
	user, token, err := getUserAndToken(ctx)
	if err != nil {
		return err
	}
	return s.AM.CheckAccess(user, token)

}

// MoveRover is an example of using motors
func (s *Server) MoveRover(ctx context.Context,
	in *pb.RoverWheelRequest) (*pb.RoverWheelResponse, error) {
	var err error
	if err = s.checkAccess(ctx); err != nil {
		log.Printf("UserValid: %s", err)
		return nil, err
	}
	if s.Motors == nil {
		return nil, ErrMotorsSoftwareBlocked
	}
	if err = s.Motors.Left(int8(in.Left)); err != nil {
		return nil, err
	}
	if err = s.Motors.Right(int8(in.Right)); err != nil {
		return nil, err
	}
	time.Sleep(1 * time.Second)

	if err = s.Motors.Left(0); err != nil {
		return nil, err
	}
	if err = s.Motors.Right(0); err != nil {
		return nil, err
	}
	return &pb.RoverWheelResponse{}, nil
}

// GetBatteryPercentage returns battery value as reported by the Board
func (s *Server) GetBatteryPercentage(ctx context.Context,
	in *pb.BatteryPercentageRequest) (*pb.BatteryPercentageResponse, error) {
	var err error
	if err = s.checkAccess(ctx); err != nil {
		log.Printf("UserValid: %s", err)
		return nil, err
	}
	if s.Board == nil {
		return nil, ErrBoardSoftwareBlocked
	}
	var batteryPercentage byte
	if batteryPercentage, err = s.Board.GetBatteryPercentage(); err != nil {
		return nil, s.getGRPCError(err)
	}
	return &pb.BatteryPercentageResponse{
		Battery: int32(batteryPercentage),
	}, nil
}

// GetAmbientLight uses ambient light sensor
func (s *Server) GetAmbientLight(ctx context.Context,
	in *pb.AmbientLightRequest) (*pb.AmbientLightResponse, error) {
	var err error
	if err = s.checkAccess(ctx); err != nil {
		log.Printf("UserValid: %s", err)
		return nil, err
	}
	if s.Board == nil {
		return nil, ErrBoardSoftwareBlocked
	}
	var light uint16
	if light, err = s.Board.GetAmbientLight(); err != nil {
		return nil, s.getGRPCError(err)
	}
	return &pb.AmbientLightResponse{
		Light: int32(light),
	}, nil
}

// GetTemperatureAndHumidity uses DHT humidity sensor
func (s *Server) GetTemperatureAndHumidity(ctx context.Context,
	in *pb.TemperatureAndHumidityRequest) (*pb.TemperatureAndHumidityResponse, error) {
	var err error
	if err = s.checkAccess(ctx); err != nil {
		log.Printf("UserValid: %s", err)
		return nil, err
	}
	if s.Board == nil {
		return nil, ErrBoardSoftwareBlocked
	}
	var t, h byte
	if t, h, err = s.Board.GetTemperatureAndHumidity(); err != nil {
		return nil, s.getGRPCError(err)
	}
	return &pb.TemperatureAndHumidityResponse{
		Temperature: int32(t),
		Humidity:    int32(h),
	}, nil
}

// ReadEncoders reads current absolute values from 4 encoders
func (s *Server) ReadEncoders(ctx context.Context,
	in *pb.ReadEncodersRequest) (*pb.ReadEncodersResponse, error) {
	var err error
	if err = s.checkAccess(ctx); err != nil {
		log.Printf("UserValid: %s", err)
		return nil, err
	}

	if s.Motors == nil {
		return nil, ErrMotorsSoftwareBlocked
	}

	var leftFront, leftBack, rightFront, rightBack int32
	if leftFront, err = s.Motors.ReadEncoder(mc.EncoderLeftFront); err != nil {
		return nil, s.getGRPCError(err)
	}
	if leftBack, err = s.Motors.ReadEncoder(mc.EncoderLeftBack); err != nil {
		return nil, s.getGRPCError(err)
	}
	if rightFront, err = s.Motors.ReadEncoder(mc.EncoderRightFront); err != nil {
		return nil, s.getGRPCError(err)
	}
	if rightBack, err = s.Motors.ReadEncoder(mc.EncoderRightBack); err != nil {
		return nil, s.getGRPCError(err)
	}
	return &pb.ReadEncodersResponse{
		LeftFront:  leftFront,
		LeftBack:   leftBack,
		RightFront: rightFront,
		RightBack:  rightBack,
	}, nil
}
