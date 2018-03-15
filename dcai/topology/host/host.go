package host

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/dcai/hardware/disk"
	"github.com/influxdata/telegraf/dcai/type"
)

type HostConfig interface {
	GetOsType() dcaitype.OSType
	Hostname() string
	DomainID() string
	HWID() string
	IPv4s() string
	IPv6s() string
	GetDisks(string) ([]*disk.DiskInfo, error)
}

// CreateSaiHostDataPoint create a data point of sai_host
func CreateSaiHostDataPoint(
	acc telegraf.Accumulator,
	hostDomainID string, hostName string, hostHWID string,
	clusterDomainID string,
	osType dcaitype.OSType, osName string, osVersion string,
	ipv4 string, ipv6 string,
) {
	hostTags := map[string]string{}
	hostFields := make(map[string]interface{})
	hostTags["domain_id"] = hostDomainID
	hostFields["host_uuid"] = hostHWID
	hostFields["cluster_domain_id"] = clusterDomainID
	hostFields["name"] = hostName
	hostFields["os_type"] = osType.String()
	hostFields["os_name"] = osName
	hostFields["os_version"] = osVersion
	hostFields["host_ip"] = ipv4
	hostFields["host_ipv6"] = ipv6
	acc.AddFields("sai_host", hostFields, hostTags)
}
