package network

import (
	"encoding/json"
	"net"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

const ipamDefaultAllocatorPath = "/var/run/xperiMoby/network/ipam/subnet.json"

// IPAM is IP Address Manager
type IPAM struct {
	SubnetAllocatorPath string
	// key: network segment value: bitmap
	Subnets *map[string]string
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	logrus.Infof("path: %v", ipam.SubnetAllocatorPath)
	defer subnetConfigFile.Close()
	if err != nil {
		logrus.Errorf("open subnetConfigFile error %v", err)
		return err
	}
	subnetJSON := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJSON)
	if err != nil {
		logrus.Errorf("read subnetConfigFile error %v", err)
		return err
	}

	err = json.Unmarshal(subnetJSON[:n], ipam.Subnets)
	if err != nil {
		logrus.Errorf("Error load allocation info, %v", err)
		return err
	}
	return nil
}

func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(ipamConfigFileDir, 0644)
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer subnetConfigFile.Close()
	if err != nil {
		logrus.Errorf("open subnetConfigFile error %v", err)
		return err
	}
	ipamConfigJSON, err := json.Marshal(ipam.Subnets)
	if err != nil {
		logrus.Errorf("json marshal error %v", err)
		return err
	}
	if _, err := subnetConfigFile.Write(ipamConfigJSON); err != nil {
		logrus.Errorf("write to subnetConfigFile error %v", err)
		return err
	}
	return nil
}

// Allocate allocates an available IP
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	ipam.Subnets = &map[string]string{}
	if err = ipam.load(); err != nil {
		logrus.Errorf("Error load allocation info. %v", err)
	}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	// Size returns the number of leading ones and total bits in the mask.
	ones, size := subnet.Mask.Size()

	// init the nw segment if not existed
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		// use "0" to fill the network segment
		// 1<<uint8(size-ones), the same with 2^(size-ones), is the number of available bits of the mask
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-ones))
	}
	for c := range (*ipam.Subnets)[subnet.String()] {
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			ip = subnet.IP

			// add array no to network segment
			for t := uint(4); t > 0; t-- {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			ip[3]++
			break
		}
	}
	ipam.dump()
	return
}

// Release releases ip addresses with bitmap algorithm
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	logrus.Infof("IP: %v, Subnet: %v", ipaddr, subnet)
	ipam.Subnets = &map[string]string{}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	if err := ipam.load(); err != nil {
		logrus.Errorf("Error load allocation info, %v", err)
	}
	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3]--
	for t := uint(4); t > 0; t-- {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
		logrus.Info("one turn")
	}
	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	ipam.dump()
	return nil
}
