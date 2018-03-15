package vspheretpgy

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/dcai"
	"github.com/influxdata/telegraf/dcai/connector/vcsa"
	"github.com/influxdata/telegraf/dcai/event"
	saicluster "github.com/influxdata/telegraf/dcai/sai/cluster"
	"github.com/influxdata/telegraf/dcai/topology/cluster"
	"github.com/influxdata/telegraf/dcai/topology/datacenter"
	"github.com/influxdata/telegraf/dcai/topology/datastore"
	"github.com/influxdata/telegraf/dcai/topology/host"
	"github.com/influxdata/telegraf/dcai/topology/host/vmware/esxi"
	"github.com/influxdata/telegraf/dcai/topology/virtualmachine"
	"github.com/influxdata/telegraf/dcai/topology/vsandiskgroup"
	"github.com/influxdata/telegraf/dcai/type"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type nodeInfo struct {
	Label    string
	DomainID string
	Name     string
	VcsaURL  string
}

func genNode(label string, domainID string, name string, vcsaURL string) (*nodeInfo, error) {
	var node *nodeInfo
	if label == "" || domainID == "" || name == "" {
		return nil, fmt.Errorf("lack of node informations. (%s-%s-%s)", label, domainID, name)
	}

	node = &nodeInfo{
		Label:    label,
		DomainID: domainID,
		Name:     name,
		VcsaURL:  vcsaURL,
	}

	return node, nil
}

// Vspheretpgy struct parse used configuration metric
type Vspheretpgy struct {
	Urls [][]string
}

var sampleConfig = `
## The full HTTP URL for your vCenter.
##
## Multiple urls can be specified as different vCenter cluster,
## please follow the inputs order as below
## urls = [
##		["First_VCenter_name","First_VCenter_URL","First_VCenter_User","First_VCenter_PW"], 
##		["Second_VCenter_name","Second_VCenter_URL","Second_VCenter_User","Second_VCenter_PW"],
##		... 
##	]
## e.g.
## urls = [
##		["vc1","192.168.0.1","john","qwerad"], 
##		["vc2","192.168.0.2","peter","akdfljd"] 
##	]
##
urls = [[]]
`

// SampleConfig return sampleConfig
// would be generated with -sample-config true
func (n *Vspheretpgy) SampleConfig() string {
	return sampleConfig
}

// Description return "Read vCenter status information" with -sample-config ture
func (n *Vspheretpgy) Description() string {
	return "Read vCenter status information"
}

// Set acc precision to Nanosecond
// to avoid point cover another
func setPrecisionForVsphere(acc *telegraf.Accumulator) {
	(*acc).SetPrecision(time.Nanosecond, 0)
}

// Gather is main process while running program
func (n *Vspheretpgy) Gather(acc telegraf.Accumulator) error {
	// setPrecision function is the same as `acc.SetPrecision(time.Nanosecond, 0)`
	setPrecisionForVsphere(&acc)

	for i, urls := range n.Urls {
		if len(urls) == 0 {
			log.Printf("Need to put vCenter information!\n")

			continue
		}

		if len(urls) != 4 {
			acc.AddError(fmt.Errorf("the %d_th vsphere configuration is incorrect! ", i+1))

			continue
		}
		// for a give set of vcsas
		vc, err := vcsa.NewVcsaConnector(urls[0], urls[1], urls[2], urls[3], true)
		if err != nil {
			acc.AddError(fmt.Errorf("failed to connect '%v", err))

			continue
		}

		dcs, err := dcai.FetchVsphereTopology(vc)
		if err != nil {
			return err
		}
		timeStamp := fmt.Sprintf("%d", time.Now().UnixNano())

		err = startAccNeo4j(vc, dcs, timeStamp, acc)
		return err

	}

	return nil
}

func init() {
	inputs.Add("vspheretpgy", func() telegraf.Input { return &Vspheretpgy{} })
}

// function to build neo4j merge depends on nodeInfo
func mergeNeo4jNdoes(node *nodeInfo, timeStamp string) string {
	vcsaNodesCmd := ""
	cypherNodesCmd := ""

	if node.VcsaURL != "" {
		vcsaNodesCmd = ", vcsa:'" + node.VcsaURL + "'"
	}

	cypherNodesCmd = "merge (" + node.Label + ":" + node.Label + " {domainId:'" + node.DomainID + "', name:'" + node.Name + "'" + vcsaNodesCmd + "}) set " + node.Label + ".time='" + timeStamp + "' "

	return cypherNodesCmd
}

// function to build link between neo4j nodes
func mergeNeo4jLinks(node1 *nodeInfo, node2 *nodeInfo, relationship string, timeStamp string) string {
	cypherNodesCmd1 := mergeNeo4jNdoes(node1, timeStamp)
	cypherNodesCmd2 := mergeNeo4jNdoes(node2, timeStamp)
	cypherLinkCmd := "merge (" + node1.Label + ")-[" + node1.Label + node2.Label + ":" + relationship + "]->(" + node2.Label + ") set " + node1.Label + node2.Label + ".time='" + timeStamp + "' "
	return cypherNodesCmd1 + cypherNodesCmd2 + cypherLinkCmd
}

// parse each relationship graphs between neo4j nodes
func accNeo4jMap(node1 *nodeInfo, node2 *nodeInfo, relationship string, timeStamp string, acc telegraf.Accumulator) {
	cypherCmd := mergeNeo4jLinks(node1, node2, relationship, timeStamp)
	acc.AddFields("db_relay", map[string]interface{}{"cmd": cypherCmd}, map[string]string{"dc_tag": "na"}, time.Now())
}

// start from VMCluster generate nodes for different VCenter Urls
func startAccNeo4j(vcsa *vcsa.VcsaConnector, dcs []*datacenter.DatacenterConfig, timeStamp string, acc telegraf.Accumulator) error {
	// create dcaiAgent to get info
	dcaiAgent, err := dcai.GetDcaiAgent()
	if err != nil {
		return err
	}

	vcsaNode, err := genNode("VMClusterCenter", dcaiAgent.GetSaiClusterDomainId(), dcaiAgent.GetSaiClusterName(), vcsa.Url)

	if err != nil {
		return err
	}
	if dcs == nil {
		// if encount case that one vcsa doesn't contain any datacenter would update the time of vcsa without update datacenters' nodes information
		acc.AddFields("db_relay", map[string]interface{}{"cmd": mergeNeo4jNdoes(vcsaNode, timeStamp)}, map[string]string{"dc_tag": "na"}, time.Now())
		log.Printf("The vcsa %s doesn't contains any datacenters", vcsa.Url)

		return nil
	}

	err = vmClusterContainsVMHostAndVMDatastore(vcsaNode, dcaiAgent, dcs, timeStamp, acc)
	if err != nil {
		return err
	}

	return nil
}

// VmCluster Contains VmHost And VmDatastore
func vmClusterContainsVMHostAndVMDatastore(vcsaNode *nodeInfo, dcaiAgent *dcai.DcaiAgent, dcs []*datacenter.DatacenterConfig, timeStamp string, acc telegraf.Accumulator) error {
	// VMDatacenter
	for _, datacenter := range dcs {
		datacenterNode, err := genNode("VMDataCenter", datacenter.DomainID(), datacenter.Name, "")
		if err != nil {
			return err
		}
		accNeo4jMap(datacenterNode, vcsaNode, "VmDataCenterContainsVmCluster", timeStamp, acc)

		// VmDataCenter Contains VSancluster
		err = vmDataCenterContainsVSancluster(datacenterNode, datacenter, timeStamp, acc)
		if err != nil {
			return err
		}

		// VSanCluster Contains Hosts
		for _, vmhost := range datacenter.Hosts {
			vmhost := vmhost.(*esxi.EsxiHostConfig)

			host.CreateSaiHostDataPoint(acc, vmhost.DomainID(), vmhost.Hostname(), vmhost.HWID(), dcaiAgent.GetSaiClusterDomainId(), vmhost.OSType, vmhost.OSName, vmhost.OSVersion, vmhost.IPv4s(), vmhost.IPv6s())
			saicluster.CreateSaiClusterDataPoint(acc, dcaiAgent.GetSaiClusterDomainId(), dcaiAgent.GetSaiClusterName())
			event.SendMetricsMonitoring(acc, vmhost, dcaiAgent.GetSaiClusterDomainId(), fmt.Sprintf("1 point(s) of vmhost %s was written to DB", vmhost.Name), dcaitype.EventTitleHostDataSent, dcaitype.LogLevelInfo)

			vmhostNode, err := genNode("VMHost", vmhost.DomainID(), vmhost.Name, "")
			if err != nil {
				return err
			}
			accNeo4jMap(vmhostNode, vcsaNode, "VmClusterContainsVmHost", timeStamp, acc)

			// VmHostContains ;
			err = vmHostContainsVMDisk(vcsaNode, vmhostNode, vmhost, timeStamp, acc)
			if err != nil {
				return err
			}

			err = vmHostContainsVMDatastore(vcsaNode, vmhostNode, vmhost, timeStamp, acc)
			if err != nil {
				return err
			}

			err = vmHostHasVMDiskGroup(vcsaNode, vmhostNode, vmhost, timeStamp, acc)
			if err != nil {
				return err
			}

			err = vmHostHostsVMVirtualMachine(vcsaNode, vmhostNode, vmhost, timeStamp, acc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// VmDataCenter Contains VSancluster
func vmDataCenterContainsVSancluster(vmdatacenterNode *nodeInfo, datacenter *datacenter.DatacenterConfig, timeStamp string, acc telegraf.Accumulator) error {
	for _, vsancluster := range datacenter.Clusters {
		vsanclusterNode, err := genNode("VMVSanCluster", vsancluster.DomainID(), vsancluster.Name, "")
		if err != nil {
			return err
		}
		accNeo4jMap(vmdatacenterNode, vsanclusterNode, "VmDataCenterContainsVSanCluster", timeStamp, acc)

		// VSanCluster Contains VSanDiskGroup
		err = vSanClusterContainsVSanDiskGroup(vsanclusterNode, vsancluster, timeStamp, acc)
		if err != nil {
			return err
		}

	}
	return nil
}

// VSanCluster Contains VSanDiskGroup
func vSanClusterContainsVSanDiskGroup(vsanclusterNode *nodeInfo, vsancluster *cluster.ClusterConfig, timeStamp string, acc telegraf.Accumulator) error {
	for _, vsanhost := range vsancluster.Hosts {
		vsanhost := vsanhost.(*esxi.EsxiHostConfig)

		// VsanDatastore Contains VmDiskGroup
		for _, vsandatastore := range vsanhost.Datastores {
			if strings.Contains(vsandatastore.Uuid, "vsan") {
				vsandatastoreNode, err := genNode("VMDatastore", vsandatastore.DomainID(), vsandatastore.Name, "")
				if err != nil {
					return err
				}

				vsandiskgroupNode, err := genNode("VMVSanDiskGroup", vsanhost.DomainID(), vsanhost.Name, "")
				if err != nil {
					return err
				}

				accNeo4jMap(vsandatastoreNode, vsandiskgroupNode, "VsanDatastoreContainsVmDiskGroup", timeStamp, acc)

			}
		}

		for _, vsandiskgroup := range vsanhost.VsanDiskgroups {
			vsandiskgroupNode, err := genNode("VMVSanDiskGroup", vsandiskgroup.DomainID(), vsandiskgroup.Name, "")
			if err != nil {
				return err
			}
			accNeo4jMap(vsanclusterNode, vsandiskgroupNode, "VSanClusterContainsVSanDiskGroup", timeStamp, acc)

			// VSanDiskGroup -> VMDisk(Capacity/Cache)
			err = vSanDiskGroupContainsVMDiskCache(vsandiskgroupNode, vsandiskgroup, timeStamp, acc)
			if err != nil {
				return err
			}

			err = vSanDiskGroupContainsVMDiskCapacity(vsandiskgroupNode, vsandiskgroup, timeStamp, acc)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

// VSanDiskGroup -> VMDisk(Cache)
func vSanDiskGroupContainsVMDiskCache(vsandiskgroupNode *nodeInfo, vsandiskgroup *vsandiskgroup.VsanDiskgroupInfo, timeStamp string, acc telegraf.Accumulator) error {
	for _, vmdiskcache := range vsandiskgroup.CacheDisks {
		vmdiskcacheNode, err := genNode("VMDisk", vmdiskcache.DomainID(), vmdiskcache.Name, "")
		if err != nil {
			return err
		}
		accNeo4jMap(vsandiskgroupNode, vmdiskcacheNode, "VSanDiskGroupHasCacheVmDisk", timeStamp, acc)
	}
	return nil
}

// VSanDiskGroup -> VMDisk(Capacity)
func vSanDiskGroupContainsVMDiskCapacity(vsandiskgroupNode *nodeInfo, vsandiskgroup *vsandiskgroup.VsanDiskgroupInfo, timeStamp string, acc telegraf.Accumulator) error {
	for _, vmdiskcapacity := range vsandiskgroup.CapacityDisks {
		vmdiskcapacityNode, err := genNode("VMDisk", vmdiskcapacity.DomainID(), vmdiskcapacity.Name, "")
		if err != nil {
			return err
		}
		accNeo4jMap(vsandiskgroupNode, vmdiskcapacityNode, "VSanDiskGroupHasCapacityVmDisk", timeStamp, acc)
	}
	return nil
}

// VmHostContainsVmDisk
func vmHostContainsVMDisk(vcsaNode, vmhostNode *nodeInfo, vmhost host.HostConfig, timeStamp string, acc telegraf.Accumulator) error {
	for _, vmdisk := range vmhost.(*esxi.EsxiHostConfig).Disks {
		vmdiskNode, err := genNode("VMDisk", vmdisk.DomainID(), vmdisk.Name, "")
		if err != nil {
			return err
		}
		accNeo4jMap(vmhostNode, vmdiskNode, "VmHostContainsVmDisk", timeStamp, acc)
	}
	return nil
}

// VmHostContainsVmDisk
func vmHostContainsVMDatastore(vcsaNode, vmhostNode *nodeInfo, vmhost host.HostConfig, timeStamp string, acc telegraf.Accumulator) error {
	// VMDataCenter -> global -> VMHost -> VMDatastore
	for _, vmdatastore := range vmhost.(*esxi.EsxiHostConfig).Datastores {
		vmdatastoreNode, err := genNode("VMDatastore", vmdatastore.DomainID(), vmdatastore.Name, "")
		if err != nil {
			return err
		}
		// VmClusterCenter contains VmDatastore
		accNeo4jMap(vcsaNode, vmdatastoreNode, "VmClusterContainsVmDatastore", timeStamp, acc)
		accNeo4jMap(vmhostNode, vmdatastoreNode, "VmHostHasVmDatastore", timeStamp, acc)

		// VMDatastore contains VMDisk
		err = vmDatastoreComposesOfVMDisk(vmdatastoreNode, vmdatastore, timeStamp, acc)
		if err != nil {
			return err
		}
	}
	return nil
}

// VmHostHasVmDiskGroup
func vmHostHasVMDiskGroup(vcsaNode, vmhostNode *nodeInfo, vmhost host.HostConfig, timeStamp string, acc telegraf.Accumulator) error {
	// VMDataCenter -> global -> VMHost -> VMVSanDiskGroup
	for _, vsandiskgroup := range vmhost.(*esxi.EsxiHostConfig).VsanDiskgroups {
		vsandiskgroupNode, err := genNode("VMVSanDiskGroup", vsandiskgroup.DomainID(), vsandiskgroup.Name, "")
		if err != nil {
			return err
		}
		accNeo4jMap(vmhostNode, vsandiskgroupNode, "VmHostHasVmDiskGroup", timeStamp, acc)
	}
	return nil
}

// VmHostHostsVmVirtualMachine
func vmHostHostsVMVirtualMachine(vcsaNode, vmhostNode *nodeInfo, vmhost host.HostConfig, timeStamp string, acc telegraf.Accumulator) error {
	// VMDataCenter -> global -> VMHost -> VMVirtualMachine
	for _, vm := range vmhost.(*esxi.EsxiHostConfig).VMs {
		vmNode, err := genNode("VMVirtualMachine", vm.DomainID(), vm.Name, "")
		if err != nil {
			return err
		}
		accNeo4jMap(vmhostNode, vmNode, "VmHostHostsVmVirtualMachine", timeStamp, acc)

		// VmVirtualMachineContains
		err = vmVirtualMachineContainsVMDatastore(vmNode, vm, timeStamp, acc)
		if err != nil {
			return err
		}

		err = vmVirtualMachineContainsVMSnapshot(vmNode, vm, timeStamp, acc)
		if err != nil {
			return err
		}
	}
	return nil
}

// VMDatastore Composes Of VMDisk
func vmDatastoreComposesOfVMDisk(vmdatastoreNode *nodeInfo, vmdatastore *datastore.DatastoreInfo, timeStamp string, acc telegraf.Accumulator) error {
	for _, vmdisk := range vmdatastore.Disks {
		vmdiskNode, err := genNode("VMDisk", vmdisk.DomainID(), vmdisk.Name, "")
		if err != nil {
			return err
		}
		accNeo4jMap(vmdatastoreNode, vmdiskNode, "VmDatastoreComposesOfVmDisk", timeStamp, acc)
	}
	return nil
}

// VMVirtualmachine -> VmDataStore
func vmVirtualMachineContainsVMDatastore(vmNode *nodeInfo, vm *virtualmachine.VirtualmachineInfo, timeStamp string, acc telegraf.Accumulator) error {
	// VMDataCenter -> global -> VMHost -> VMVirtualMachine -> VMDatastore
	for _, vmdatastore := range vm.Datastores {
		vmdatastoreNode, err := genNode("VMDatastore", vmdatastore.DomainID(), vmdatastore.Name, "")
		if err != nil {
			return err
		}
		accNeo4jMap(vmNode, vmdatastoreNode, "VmVirtualMachineUsesVmDatastore", timeStamp, acc)
	}
	return nil
}

// VMVirtualmachine -> VmSnapShot
func vmVirtualMachineContainsVMSnapshot(vmNode *nodeInfo, vm *virtualmachine.VirtualmachineInfo, timeStamp string, acc telegraf.Accumulator) error {
	// VMDataCenter -> global -> VMHost -> VMVirtualMachine -> VMSnapshot
	for _, vmsnapshot := range vm.Snapshots {
		vmsnapshotNode, err := genNode("VMSnapshot", vmsnapshot.DomainID(), vmsnapshot.Name, "")
		if err != nil {
			return err
		}
		accNeo4jMap(vmNode, vmsnapshotNode, "VmVirtualMachineTakesVmSnapshot", timeStamp, acc)
	}
	return nil
}

