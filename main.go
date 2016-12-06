package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dasfoo/i2c"
	"github.com/dasfoo/rover/auth"
	"github.com/dasfoo/rover/bb"
	"github.com/dasfoo/rover/camera"
	"github.com/dasfoo/rover/mc"
	"github.com/dasfoo/rover/network"
	"golang.org/x/net/context"

	dns "google.golang.org/api/dns/v1"
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

	testMode = flag.Bool("test", false,
		"Testing mode (running application from dev environment)")
	listenAddress = flag.String("listen", "",
		"Listen address: [<ip>]:<port>")
	gcsBucket = flag.String("gcs_bucket", "rover-auth",
		"Name of GCS bucket containing authorization data")
	domainsString = flag.String("domains", "rover.dasfoo.org,fb.rover.dasfoo.org",
		"List of domains for DNS updates, first domain will get DNS updates, "+
			"but TLS certificate will be obtained for all of them")
	cloudDNSZone = flag.String("cloud_dns_zone", "",
		"Google Cloud DNS Zone name, defaults to the first domain, dots replaced with dashes")

	domains []string
	am      *auth.Manager
)

func getError(e error) error {
	return grpc.Errorf(codes.Unavailable, "%s", e.Error())
}

// MoveRover implements
func (s *server) MoveRover(ctx context.Context,
	in *pb.RoverWheelRequest) (*pb.RoverWheelResponse, error) {
	if err := checkAccess(ctx); err != nil {
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
	if err := checkAccess(ctx); err != nil {
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
	if err := checkAccess(ctx); err != nil {
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
	if err := checkAccess(ctx); err != nil {
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

func updateDNS(ip string) error {
	c, e := network.NewDNSClient(context.Background(), *cloudDNSZone)
	if e != nil {
		return e
	}
	return c.UpdateDNS(context.Background(),
		&dns.ResourceRecordSet{
			Name:    domains[0] + ".",
			Type:    "A",
			Rrdatas: []string{ip},
			Ttl:     60,
		}, true)
}

func (s *server) ReadEncoders(ctx context.Context,
	in *pb.ReadEncodersRequest) (*pb.ReadEncodersResponse, error) {

	if err := checkAccess(ctx); err != nil {
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
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcHandler.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func startForwarding() error {
	_, port, err := net.SplitHostPort(*listenAddress)
	if err != nil {
		return err
	}
	var (
		externalIP string
		portInt    int
	)
	portInt, err = strconv.Atoi(port)
	if err == nil {
		externalIP, err = network.SetupForwarding(uint16(portInt), uint16(portInt))
		if err == nil {
			go func() {
				if err = updateDNS(externalIP); err != nil {
					log.Println(err)
				}
			}()
		}
	}
	return err
}

func getFirstValue(md metadata.MD, name string) (string, error) {
	values := md[name]
	if len(values) != 1 {
		return "", fmt.Errorf("Metadata key not found: %s", name)
	}
	return values[0], nil
}

const (
	authUserKey  = "auth-user"
	authTokenKey = "auth-token"
)

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

func checkAccess(ctx context.Context) error {
	user, token, err := getUserAndToken(ctx)
	if err != nil {
		return err
	}
	return am.CheckAccess(user, token)
}

func startServer() error {
	s := grpc.NewServer()
	pb.RegisterRoverServiceServer(s, &server{})
	httpSrv := &http.Server{
		Addr: *listenAddress,
		Handler: routingHandler(s, http.HandlerFunc((&camera.Server{
			ValidatePassword: func(password string) error {
				userAndToken := strings.Split(password, ":")
				if len(userAndToken) != 2 {
					return errors.New("Invalid password format")
				}
				return am.CheckAccess(userAndToken[0], userAndToken[1])
			},
		}).Handler)),
	}

	c, err := network.NewACMEClient(context.Background(), ".config/acme")
	// TODO(dotdoom): set c.DNS
	if err == nil {
		err = c.CheckOrRefreshCertificate(context.Background(), domains...)
	}
	if err == nil {
		certFile, keyFile := c.GetDomainsCertpairPath(domains...)
		log.Println("Starting HTTPS server")
		return httpSrv.ListenAndServeTLS(certFile, keyFile)
	}
	log.Println("Starting HTTP server, no TLS:", err)
	return httpSrv.ListenAndServe()
}

func main() {
	if os.Getenv("ROVER_LOG_TIMESTAMP") == "false" {
		log.SetFlags(log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	}

	flag.Parse()

	if *testMode {
		log.Println("*** THE APPLICATION IS RUNNING IN TESTING MODE ***")
	}

	domains = strings.Split(*domainsString, ",")
	if len(domains) < 1 {
		// TODO(dotdoom): ignore and proceed w/o DNS & HTTPS
		log.Fatal("Need at least one domain")
	}
	if *cloudDNSZone == "" {
		*cloudDNSZone = strings.Replace(domains[0], ".", "-", -1)
	}

	if bus, err := i2c.NewBus(1); err != nil {
		log.Fatal(err)
	} else {
		// Silence i2c bus log
		//bus.SetLogger(func(string, ...interface{}) {})

		board = bb.NewBB(bus, bb.Address)
		motors = mc.NewMC(bus, mc.Address)
	}

	var ame error
	am, ame = auth.NewManager(context.Background(), *gcsBucket)
	if ame != nil {
		// TODO(dotdoom): this is not really fatal. Keep going without auth if it fails,
		// but only if gcsBucket is not supplied.
		log.Fatal("Can't initialize auth manager:", ame)
	}

	if err := startForwarding(); err != nil {
		log.Println("Failed to setup forwarding:", err)
	}
	if err := startServer(); err != nil {
		log.Println("Failed to start server:", err)
	}
}
