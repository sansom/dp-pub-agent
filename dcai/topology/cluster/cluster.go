package cluster

import (
	"fmt"

	"github.com/influxdata/telegraf/dcai/topology/host"
	"github.com/influxdata/telegraf/dcai/type"
)

type ClusterConfig struct {
	Name        string
	Uuid        string
	ClusterType dcaitype.ClusterType
	Hosts       []host.HostConfig
}

func NewClusterConfig(t dcaitype.ClusterType, uuid string, name string) (*ClusterConfig, error) {
	c := new(ClusterConfig)
	c.ClusterType = t

	if uuid == "" {
		return nil, fmt.Errorf("cluster uuid cannot be empty")
	} else {
		c.Uuid = uuid
	}
	c.Name = name

	return c, nil
}

func (c *ClusterConfig) ClusterName() string {
	return c.Name
}

func (c *ClusterConfig) DomainID() string {
	return c.Uuid
}

func (c *ClusterConfig) AppendHost(h host.HostConfig) error {
	if h != nil {
		c.Hosts = append(c.Hosts, h)
		return nil
	} else {
		return fmt.Errorf("Appending a nil host")
	}
}
