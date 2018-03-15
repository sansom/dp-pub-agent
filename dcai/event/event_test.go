package event

import (
	"testing"

	"github.com/influxdata/telegraf/dcai"
	"github.com/influxdata/telegraf/dcai/hardware/linux/dmidecode"
	"github.com/influxdata/telegraf/dcai/hardware/nic"
	"github.com/influxdata/telegraf/dcai/topology/host/linux"
	"github.com/influxdata/telegraf/dcai/type"
	"github.com/influxdata/telegraf/internal/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mockConfig = &config.Config{
		Agent: &config.AgentConfig{
			AgentType: "linux",
		},
	}

	linuxhost = linux.LinuxHostConfig{
		Name:      "prophetstor-Naomi",
		OSType:    dcaitype.OSLinux,
		OSName:    "Ubuntu",
		OSVersion: "16.04.3 LTS (Xenial Xerus)",
		CPUs:      nil,
		MEMs:      nil,
		NICs: []*nic.NetworkInfo{
			&nic.NetworkInfo{Name: "enp5s0f1", MACs: []string{"08:9e:01:c4:96:b4"}, IPv4s: []string{"172.31.86.223"}, IPv6s: []string{"fe80::cc6a:18a8:4cf9:a8a3"}},
			&nic.NetworkInfo{Name: "wlp4s0", MACs: []string{"0c:84:dc:5d:bb:91"}, IPv4s: []string{"172.31.86.129"}, IPv6s: []string{"fe80::305:cbe4:a283:9b16"}},
		},
		DmiInfo: &dmidecode.HostDmiInfo{
			Baseboard: &dmidecode.DmiBaseboardInfo{Manufacturer: "Acer", ProductName: "Dazzle_HW", SerialNumber: "NBM9W11003328034907600"},
			System:    &dmidecode.DmiSystemInfo{Manufacturer: "Acer", ProductName: "Aspire V5-573G", SerialNumber: "NXMC5TA001328034907600"},
		},
	}
)

func TestSaiEventDataPoint(t *testing.T) {

	var acc testutil.Accumulator

	dcai.NewDcaiAgent(mockConfig, "1.5.0", "", "test", "")
	err := SendMetricsMonitoring(
		&acc,
		&linuxhost,
		"dpCluster",
		"1 point(s) of host are written to DB",
		dcaitype.EventTitleHostDataSent,
		dcaitype.LogLevelInfo,
	)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("sai_event"), "expected has measurement called sai_event")
}
