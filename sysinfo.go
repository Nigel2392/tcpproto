package tcpproto

import (
	"encoding/json"
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

	return info
}

func (s *SysInfo) ToJSON() string {
	json, _ := json.Marshal(s)
	return string(json)
}

func TrimExtraSpaces(s string) string {
	var res string
	var lastChar rune
	for _, char := range s {
		if char != ' ' || lastChar != ' ' {
			res += string(char)
		}
		lastChar = char
	}
	res = strings.TrimPrefix(res, " ")
	res = strings.TrimSuffix(res, " ")
	return res
}
