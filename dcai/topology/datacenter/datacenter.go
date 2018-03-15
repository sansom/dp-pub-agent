package datacenter

import (
	"fmt"

	"github.com/influxdata/telegraf/dcai/topology/cluster"
	"github.com/influxdata/telegraf/dcai/topology/host"
)

const (
	DatacenterDefaultName = "Global"
)

type DatacenterConfig struct {
	Name     string
	Clusters []*cluster.ClusterConfig
	Hosts    []host.HostConfig
}

func NewDatacenterConfig(name string, clusters []*cluster.ClusterConfig, hosts []host.HostConfig) (*DatacenterConfig, error) {
	d := new(DatacenterConfig)
	if len(name) > 0 {
		d.Name = name
	} else {
		d.Name = DatacenterDefaultName
	}
	if clusters != nil {
		d.Clusters = clusters
	} else {
		d.Clusters = make([]*cluster.ClusterConfig, 0)
	}
	if hosts != nil {
		d.Hosts = hosts
	} else {
		d.Hosts = make([]host.HostConfig, 0)
	}
	return d, nil
}

func (d *DatacenterConfig) DatacenterName() string {
	return d.Name
}

func (d *DatacenterConfig) DomainID() string {
	return d.Name
}

func (d *DatacenterConfig) AppendCluster(c *cluster.ClusterConfig) error {
	if c != nil {
		d.Clusters = append(d.Clusters, c)
		return nil
	} else {
		return fmt.Errorf("Appending a nil cluster.")
	}
}

func (d *DatacenterConfig) AppendHost(h host.HostConfig) error {
	if h != nil {
		d.Hosts = append(d.Hosts, h)
		return nil
	} else {
		return fmt.Errorf("Appending a nil host.")
	}
}
