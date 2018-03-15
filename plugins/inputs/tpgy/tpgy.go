package tpgy

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/dcai"
	"github.com/influxdata/telegraf/dcai/event"
	saicluster "github.com/influxdata/telegraf/dcai/sai/cluster"
	"github.com/influxdata/telegraf/dcai/topology/host"
	"github.com/influxdata/telegraf/dcai/topology/host/linux"
	"github.com/influxdata/telegraf/dcai/type"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Tpgy plugin collects data from DcaiAgent
type Tpgy struct{}

var sampleConfig = `
  # no configuration
`

// Description returns description of tpgy plugin
func (t *Tpgy) Description() string {
	return "Enable tpgy input plugin"
}

// SampleConfig displays configuration instructions
func (t *Tpgy) SampleConfig() string {
	return sampleConfig
}

// Gather collects data from DcaiAgent and sends data points
func (t *Tpgy) Gather(acc telegraf.Accumulator) error {

	dcaiAgent, err := dcai.GetDcaiAgent()
	if err != nil {
		return err
	}

	h, err := dcai.FetchAgentHostConfig(dcaiAgent.Agenttype, dcaiAgent.TelegrafConfig.Agent.DmidecodePath)
	if err != nil {
		return err
	}

	err = t.gatherTpgy(acc, dcaiAgent.GetSaiClusterDomainId(), dcaiAgent.GetSaiClusterName(), h)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tpgy) gatherTpgy(acc telegraf.Accumulator, saiClusterDomainId string, saiClusterName string, h host.HostConfig) error {
	ostype := h.GetOsType()
	if ostype == dcaitype.OSLinux {
		linuxhost := h.(*linux.LinuxHostConfig)

		host.CreateSaiHostDataPoint(acc, linuxhost.DomainID(), linuxhost.Hostname(), linuxhost.HWID(), saiClusterDomainId, linuxhost.OSType, linuxhost.OSName, linuxhost.OSVersion, linuxhost.IPv4s(), linuxhost.IPv6s())
		saicluster.CreateSaiClusterDataPoint(acc, saiClusterDomainId, saiClusterName)
		event.SendMetricsMonitoring(acc, linuxhost, saiClusterDomainId, "1 point(s) of Host are written to DB", dcaitype.EventTitleHostDataSent, dcaitype.LogLevelInfo)
	}
	return nil
}

func init() {
	inputs.Add("tpgy", func() telegraf.Input { return &Tpgy{} })
}
