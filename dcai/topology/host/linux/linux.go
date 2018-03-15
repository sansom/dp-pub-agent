package linux

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf/dcai/hardware/disk"
	"github.com/influxdata/telegraf/dcai/hardware/linux/cpu"
	"github.com/influxdata/telegraf/dcai/hardware/linux/dmidecode"
	"github.com/influxdata/telegraf/dcai/hardware/linux/mem"
	"github.com/influxdata/telegraf/dcai/hardware/nic"
	"github.com/influxdata/telegraf/dcai/type"
	"github.com/influxdata/telegraf/dcai/util"
)

var (
	execcmd         = util.ExecuteCmdWithTimeout
	osNameRegexp    = regexp.MustCompile("^NAME=\"(.*)\"$")
	osVersionRegexp = regexp.MustCompile("^VERSION=\"(.*)\"$")
	osReleaseRegexp = regexp.MustCompile("^([^\\d]*)(.+)$")
)

type LinuxHostConfig struct {
	Name      string
	OSType    dcaitype.OSType
	OSName    string
	OSVersion string
	CPUs      []*cpu.CpuInfo
	MEMs      []*mem.MemInfo
	NICs      []*nic.NetworkInfo
	DmiInfo   *dmidecode.HostDmiInfo
	disks     []*disk.DiskInfo

	//	BlkDevs				[]*BlockDevInfo
}

func NewLinuxHostConfig(dmidecodePath string) (*LinuxHostConfig, error) {
	var (
		err error
	)

	h := new(LinuxHostConfig)
	h.disks = nil

	if h.Name, err = os.Hostname(); err != nil {
		return nil, err
	}

	// getting OS info
	if err = h.FetchLinuxOSInfo(); err != nil {
		log.Printf("W! Cannot get OS release information. (%s)", err.Error())
	}

	// get dmidecode info
	if h.DmiInfo, err = dmidecode.NewHostDmiInfo(dmidecodePath); err != nil {
		log.Printf("W! Cannot get host dmidecode information. (%s)", err.Error())
	}

	// get network info
	if h.NICs, err = nic.NewAllNetworkInfo(); err != nil {
		log.Printf("W! Cannot get network information. (%s)", err.Error())
	}

	// get cpu info
	if h.CPUs, err = cpu.NewAllCpuInfo(); err != nil {
		log.Printf("W! Cannot get CPU information. (%s)", err.Error())
	}

	// get memory info
	if h.MEMs, err = mem.NewAllMemInfo(dmidecodePath); err != nil {
		log.Printf("W! Cannot get memory information. (%s)", err.Error())
	}

	return h, nil
}

func (h *LinuxHostConfig) FetchLinuxOSInfo() error {
	var (
		out []byte
		err error
	)

	out, err = execcmd("cat", "/etc/os-release")
	if err == nil {
		m := []*util.FindRegexpMatchAndSetType{
			{&h.OSName, osNameRegexp},
			{&h.OSVersion, osVersionRegexp},
		}
		util.FindRegexpMatchAndSet(string(out), m)
	} else {
		// centos 6.5 has no os-release file
		// try /etc/redhat-release
		out, err = execcmd("cat", "/etc/redhat-release")
		if err != nil {
			return err
		}
		f := osReleaseRegexp.FindStringSubmatch(strings.TrimSpace(string(out)))
		if len(f) < 3 {
			return fmt.Errorf("Cannot parse /etc/redhat-release")
		}

		h.OSName = f[1]
		h.OSVersion = f[2]
	}

	h.OSType = dcaitype.OSLinux

	return nil
}

func (h *LinuxHostConfig) GetOsType() dcaitype.OSType {
	return dcaitype.OSLinux
}

// generate host HWID
func (h *LinuxHostConfig) HWID() string {
	hashkey := h.DmiInfo.String()
	// add MAC address to hash key
	for _, v := range h.NICs {
		hashkey = hashkey + strings.Join(v.MACs, "")
	}
	return util.GenHash(hashkey)
}

func (h *LinuxHostConfig) Hostname() string {
	return h.Name
}

func (h *LinuxHostConfig) DomainID() string {
	return h.HWID()
}

func (h *LinuxHostConfig) IPv4s() string {
	ipv4s := []string{}
	for _, nic := range h.NICs {
		ipv4s = append(ipv4s, nic.IPv4s...)
	}
	return strings.Join(ipv4s, ",")
}

func (h *LinuxHostConfig) IPv6s() string {
	ipv6s := []string{}
	for _, nic := range h.NICs {
		ipv6s = append(ipv6s, nic.IPv6s...)
	}
	return strings.Join(ipv6s, ",")
}

func (h *LinuxHostConfig) GetDisks(smartctlPath string) ([]*disk.DiskInfo, error) {
	var err error
	if h.disks != nil {
		return h.disks, nil
	}

	if h.disks, err = disk.GetLocalDisks(smartctlPath); err != nil {
		return nil, err
	}
	return h.disks, nil
}

func (h *LinuxHostConfig) SetDisks(disks []*disk.DiskInfo) {
	h.disks = disks
}
