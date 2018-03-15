package virtualmachine

import (
	"github.com/influxdata/telegraf/dcai/topology/datastore"
	"github.com/influxdata/telegraf/dcai/topology/snapshot"
)

type VirtualmachineInfo struct {
	Name       string
	Uuid       string
	Datastores []*datastore.DatastoreInfo
	Snapshots  []*snapshot.SnapshotInfo
}

func (vm *VirtualmachineInfo) DomainID() string {
	return vm.Uuid
}

func NewVirtualmachineInfo(name string, uuid string, dss []*datastore.DatastoreInfo, snps []*snapshot.SnapshotInfo) (*VirtualmachineInfo, error) {
	vm := new(VirtualmachineInfo)
	vm.Name = name
	vm.Uuid = uuid

	if dss != nil {
		vm.Datastores = dss
	} else {
		vm.Datastores = make([]*datastore.DatastoreInfo, 0)
	}
	if snps != nil {
		vm.Snapshots = snps
	} else {
		vm.Snapshots = make([]*snapshot.SnapshotInfo, 0)
	}
	return vm, nil
}
