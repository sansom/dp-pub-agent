package datastore

import (
	"github.com/influxdata/telegraf/dcai/hardware/disk"
)

type DatastoreInfo struct {
	Name  string
	Uuid  string
	Disks []*disk.DiskInfo
}

func (ds *DatastoreInfo) DomainID() string {
	return ds.Uuid
}

func NewDatastoreInfo(name string, uuid string, disks []*disk.DiskInfo) (*DatastoreInfo, error) {
	ds := new(DatastoreInfo)
	ds.Name = name
	ds.Uuid = uuid

	if disks != nil {
		ds.Disks = disks
	} else {
		ds.Disks = make([]*disk.DiskInfo, 0)
	}
	return ds, nil
}
