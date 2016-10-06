package network

import (
	"errors"
	"log"
	"net"

	"github.com/huin/goupnp/dcps/internetgateway1"
	"github.com/huin/goupnp/scpd"
)

func getLocalIPAddressForGateway() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close() }()
	var addr string
	addr, _, err = net.SplitHostPort(conn.LocalAddr().String())
	return addr, err
}

// SetupForwarding contacts a local gateway via UPnP and instructs it
// to forward externalPort to localPort on the internal local IP address discovered.
func SetupForwarding(localPort, externalPort uint16) (string, error) {
	clients, _, err := internetgateway1.NewWANIPConnection1Clients()
	var localIP string
	localIP, err = getLocalIPAddressForGateway()
	for _, c := range clients {
		var disco *scpd.SCPD
		disco, err = c.ServiceClient.Service.RequestSCDP()
		if err == nil {
			if disco.GetAction("AddPortMapping") != nil {
				if disco.GetAction("DeletePortMapping") != nil {
					if err = c.DeletePortMapping("", externalPort, "TCP"); err != nil {
						log.Printf("Deleting port forwarding mapping for %d: %s\n",
							externalPort, err)
					}
				}
				err = c.AddPortMapping("", externalPort, "TCP", localPort, localIP, true,
					"Rover", 0)
				if err == nil {
					if disco.GetAction("GetExternalIPAddress") != nil {
						ip, _ := c.GetExternalIPAddress()
						return ip, nil
					}
					return "", nil
				}
			}
		}
	}
	if err == nil {
		err = errors.New("No WAN IP gateways providing port mapping functions discovered")
	}
	return "", err
}
