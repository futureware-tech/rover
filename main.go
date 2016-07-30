package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dasfoo/i2c"
	"github.com/dasfoo/rover/bb"
	"github.com/dasfoo/rover/camera"
	"github.com/dasfoo/rover/mc"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/storage/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	pb "github.com/dasfoo/rover/proto"
)

// server is used to implement roverserver.RoverServiceServer.
type server struct{}

var (
	board  *bb.BB
	motors *mc.MC

	tls         = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	laddr       = flag.String("laddr", "", "laddr")
	test        = flag.Bool("test", false, "Flag for startup script")
	certFile    = flag.String("cert_file", "", "The TLS cert file")
	keyFile     = flag.String("key_file", "", "The TLS key file")
	cloudBucket = flag.String("cloud_bucket", "rover-auth", "GCS bucket containing auth data")
)

func getError(e error) error {
	errf := grpc.Errorf // Confuse `go vet' to not check this `Errorf' call. :(
	// See https://github.com/grpc/grpc-go/issues/90
	return errf(codes.Unavailable, "%s", e.Error())
}

// MoveRover implements
func (s *server) MoveRover(ctx context.Context,
	in *pb.RoverWheelRequest) (*pb.RoverWheelResponse, error) {
	if err := validateUser(ctx); err != nil {
		log.Printf("UserValid: %s", err)
		return nil, err
	}
	if e := motors.Left(int8(in.Left)); e != nil {
		return nil, e
	}
	if e := motors.Right(int8(in.Right)); e != nil {
		return nil, e
	}
	time.Sleep(1 * time.Second)

	if e := motors.Left(0); e != nil {
		return nil, e
	}
	if e := motors.Right(0); e != nil {
		return nil, e
	}
	return &pb.RoverWheelResponse{}, nil
}

func (s *server) GetBatteryPercentage(ctx context.Context,
	in *pb.BatteryPercentageRequest) (*pb.BatteryPercentageResponse, error) {
	if err := validateUser(ctx); err != nil {
		log.Printf("UserValid: %s", err)
		return nil, err
	}
	var batteryPercentage byte
	var e error
	if batteryPercentage, e = board.GetBatteryPercentage(); e != nil {
		return nil, getError(e)
	}
	return &pb.BatteryPercentageResponse{
		Battery: int32(batteryPercentage),
	}, nil
}

func (s *server) GetAmbientLight(ctx context.Context,
	in *pb.AmbientLightRequest) (*pb.AmbientLightResponse, error) {
	if err := validateUser(ctx); err != nil {
		log.Printf("UserValid: %s", err)
		return nil, err
	}
	var light uint16
	var e error
	if light, e = board.GetAmbientLight(); e != nil {
		return nil, getError(e)
	}
	return &pb.AmbientLightResponse{
		Light: int32(light),
	}, nil
}

func (s *server) GetTemperatureAndHumidity(ctx context.Context,
	in *pb.TemperatureAndHumidityRequest) (*pb.TemperatureAndHumidityResponse, error) {
	if err := validateUser(ctx); err != nil {
		log.Printf("UserValid: %s", err)
		return nil, err
	}
	var t, h byte
	var e error
	if t, h, e = board.GetTemperatureAndHumidity(); e != nil {
		return nil, getError(e)
	}
	return &pb.TemperatureAndHumidityResponse{
		Temperature: int32(t),
		Humidity:    int32(h),
	}, nil
}

func downloadPassword() (string, error) {
	var (
		client *http.Client
		svc    *storage.Service
		e      error
	)

	if client, e = google.DefaultClient(oauth2.NoContext,
		storage.DevstorageReadOnlyScope); e == nil {
		if svc, e = storage.New(client); e != nil {
			return "", e
		}
	} else {
		return "", e
	}

	var (
		r     *http.Response
		entry struct {
			SecretKey string `json:"secret_key"`
		}
	)
	if r, e = svc.Objects.Get(*cloudBucket, "password.json").Download(); e == nil {
		var b bytes.Buffer
		if _, e = b.ReadFrom(r.Body); e == nil {
			e = json.Unmarshal(b.Bytes(), &entry)
		}
	}
	return entry.SecretKey, e
}

func validatePassword(password string) error {
	pwd, err := downloadPassword()
	if err != nil {
		return err
	}
	if password != pwd {
		return errors.New("Wrong password for getting data")
	}
	return nil
}

func readAuthFromMetadata(ctx context.Context) (string, error) {
	const key = "authentication"
	// FromContext returns error as a bool
	val, err := metadata.FromContext(ctx)
	if !err {
		return "", errors.New("Error appears getting metadata")
	}
	return val[key][0], nil
}

func validateUser(ctx context.Context) error {
	pwd, err := readAuthFromMetadata(ctx)
	if err != nil {
		return err
	}
	if err := validatePassword(pwd); err != nil {
		return err
	}
	return nil
}

func (s *server) ReadEncoders(ctx context.Context,
	in *pb.ReadEncodersRequest) (*pb.ReadEncodersResponse, error) {

	if err := validateUser(ctx); err != nil {
		log.Printf("UserValid: %s", err)
		return nil, err
	}

	var leftFront, leftBack, rightFront, rightBack int32
	var e error
	if leftFront, e = motors.ReadEncoder(mc.EncoderLeftFront); e != nil {
		return nil, getError(e)
	}
	if leftBack, e = motors.ReadEncoder(mc.EncoderLeftBack); e != nil {
		return nil, getError(e)
	}
	if rightFront, e = motors.ReadEncoder(mc.EncoderRightFront); e != nil {
		return nil, getError(e)
	}
	if rightBack, e = motors.ReadEncoder(mc.EncoderRightBack); e != nil {
		return nil, getError(e)
	}
	return &pb.ReadEncodersResponse{
		LeftFront:  leftFront,
		LeftBack:   leftBack,
		RightFront: rightFront,
		RightBack:  rightBack,
	}, nil
}

func routingHandler(grpcHandler http.Handler, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(dotdoom): find & use a constant instead of hardcode for header name and value
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcHandler.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func startServer() error {
	s := grpc.NewServer()
	pb.RegisterRoverServiceServer(s, &server{})
	httpSrv := &http.Server{
		Addr: *laddr,
		Handler: routingHandler(s, http.HandlerFunc((&camera.Server{
			ValidatePassword: validatePassword,
		}).Handler)),
	}
	if *tls {
		return httpSrv.ListenAndServeTLS(*certFile, *keyFile)
	}
	return httpSrv.ListenAndServe()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	flag.Parse()
	log.Println("Properties from command line:", *laddr)
	log.Println("Flag for startup script", *test)
	if bus, err := i2c.NewBus(1); err != nil {
		log.Fatal(err)
	} else {
		// Silence i2c bus log
		//bus.SetLogger(func(string, ...interface{}) {})

		board = bb.NewBB(bus, bb.Address)
		motors = mc.NewMC(bus, mc.Address)
	}
	if err := startServer(); err != nil {
		log.Println("Failed to start server :", err)
	}
}
