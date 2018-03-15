package cluster

import (
	"github.com/influxdata/telegraf"
)

const (
	SaiClusterDefaultDomainId = "dpCluster"
	SaiClusterDefaultName     = "DiskProphet for Lab Test"
)

// CreateSaiClusterDataPoint create a data point of sai_cluster
func CreateSaiClusterDataPoint(
	acc telegraf.Accumulator,
	saiClusterDomainId string,
	saiClusterName string,
) {
	saiClID := SaiClusterDefaultDomainId
	if saiClusterDomainId != "" {
		saiClID = saiClusterDomainId
	}
	saiClName := SaiClusterDefaultName
	if saiClusterName != "" {
		saiClName = saiClusterName
	}
	clusterTags := map[string]string{}
	clusterFields := make(map[string]interface{})
	clusterTags["domain_id"] = saiClID
	clusterFields["name"] = saiClName
	acc.AddFields("sai_cluster", clusterFields, clusterTags)
}
