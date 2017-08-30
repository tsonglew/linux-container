package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// BridgeNetworkDriver is network driver of bridge
type BridgeNetworkDriver struct {
}

// Name returns bridge
func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}

// Create creates a BridgeNetworkDriver
func (d *BridgeNetworkDriver) Create(subnet, name string) (*Network, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	n := &Network{
		Name:    name,
		IPRange: ipRange,
		Driver:  d.Name(),
	}
	// configure Linux Bridge
	if err := d.initBridge(n); err != nil {
		logrus.Errorf("error init bridge: %v", err)
	}
	return n, nil
}

func (d *BridgeNetworkDriver) initBridge(n *Network) error {
	// create virtual device
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("Error add bridge: %s, Error: %v", bridgeName, err)
	}
	// configure address and router of Bridge device
	gatewayIP := *n.IPRange
	gatewayIP.IP = n.IPRange.IP
	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return fmt.Errorf("Error assigning address: %s on bridge: %s with an error of: %v", gatewayIP, bridgeName, err)
	}
	// start Bridge device
	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("Error set bridge up: %s, Error: %v", bridgeName, err)
	}
	// set SNAT of iptables
	if err := setupIPTables(bridgeName, n.IPRange); err != nil {
		return fmt.Errorf("Error setting iptables for %s: %v", bridgeName, err)
	}
	return nil
}

// createBridgeInterface creates Bridge virtual device
func createBridgeInterface(bridgeName string) error {
	if _, err := net.InterfaceByName(bridgeName); err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}
	// a brand new Link object
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	br := &netlink.Bridge{la, nil, nil}
	// create virtual device
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("Bridge creation failed for bridge %s: %v", bridgeName, err)
	}
	return nil
}

// setInterfaceIP configures address and router of Bridge device
func setInterfaceIP(name, rawIP string) error {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("error get interface: %v", err)
	}
	// network segment: 192.168.0.0/24 & raw ip: 192.168.0.1
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{ipNet, "", 0, 0, nil, nil, 0, 0}
	return netlink.AddrAdd(iface, addr)
}

// setInterfaceUP sets network interfaces up
func setInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("Error retrieving a link named %s: %v", interfaceName, err)
	}
	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("Error enabling interface for %s: %v", interfaceName, err)
	}
	return nil
}

// setIPTables MASQUERADE of iptables's bridges
func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	if _, err := exec.Command("iptables", strings.Split(iptablesCmd, " ")...).CombinedOutput(); err != nil {
		logrus.Errorf("iptablese Ouput, %v", err)
		return err
	}
	return nil
}

// Delete deletes all network devices
func (d *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	return netlink.LinkDel(br)
}

// Connect connects container to the former built network
func (d *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	la.MasterIndex = br.Attrs().Index

	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}

	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("Error Add Endpoint Device: %v", err)
	}

	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("Error Add Endpoint Device: %v", err)
	}
	return nil
}

// Disconnect disconnects container from network
func (d *BridgeNetworkDriver) Disconnect(network *Network, endpoint *Endpoint) error {
	return nil
}
