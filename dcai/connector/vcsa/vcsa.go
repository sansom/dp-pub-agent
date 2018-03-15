package vcsa

import (
	"context"
	"fmt"
	"github.com/influxdata/telegraf/dcai/hardware/disk"
	"github.com/influxdata/telegraf/dcai/hardware/nic"
	"github.com/influxdata/telegraf/dcai/topology/cluster"
	"github.com/influxdata/telegraf/dcai/topology/datacenter"
	"github.com/influxdata/telegraf/dcai/topology/datastore"
	"github.com/influxdata/telegraf/dcai/topology/host"
	"github.com/influxdata/telegraf/dcai/topology/host/vmware/esxi"
	"github.com/influxdata/telegraf/dcai/topology/snapshot"
	"github.com/influxdata/telegraf/dcai/topology/virtualmachine"
	"github.com/influxdata/telegraf/dcai/topology/vsandiskgroup"
	"github.com/influxdata/telegraf/dcai/type"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/units"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"net/url"
	"regexp"
)

var (
	wwnRegexp  = regexp.MustCompile("^naa\\.(.*)")
	vmfsRegexp = regexp.MustCompile(".*/(.*)/$")
)

type VcsaConnector struct {
	Name          string
	Url           string
	Username      string
	Password      string
	AllowInsecure bool
	Client        *govmomi.Client
	ClientCtx     *context.Context
}

func NewVcsaConnector(name string, url string, user string, pw string, insecure bool) (*VcsaConnector, error) {
	vcsa := new(VcsaConnector)
	vcsa.Name = name
	vcsa.Url = url
	vcsa.Username = user
	vcsa.Password = pw
	vcsa.AllowInsecure = insecure

	return vcsa, nil
}

func (vcsa *VcsaConnector) ConnectVsphere() error {
	var u *url.URL
	var err error
	u, err = url.Parse(fmt.Sprintf("https://%s/sdk", vcsa.Url))
	if err != nil {
		return err
	}
	u.User = url.UserPassword(vcsa.Username, vcsa.Password)
	ctx := context.Background()
	vcsa.ClientCtx = &ctx

	// Connect and login to ESX or vCenter
	vcsa.Client, err = govmomi.NewClient(*vcsa.ClientCtx, u, vcsa.AllowInsecure)
	if err != nil {
		vcsa.ClientCtx = nil
	}

	return err
}

func (vcsa *VcsaConnector) DisconnectVsphere() error {
	if vcsa.Client == nil || vcsa.ClientCtx == nil {
		return fmt.Errorf("vcsa client is not connected.")
	}
	vcsa.Client.Logout(*vcsa.ClientCtx)
	return nil
}

func (vcsa *VcsaConnector) Retrieve(mor *types.ManagedObjectReference, ps []string, recursive bool, dst interface{}) error {
	m := view.NewManager(vcsa.Client.Client)
	v, err := m.CreateContainerView(*vcsa.ClientCtx, *mor, ps, recursive)
	if err != nil {
		return err
	}
	defer v.Destroy(*vcsa.ClientCtx)
	err = v.Retrieve(*vcsa.ClientCtx, ps, nil, dst)
	if err != nil {
		return err
	}
	return nil
}

func (vcsa *VcsaConnector) GetTopology() ([]*datacenter.DatacenterConfig, error) {
	var (
		dcmos []mo.Datacenter
		hss   []host.HostConfig
		cls   []*cluster.ClusterConfig
		err   error
		dcs   []*datacenter.DatacenterConfig
	)

	err = vcsa.Retrieve(&vcsa.Client.ServiceContent.RootFolder, []string{"Datacenter"}, true, &dcmos)
	if err != nil {
		return nil, err
	}

	for _, dcmo := range dcmos {
		// get hosts for this datacenters
		hss, err = vcsa.getHosts(&dcmo, &cls)
		if err != nil {
			return nil, err
		}

		d, err := datacenter.NewDatacenterConfig(dcmo.Entity().Name, cls, hss)
		if err != nil {
			return nil, err
		}

		dcs = append(dcs, d)
	}

	return dcs, nil
}

func findCluster(uuid string, cls *[]*cluster.ClusterConfig) *cluster.ClusterConfig {
	for _, cl := range *cls {
		if uuid == cl.DomainID() {
			return cl
		}
	}
	return nil
}

func findDatastore(dsmo *mo.Datastore, refdss *[]*datastore.DatastoreInfo) *datastore.DatastoreInfo {
	for _, refds := range *refdss {
		info := dsmo.Info.GetDatastoreInfo()
		if str := vmfsRegexp.FindStringSubmatch(info.Url); len(str) > 1 {
			if refds.Uuid == str[1] {
				return refds
			}
		}
	}
	return nil
}

func findDiskByName(name string, refdisks *[]*disk.DiskInfo) *disk.DiskInfo {
	for _, refd := range *refdisks {
		if name == refd.Name {
			return refd
		}
	}
	return nil
}

func (vcsa *VcsaConnector) getHosts(dsmo *mo.Datacenter, cls *[]*cluster.ClusterConfig) ([]host.HostConfig, error) {
	var (
		err   error
		hsmos []mo.HostSystem
		disks []*disk.DiskInfo
		h     *esxi.EsxiHostConfig
		hosts []host.HostConfig
		dgs   []*vsandiskgroup.VsanDiskgroupInfo
		vms   []*virtualmachine.VirtualmachineInfo
		dss   []*datastore.DatastoreInfo
		cl    *cluster.ClusterConfig
	)
	err = vcsa.Retrieve(&dsmo.HostFolder, []string{"HostSystem"}, true, &hsmos)
	if err != nil {
		return nil, err
	}
	for _, hsmo := range hsmos {

		var vnics = make([]*nic.NetworkInfo, 0)
		for _, vn := range hsmo.Config.Network.Vnic {
			var ipv6s []string
			if vn.Spec.Ip.IpV6Config != nil {
				for _, ipv6 := range vn.Spec.Ip.IpV6Config.IpV6Address {
					ipv6s = append(ipv6s, ipv6.IpAddress)
				}
			}
			vnic, _ := nic.NewNetworkInfo(vn.Device, []string{vn.Spec.Mac}, []string{vn.Spec.Ip.IpAddress}, ipv6s)
			vnics = append(vnics, vnic)
		}

		// get disks for this host
		disks, err = vcsa.getDisks(&hsmo)
		if err != nil {
			return nil, err
		}

		// get datastores
		dss, err = vcsa.getDatastores(&hsmo, &disks)
		if err != nil {
			return nil, err
		}

		// get vsan disk group provided by this host
		cl, dgs, err = vcsa.getVSANDiskgroups(&hsmo, &disks, cls)
		if err != nil {
			return nil, err
		}

		// get virtualmachines
		vms, err = vcsa.getVirtualMachines(&hsmo, &dss)
		if err != nil {
			return nil, err
		}

		h, err = esxi.NewEsxiHostConfig(
			hsmo.Summary.Config.Name,
			hsmo.Hardware.SystemInfo.Uuid,
			hsmo.Config.Product.Name,
			hsmo.Config.Product.Version,
			vnics,
			disks,
			vms,
			dss,
			cl,
			dgs,
		)
		if err != nil {
			return nil, err
		}

		if cl != nil {
			err = cl.AppendHost(h)
			if err != nil {
				return nil, err
			}
		}
		hosts = append(hosts, h)
	}

	return hosts, nil
}

func getDiskWwn(canonicalName string, uuid string) string {
	if str := wwnRegexp.FindStringSubmatch(canonicalName); len(str) > 1 {
		return str[1]
	} else {
		return uuid
	}
}

func (vcsa *VcsaConnector) getDisks(hsmo *mo.HostSystem) ([]*disk.DiskInfo, error) {
	var disks []*disk.DiskInfo

	for _, e := range hsmo.Config.StorageDevice.ScsiLun {
		if sd, ok := e.(*types.HostScsiDisk); ok {
			var capacity int64
			var lun = e.GetScsiLun()
			var d = disk.DiskInfo{}

			capacity = int64(sd.Capacity.Block) * int64(sd.Capacity.BlockSize)
			if sd.Ssd != nil && *sd.Ssd {
				d.Type = dcaitype.DiskTypeSSD
			} else {
				d.Type = dcaitype.DiskTypeHDD
			}
			d.Size = units.ByteSize(capacity).String()
			d.SectorSize = fmt.Sprint(sd.Capacity.BlockSize)
			d.Name = lun.CanonicalName
			d.WWN = getDiskWwn(lun.CanonicalName, lun.Uuid)
			d.Vendor = lun.Vendor
			d.Model = lun.Model
			d.FirmwareVersion = lun.Revision
			d.SerialNumber = lun.SerialNumber
			d.Status = dcaitype.DiskStatusUnknown

			disks = append(disks, &d)
		}
	}

	return disks, nil
}

func (vcsa *VcsaConnector) getVSANDiskgroups(hsmo *mo.HostSystem, refdisks *[]*disk.DiskInfo, cls *[]*cluster.ClusterConfig) (*cluster.ClusterConfig, []*vsandiskgroup.VsanDiskgroupInfo, error) {
	var (
		cl  *cluster.ClusterConfig
		err error
		dgs []*vsandiskgroup.VsanDiskgroupInfo
	)

	// looking for vsan cluster
	if hsmo.Config.VsanHostConfig.ClusterInfo.Uuid != "" {
		cl = findCluster(hsmo.Config.VsanHostConfig.ClusterInfo.Uuid, cls)
		if cl == nil {
			cl, err = cluster.NewClusterConfig(dcaitype.ClustervSAN, hsmo.Config.VsanHostConfig.ClusterInfo.Uuid, hsmo.Config.VsanHostConfig.ClusterInfo.Uuid)
			if err != nil {
				return nil, nil, err
			}
			*cls = append(*cls, cl)
		}
	}

	// looking for cache disk and capacity disks
	for _, dm := range hsmo.Config.VsanHostConfig.StorageInfo.DiskMapInfo {
		dg, err := vsandiskgroup.NewVsanDiskgroupInfo(hsmo.Summary.Config.Name, hsmo.Hardware.SystemInfo.Uuid, nil, nil)
		if err != nil {
			return nil, nil, err
		}
		d := dm.Mapping
		founddisk := findDiskByName(d.Ssd.CanonicalName, refdisks)
		if founddisk == nil {
			return nil, nil, fmt.Errorf("Cannot find disk %v", d.Ssd.CanonicalName)
		}
		dg.AppendCacheDisk(founddisk)
		for _, nssd := range d.NonSsd {
			founddisk := findDiskByName(nssd.CanonicalName, refdisks)
			if founddisk == nil {
				return nil, nil, fmt.Errorf("Cannot find disk %v", nssd.CanonicalName)
			}
			dg.AppendCapacityDisk(founddisk)
		}
		dgs = append(dgs, dg)
	}
	return cl, dgs, nil
}

func (vcsa *VcsaConnector) getVirtualMachines(hsmo *mo.HostSystem, dssref *[]*datastore.DatastoreInfo) ([]*virtualmachine.VirtualmachineInfo, error) {
	var (
		err   error
		vmmos []mo.VirtualMachine
		vms   []*virtualmachine.VirtualmachineInfo
		vm    *virtualmachine.VirtualmachineInfo
		snps  []*snapshot.SnapshotInfo
		dss   []*datastore.DatastoreInfo
	)

	property.DefaultCollector(vcsa.Client.Client).Retrieve(*vcsa.ClientCtx, hsmo.Vm, nil, &vmmos)

	for _, vmmo := range vmmos {
		// get datastore of this vm
		dss, _ = vcsa.getVMDatastores(&vmmo, dssref)

		// get snapshot of this vm
		snps, _ = vcsa.getVMSnapshots(&vmmo)

		vm, err = virtualmachine.NewVirtualmachineInfo(vmmo.Summary.Config.Name, vmmo.Config.Uuid, dss, snps)
		if err != nil {
			return nil, err
		}
		vms = append(vms, vm)
	}

	return vms, nil
}

func (vcsa *VcsaConnector) getVMFromSnapshotTree(root *types.VirtualMachineSnapshotTree, snps *[]*snapshot.SnapshotInfo) error {
	if root == nil {
		return nil
	}
	snp, err := snapshot.NewSnapshotInfo(root.Name, fmt.Sprintf("%d", root.Id))
	if err != nil {
		return err
	}
	*snps = append(*snps, snp)
	for _, child := range root.ChildSnapshotList {
		err := vcsa.getVMFromSnapshotTree(&child, snps)
		if err != nil {
			return err
		}
	}
	return nil
}

func (vcsa *VcsaConnector) getVMSnapshots(vmmo *mo.VirtualMachine) ([]*snapshot.SnapshotInfo, error) {
	var snps []*snapshot.SnapshotInfo

	if vmmo.Snapshot == nil {
		return nil, nil
		//fmt.Errorf("vm %s does not have snapshot", vmmo.Summary.Config.Name)
	}
	for _, snpr := range vmmo.Snapshot.RootSnapshotList {
		err := vcsa.getVMFromSnapshotTree(&snpr, &snps)
		if err != nil {
			return nil, err
		}
	}
	return snps, nil
}

func (vcsa *VcsaConnector) getVMDatastores(vmmo *mo.VirtualMachine, dssref *[]*datastore.DatastoreInfo) ([]*datastore.DatastoreInfo, error) {
	var (
		dsmos []mo.Datastore
		dss   []*datastore.DatastoreInfo
	)

	property.DefaultCollector(vcsa.Client.Client).Retrieve(*vcsa.ClientCtx, vmmo.Datastore, nil, &dsmos)

	for _, dsmo := range dsmos {
		ds := findDatastore(&dsmo, dssref)
		if ds == nil {
			return nil, fmt.Errorf("Cannot find datastore %v", dsmo)
		}
		dss = append(dss, ds)
	}

	return dss, nil
}

func (vcsa *VcsaConnector) getDatastores(hsmo *mo.HostSystem, refdisks *[]*disk.DiskInfo) ([]*datastore.DatastoreInfo, error) {
	var (
		dss   []*datastore.DatastoreInfo
		dsmos []mo.Datastore
	)
	property.DefaultCollector(vcsa.Client.Client).Retrieve(*vcsa.ClientCtx, hsmo.Datastore, nil, &dsmos)

	for _, dsmo := range dsmos {
		info := dsmo.Info.GetDatastoreInfo()
		uuid := ""
		if str := vmfsRegexp.FindStringSubmatch(info.Url); len(str) > 1 {
			uuid = str[1]
		}

		// get disk ref
		var disks []*disk.DiskInfo
		s, ok := dsmo.Info.(*types.VmfsDatastoreInfo)
		if ok {
			for _, backingDisk := range s.Vmfs.Extent {
				founddisk := findDiskByName(backingDisk.DiskName, refdisks)
				if founddisk == nil {
					return nil, fmt.Errorf("Cannot find disk %v", backingDisk.DiskName)
				}
				disks = append(disks, founddisk)
			}
		}

		ds, err := datastore.NewDatastoreInfo(info.Name, uuid, disks)
		if err != nil {
			return nil, err
		}
		dss = append(dss, ds)
	}

	return dss, nil
}
