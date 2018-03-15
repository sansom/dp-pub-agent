package esxi

import (
	"fmt"
	"strings"

	"github.com/influxdata/telegraf/dcai/hardware/disk"
	"github.com/influxdata/telegraf/dcai/hardware/nic"
	"github.com/influxdata/telegraf/dcai/topology/cluster"
	"github.com/influxdata/telegraf/dcai/topology/datastore"
	"github.com/influxdata/telegraf/dcai/topology/virtualmachine"
	"github.com/influxdata/telegraf/dcai/topology/vsandiskgroup"
	"github.com/influxdata/telegraf/dcai/type"
)

type EsxiHostConfig struct {
	Name           string
	Uuid           string
	OSType         dcaitype.OSType
	OSName         string
	OSVersion      string
	VNICs          []*nic.NetworkInfo
	Disks          []*disk.DiskInfo
	VMs            []*virtualmachine.VirtualmachineInfo
	Datastores     []*datastore.DatastoreInfo
	Vsan           *cluster.ClusterConfig
	VsanDiskgroups []*vsandiskgroup.VsanDiskgroupInfo
}

func NewEsxiHostConfig(name string, uuid string, osname string, osversion string, vnics []*nic.NetworkInfo, disks []*disk.DiskInfo, vms []*virtualmachine.VirtualmachineInfo, datastores []*datastore.DatastoreInfo, vsan *cluster.ClusterConfig, vsandgs []*vsandiskgroup.VsanDiskgroupInfo) (*EsxiHostConfig, error) {

	h := new(EsxiHostConfig)
	h.Name = name
	if uuid == "" {
		return nil, fmt.Errorf("host uuid cannot be empty")
	}
	h.Uuid = uuid
	h.OSType = dcaitype.OSVMware
	h.OSName = osname
	h.OSVersion = osversion

	if vnics != nil {
		h.VNICs = vnics
	} else {
		h.VNICs = make([]*nic.NetworkInfo, 0)
	}

	if disks != nil {
		h.Disks = disks
	} else {
		h.Disks = make([]*disk.DiskInfo, 0)
	}
	if vms != nil {
		h.VMs = vms
	} else {
		h.VMs = make([]*virtualmachine.VirtualmachineInfo, 0)
	}

	if datastores != nil {
		h.Datastores = datastores
	} else {
		h.Datastores = make([]*datastore.DatastoreInfo, 0)
	}

	h.Vsan = vsan
	if vsandgs != nil {
		h.VsanDiskgroups = vsandgs
	} else {
		h.VsanDiskgroups = make([]*vsandiskgroup.VsanDiskgroupInfo, 0)
	}

	return h, nil
}

func (e *EsxiHostConfig) GetOsType() dcaitype.OSType {
	return dcaitype.OSVMware
}

func (e *EsxiHostConfig) Hostname() string {
	return e.Name
}

func (e *EsxiHostConfig) DomainID() string {
	return e.HWID()
}

func (e *EsxiHostConfig) HWID() string {
	return e.Uuid
}

func (e *EsxiHostConfig) IPv4s() string {
	ipv4s := []string{}
	for _, nic := range e.VNICs {
		ipv4s = append(ipv4s, nic.IPv4s...)
	}
	return strings.Join(ipv4s, ",")
}

func (e *EsxiHostConfig) IPv6s() string {
	ipv6s := []string{}
	for _, nic := range e.VNICs {
		ipv6s = append(ipv6s, nic.IPv6s...)
	}
	return strings.Join(ipv6s, ",")
}

func (e *EsxiHostConfig) GetDisks(smartctlPath string) ([]*disk.DiskInfo, error) {
	return e.Disks, nil
}
