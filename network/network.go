package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/kasheemlew/xperiMoby/container"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

var (
	defaultNetworkPath = "/var/run/xperiMoby/network/network/"
	drivers            = map[string]NetworkDriver{}
	networks           = map[string]*Network{}
)

// Endpoint will be used to connect container and network
type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	PortMapping []string         `json:"portmapping"`
	Network     *Network
}

// Network will be connected to container with Bridge
type Network struct {
	Name    string
	IPRange *net.IPNet
	Driver  string
}

// NetworkDriver is a interface of Network Device, deciding which driver to use in the network
type NetworkDriver interface {
	Name() string
	Create(subnet, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network *Network, endpoint *Endpoint) error
}

// CreateNetwork create a new network instance
func CreateNetwork(driver, subnet, name string) error {
	// convert network segment string to object
	_, cidr, _ := net.ParseCIDR(subnet)
	gatewayIP, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = gatewayIP
	logrus.Infof("get cidr: %v", cidr)

	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return nil
	}
	return nw.dump(defaultNetworkPath)
}

func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.Remove(path.Join(dumpPath, nw.Name))
}

func (nw *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}
	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Errorf("OpenFile error: %v", err)
		return err
	}
	defer nwFile.Close()

	nwJSON, err := json.Marshal(nw)
	if err != nil {
		logrus.Errorf("Marshal error: %v", err)
		return err
	}

	if _, err := nwFile.Write(nwJSON); err != nil {
		logrus.Errorf("WriteFile error: %v", err)
		return err
	}
	return nil
}

func (nw *Network) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	defer nwConfigFile.Close()
	if err != nil {
		logrus.Errorf("open nwConfigFile error: %v", err)
		return err
	}
	nwJSON := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJSON)
	if err != nil {
		logrus.Errorf("read nwConfigFile error: %v", err)
		return err
	}
	err = json.Unmarshal(nwJSON[:n], nw)
	if err != nil {
		logrus.Errorf("Error load nw info %v", err)
		return err
	}
	return nil
}

// Connect connect container to network
func Connect(networkName string, cinfo *container.ContainerInfo) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("No Such Network: %s", networkName)
	}
	// distribute IP address for container from network ip segment
	ip, err := ipAllocator.Allocate(network.IPRange)
	if err != nil {
		return err
	}
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.ID, network.Name),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	if err = configEndpointIPAddressAndRoute(ep, cinfo); err != nil {
		return err
	}
	// configure the port mapping between host and container
	return configPortMapping(ep, cinfo)
}

// Init inits network configs
func Init() error {
	// load network driver
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	// check network path and create if needed
	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return err
		}
	}
	// load network configures to map `networks`
	filepath.Walk(
		defaultNetworkPath,
		func(nwPath string, info os.FileInfo, err error) error {
			// pass directories
			if info.IsDir() {
				return nil
			}
			_, nwName := path.Split(nwPath)
			nw := &Network{
				Name: nwName,
			}
			if err := nw.load(nwPath); err != nil {
				logrus.Errorf("load network error: %v", err)
				return err
			}
			networks[nwName] = nw
			return nil
		},
	)
	return nil
}

// ListNetwork list
func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIPRange\tDriver\n")
	for _, nw := range networks {
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\n",
			nw.Name,
			nw.IPRange.String(),
			nw.Driver,
		)
	}
	if err := w.Flush(); err != nil {
		logrus.Errorf("Flush error %v", err)
		return
	}
}

// DeleteNetwork deletes network and drivers
func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("No Such Network: %s", networkName)
	}

	if err := ipAllocator.Release(nw.IPRange, &nw.IPRange.IP); err != nil {
		return fmt.Errorf("Error Remove Network gateway ip: %s", err)
	}

	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("Error Remove Network DriverError: %s", err)
	}

	return nw.remove(defaultNetworkPath)
}

func enterContainerNetns(enLink *netlink.Link, cinfo *container.ContainerInfo) func() {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		logrus.Errorf("error get container net namespace, %v", err)
	}
	nsFD := f.Fd()
	runtime.LockOSThread()

	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		logrus.Errorf("error set link netns , %v", err)
	}

	origns, err := netns.Get()
	if err != nil {
		logrus.Errorf("error get current netns, %v", err)
	}

	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		logrus.Errorf("error set netns, %v", err)
	}
	return func() {
		netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

func configEndpointIPAddressAndRoute(ep *Endpoint, cinfo *container.ContainerInfo) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}

	defer enterContainerNetns(&peerLink, cinfo)()

	interfaceIP := *ep.Network.IPRange
	interfaceIP.IP = ep.IPAddress

	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}

	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}

	if err = setInterfaceUP("lo"); err != nil {
		return err
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IPRange.IP,
		Dst:       cidr,
	}

	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}

	return nil
}

func configPortMapping(ep *Endpoint, cinfo *container.ContainerInfo) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			logrus.Errorf("port mapping format error, %v", pm)
			continue
		}
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			logrus.Errorf("iptables Output, %v", output)
			continue
		}
	}
	return nil
}
