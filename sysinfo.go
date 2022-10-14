package tcpproto

import (
	"encoding/json"
	"net"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

// SysInfo saves the basic system information
type SysInfo struct {
	Hostname string `json:"hostname"`
	Platform string `json:"platform"`
	CPU      string `json:"cpu"`
	RAM      uint64 `json:"ram"`
	Disk     uint64 `json:"disk"`
	MacAddr  string `json:"macaddr"`
}

func GetSysInfo() *SysInfo {
	hostStat, _ := host.Info()
	cpuStat, _ := cpu.Info()
	vmStat, _ := mem.VirtualMemory()
	diskStat, _ := disk.Usage("\\") // If you're in Unix change this "\\" for "/"

	info := new(SysInfo)

	info.Hostname = TrimExtraSpaces(hostStat.Hostname)
	info.Platform = TrimExtraSpaces(hostStat.Platform)
	info.CPU = TrimExtraSpaces(cpuStat[0].ModelName)
	info.RAM = vmStat.Total / 1024 / 1024
	info.Disk = diskStat.Total / 1024 / 1024
	info.MacAddr, _ = GetMACAddr()

	return info
}

func (s *SysInfo) ToJSON() string {
	json, _ := json.Marshal(s)
	return string(json)
}

func GetMACAddr() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		CONF.LOGGER.Error(err.Error())
		return "", err
	}
	var currentIP, currentNetworkHardwareName string
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		// = GET LOCAL IP ADDRESS
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				currentIP = ipnet.IP.String()
			}
		}
	}
	// get all the system's or local machine's network interfaces
	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {
		if addrs, err := interf.Addrs(); err == nil {
			for _, addr := range addrs {
				// only interested in the name with current IP address
				if strings.Contains(addr.String(), currentIP) {
					currentNetworkHardwareName = interf.Name
				}
			}
		}
	}
	netInterface, err := net.InterfaceByName(currentNetworkHardwareName)
	if err != nil {
		CONF.LOGGER.Error(err.Error())
		return "", err
	}
	macAddress := netInterface.HardwareAddr
	// verify if the MAC address can be parsed properly
	hwAddr, err := net.ParseMAC(macAddress.String())
	if err != nil {
		return "", err
	}
	return hwAddr.String(), nil
}
