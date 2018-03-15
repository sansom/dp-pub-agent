package cpu

import (
	"bytes"
	"github.com/influxdata/telegraf/dcai/util"
	"regexp"
)

var (
	execcmd             = util.ExecuteCmdWithTimeout
	cpuProcessorRegexp  = regexp.MustCompile("^\\s*processor\\s+:\\s*(\\d+).*$")
	cpuVendorIDRegexp   = regexp.MustCompile("^\\s*vendor_id\\s+:\\s*(\\w.*\\w)\\s*$")
	cpuModelNameRegexp  = regexp.MustCompile("^\\s*model name\\s+:\\s*(\\w.*\\w)\\s*$")
	cpuMHzRegexp        = regexp.MustCompile("^\\s*cpu MHz\\s+:\\s*(\\w.*\\w)\\s*$")
	cpuCacheSizeRegexp  = regexp.MustCompile("^\\s*cache size\\s+:\\s*(\\w.*\\w)\\s*$")
	cpuPhysicalIDRegexp = regexp.MustCompile("^\\s*physical id\\s+:\\s*(\\d+).*$")
	cpuCoreIDRegexp     = regexp.MustCompile("^\\s*core id\\s+:\\s*(\\d+).*$")
)

type CpuInfo struct {
	ProcessorID string
	VendorID    string
	ModelName   string
	CurrentMHz  string
	CacheSize   string
	PhysicalID  string
	CoreID      string
}

func NewAllCpuInfo() ([]*CpuInfo, error) {
	var (
		cpus         []*CpuInfo
		cpu          *CpuInfo
		out          []byte
		err          error
		cputxt       []byte
		cpustart     int
		nextcpustart int
	)

	if out, err = execcmd("cat", "/proc/cpuinfo"); err != nil {
		return nil, err
	}

	cpustart = bytes.Index(out, []byte("processor"))
	for nextcpustart != -1 {
		nextcpustart = bytes.Index(out[cpustart+1:], []byte("processor"))
		if nextcpustart != -1 {
			cputxt = out[cpustart : cpustart+1+nextcpustart]
		} else {
			cputxt = out[cpustart:]
		}
		cpustart = cpustart + 1 + nextcpustart
		cpu = new(CpuInfo)
		m := []*util.FindRegexpMatchAndSetType{
			{&cpu.ProcessorID, cpuProcessorRegexp},
			{&cpu.VendorID, cpuVendorIDRegexp},
			{&cpu.ModelName, cpuModelNameRegexp},
			{&cpu.CurrentMHz, cpuMHzRegexp},
			{&cpu.CacheSize, cpuCacheSizeRegexp},
			{&cpu.PhysicalID, cpuPhysicalIDRegexp},
			{&cpu.CoreID, cpuCoreIDRegexp},
		}
		util.FindRegexpMatchAndSet(string(cputxt), m)
		cpus = append(cpus, cpu)
	}
	return cpus, nil
}
