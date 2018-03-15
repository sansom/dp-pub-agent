package smart

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/dcai"
	esxicon "github.com/influxdata/telegraf/dcai/connector/esxi"
	"github.com/influxdata/telegraf/dcai/event"
	"github.com/influxdata/telegraf/dcai/hardware/disk"
	"github.com/influxdata/telegraf/dcai/topology/host/vmware/esxi"
	"github.com/influxdata/telegraf/dcai/type"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type VsphereSmart struct {
	Vspheres [][]string
}

var sampleConfig = `
  ## This plugin collects smart data from vsphere servers.
  ## To specify a vsphere server list, the syntax is as follows.
  ## vspheres = [
  ##		["vsphere_IP_or_DN", "username", "password", "smartctl_path(optional)"], 
  ##		... 
  ##	]
  ## e.g.
  ## vspheres = [
  ##		["192.168.0.1", "root", "password", "/opt/smartmontools/smartctl"], 
  ##		["192.168.0.2", "root", "password", "/opt/smartmontools/smartctl"] 
  ##	]
  ##
  vspheres = []
`

func (m *VsphereSmart) SampleConfig() string {
	return sampleConfig
}

func (m *VsphereSmart) Description() string {
	return "Collect metrics from vsphere storage devices supporting S.M.A.R.T."
}

func parseConfig(acc telegraf.Accumulator, m *VsphereSmart) ([]*esxicon.EsxiConnector, error) {

	var esxilist []*esxicon.EsxiConnector
	if m.Vspheres == nil {
		return nil, fmt.Errorf("Invalid vsphere server list")
	} else {
		for i, c := range m.Vspheres {
			if len(c) < 4 {
				return nil, fmt.Errorf("Insufficient vsphere server parameters at index %d", i)
			}
			ec, err := esxicon.NewEsxiConnector(c[0], c[1], c[2], c[3])
			if err != nil {
				acc.AddError(fmt.Errorf("%s. Skip esxi at index %d at this time.", err, i))
			} else {
				esxilist = append(esxilist, ec)
			}
		}
	}
	return esxilist, nil
}

func (m *VsphereSmart) Gather(acc telegraf.Accumulator) error {

	dcaiAgent, err := dcai.GetDcaiAgent()
	if err != nil {
		return err
	}

	esxis, err := parseConfig(acc, m)
	if err != nil {
		return err
	}

	for _, ec := range esxis {
		disks, err := ec.GetDiskPathList()
		if err != nil {
			fmt.Printf("W! Cannot get disk list at %s. %s", ec.Address, err)
			continue
		}
		hostname, _ := ec.GetHostname()
		hostId, err := ec.GetHWID()
		if err != nil {
			return err
		}
		h, _ := esxi.NewEsxiHostConfig(hostname, hostId, "", "", nil, nil, nil, nil, nil, nil)

		for _, dh := range disks {
			smartRawOutput := ec.GetDiskSmartRawOutput(dh)
			d, err := disk.NewDiskInfoBySmartctlOutput(dh, smartRawOutput)
			if err != nil {
				return err
			}
			if !disk.IsValidDisk(d) {
				continue
			}

			disk.CollectSmartMetricsBySmartctlOutput(acc, dcaiAgent.GetSaiClusterDomainId(), h.DomainID(), dh, smartRawOutput)
			disk.CollectSaiDiskBySmartctlOutput(acc, dcaiAgent.GetSaiClusterDomainId(), h.DomainID(), dh, smartRawOutput)

			event.SendMetricsMonitoring(acc, h, h.DomainID(), fmt.Sprintf("1 point(s) of Host %s was written to DB", h.Name), dcaitype.EventTitleHostDataSent, dcaitype.LogLevelInfo)
		}
	}

	return nil
}

func init() {
	m := VsphereSmart{}
	inputs.Add("vspheresmart", func() telegraf.Input {
		return &m
	})
}
