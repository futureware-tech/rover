package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dasfoo/i2c"
	"github.com/dasfoo/rover/auth"
	"github.com/dasfoo/rover/bb"
	"github.com/dasfoo/rover/camera"
	"github.com/dasfoo/rover/mc"
	"github.com/dasfoo/rover/network"
	"github.com/dasfoo/rover/rpc"
	"golang.org/x/net/context"

	dns "google.golang.org/api/dns/v1"
)

var (
	board  *bb.BB
	motors *mc.MC

	testMode = flag.Bool("test", false,
		"Testing mode (running application from dev environment)")
	listenAddress = flag.String("listen", "",
		"Listen address: [<ip>]:<port>")
	gcsBucket = flag.String("gcs_bucket", "",
		"Name of GCS bucket containing authorization data")
	domainsString = flag.String("domains", "rover.dasfoo.org,fb.rover.dasfoo.org",
		"List of domains for DNS updates, first domain will get DNS updates, "+
			"but TLS certificate will be obtained for all of them")
	cloudDNSZone = flag.String("cloud_dns_zone", "",
		"Google Cloud DNS Zone name, defaults to the first domain, dots replaced with dashes")

	domains []string
	am      *auth.Manager
)

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

// https://github.com/grpc/grpc-go/issues/106#issuecomment-246978683
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

func startServer() error {
	httpSrv := &http.Server{
		Addr: *listenAddress,
		Handler: routingHandler(
			(&rpc.Server{
				AM:     am,
				Motors: motors,
				Board:  board,
			}).CreateGRPCServer(),
			http.HandlerFunc((&camera.Server{
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
	if domains[0] == "" {
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
		log.Fatal("Can't initialize auth manager:", ame)
	}

	if err := startForwarding(); err != nil {
		log.Println("Failed to setup forwarding:", err)
	}
	if err := startServer(); err != nil {
		log.Println("Failed to start server:", err)
	}
}
