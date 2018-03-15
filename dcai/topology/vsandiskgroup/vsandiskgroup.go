package vsandiskgroup

import (
	"fmt"
	"github.com/influxdata/telegraf/dcai/hardware/disk"
)

type VsanDiskgroupInfo struct {
	Name          string
	Uuid          string
	CacheDisks    []*disk.DiskInfo
	CapacityDisks []*disk.DiskInfo
}

func (dg *VsanDiskgroupInfo) DomainID() string {
	return dg.Uuid
}

func NewVsanDiskgroupInfo(name string, uuid string, cdisks []*disk.DiskInfo, capdisks []*disk.DiskInfo) (*VsanDiskgroupInfo, error) {

	dg := new(VsanDiskgroupInfo)
	dg.Name = name
	dg.Uuid = uuid
	if cdisks != nil {
		dg.CacheDisks = cdisks
	} else {
		dg.CacheDisks = make([]*disk.DiskInfo, 0)
	}
	if capdisks != nil {
		dg.CapacityDisks = capdisks
	} else {
		dg.CapacityDisks = make([]*disk.DiskInfo, 0)
	}

	return dg, nil
}

func (dg *VsanDiskgroupInfo) AppendCacheDisk(d *disk.DiskInfo) error {
	if dg.CacheDisks == nil {
		return fmt.Errorf("CacheDisks is nil")
	}
	if d == nil {
		return fmt.Errorf("Appending a nil disk")
	}

	dg.CacheDisks = append(dg.CacheDisks, d)
	return nil
}

func (dg *VsanDiskgroupInfo) AppendCapacityDisk(d *disk.DiskInfo) error {
	if dg.CapacityDisks == nil {
		return fmt.Errorf("CapacityDisks is nil")
	}
	if d == nil {
		return fmt.Errorf("Appending a nil disk")
	}

	dg.CapacityDisks = append(dg.CapacityDisks, d)
	return nil
}
