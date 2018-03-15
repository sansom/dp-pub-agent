package tpgy

import (
	"testing"

	"github.com/influxdata/telegraf/dcai/hardware/disk"
	"github.com/influxdata/telegraf/dcai/hardware/linux/dmidecode"
	"github.com/influxdata/telegraf/dcai/hardware/nic"
	"github.com/influxdata/telegraf/dcai/topology/cluster"
	"github.com/influxdata/telegraf/dcai/topology/datacenter"
	"github.com/influxdata/telegraf/dcai/topology/host"
	"github.com/influxdata/telegraf/dcai/topology/host/linux"
	"github.com/influxdata/telegraf/dcai/type"
	"github.com/influxdata/telegraf/testutil"
)

var (
	linuxhost = linux.LinuxHostConfig{
		Name:      "prophetstor-Naomi",
		OSType:    dcaitype.OSLinux,
		OSName:    "Ubuntu",
		OSVersion: "16.04.3 LTS (Xenial Xerus)",
		CPUs:      nil,
		MEMs:      nil,
		NICs: []*nic.NetworkInfo{
			&nic.NetworkInfo{Name: "enp5s0f1", MACs: []string{"08:9e:01:c4:96:b4"}, IPv4s: []string{"172.31.86.223"}, IPv6s: []string{"fe80::cc6a:18a8:4cf9:a8a3"}},
			&nic.NetworkInfo{Name: "wlp4s0", MACs: []string{"0c:84:dc:5d:bb:91"}, IPv4s: []string{"172.31.86.129"}, IPv6s: []string{"fe80::5f90:ce1b:d120:6f0e"}},
		},
		DmiInfo: &dmidecode.HostDmiInfo{
			Baseboard: &dmidecode.DmiBaseboardInfo{Manufacturer: "Acer", ProductName: "Dazzle_HW", SerialNumber: "NBM9W11003328034907600"},
			System:    &dmidecode.DmiSystemInfo{Manufacturer: "Acer", ProductName: "Aspire V5-573G", SerialNumber: "NXMC5TA001328034907600"},
		},
	}

	Disks = []*disk.DiskInfo{
		&disk.DiskInfo{
			Name:              "/dev/sda",
			Status:            dcaitype.DiskStatusGood,
			Type:              dcaitype.DiskTypeHDDSATA,
			FirmwareVersion:   "0001SDM1",
			Model:             "ST500LT012-9WS142",
			SataVersion:       "SATA 2.6, 3.0 Gb/s (current: 3.0 Gb/s)",
			SectorSize:        "512 bytes logical, 4096 bytes physical",
			WWN:               "5000c50069fcbd0a",
			SerialNumber:      "W0VBX8CV",
			Size:              "500 GB",
			SmartHealthStatus: "PASSED",
			Vendor:            "Seagate Laptop Thin HDD",
			TransportProtocol: "",
		},
	}

	mockdatacenterconfig = &datacenter.DatacenterConfig{
		Name: "Global",
		Clusters: []*cluster.ClusterConfig{
			&cluster.ClusterConfig{
				Name:        "dpCluster",
				ClusterType: dcaitype.ClusterDefaultCluster,
				Hosts: []host.HostConfig{
					&linuxhost,
				},
			},
		},
	}
)

func TestGatherTpgy(t *testing.T) {

	s := &Tpgy{}

	var acc testutil.Accumulator

	linuxhost.SetDisks(Disks)

	s.gatherTpgy(&acc, "dpCluster", "DiskProphet for Lab Test", &linuxhost)

	var testsTpgyHost = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"host_uuid":         "b17d80c7c549034e96a7bb73952b6ae3",
				"name":              "prophetstor-Naomi",
				"cluster_domain_id": "dpCluster",
				"os_type":           "linux",
				"os_name":           "Ubuntu",
				"os_version":        "16.04.3 LTS (Xenial Xerus)",
				"host_ip":           "172.31.86.223,172.31.86.129",
				"host_ipv6":         "fe80::cc6a:18a8:4cf9:a8a3,fe80::5f90:ce1b:d120:6f0e",
			},
			map[string]string{
				"domain_id": linuxhost.HWID(),
			},
		},
	}
	for _, test := range testsTpgyHost {
		acc.AssertContainsTaggedFields(t, "sai_host", test.fields, test.tags)
	}
	var testsTpgyCluster = []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			map[string]interface{}{
				"name": "DiskProphet for Lab Test",
			},
			map[string]string{
				"domain_id": "dpCluster",
			},
		},
	}
	for _, test := range testsTpgyCluster {
		acc.AssertContainsTaggedFields(t, "sai_cluster", test.fields, test.tags)
	}
}
