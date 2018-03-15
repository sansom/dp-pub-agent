package agent

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/dcai"
	"github.com/influxdata/telegraf/metric"
)

const agenttype = "Agent"

func getNewMeasurementMap() map[string]int {
	return map[string]int{
		"cpu":            0,
		"disk":           0,
		"diskio":         0,
		"sai_host":       0,
		"hostsystem":     0,
		"mem":            0,
		"net":            0,
		"sai_disk_smart": 0,
		"system":         0,
		"sys":            0,
	}
}

// AddSaiAgentMetric add a sai_agent metric
func AddSaiAgentMetric(metrics *[]telegraf.Metric) error {

	d, err := dcai.GetDcaiAgent()
	if err != nil {
		return err
	}

	h, err := dcai.FetchAgentHostConfig(d.Agenttype, d.TelegrafConfig.Agent.DmidecodePath)
	if err != nil {
		return err
	}

	measurements := getNewMeasurementMap()

	for _, metric := range *metrics {
		if v, exist := measurements[metric.Name()]; exist {
			count := v
			measurements[metric.Name()] = count + 1
		}
	}

	saiAgentTags := map[string]string{}
	saiAgentFields := make(map[string]interface{})
	saiAgentTags["agent_domain_id"] = h.IPv4s()
	saiAgentTags["cluster_domain_id"] = d.GetSaiClusterDomainId()
	saiAgentFields["agent_type"] = agenttype
	saiAgentFields["heartbeat_interval"] = int64(d.TelegrafConfig.Agent.Interval.Duration / time.Second)
	saiAgentFields["host_ip"] = h.IPv4s()
	saiAgentFields["host_ipv6"] = h.IPv6s()
	saiAgentFields["host_name"] = h.Hostname()
	saiAgentFields["needs_warning"] = false
	saiAgentFields["send"] = time.Now().UnixNano() / int64(time.Millisecond)
	saiAgentFields["agent_version"] = d.GetTelegrafVersion()

	fields := map[string]bool{
		"is_cpu_error":              noOutput(measurements, []string{"cpu", "hostsystem"}),
		"is_disk_error":             noOutput(measurements, []string{"disk", "hostsystem"}),
		"is_diskio_error":           noOutput(measurements, []string{"diskio", "hostsystem"}),
		"is_host_error":             noOutput(measurements, []string{"sai_host"}),
		"is_memory_error":           noOutput(measurements, []string{"mem", "hostsystem"}),
		"is_network_error":          noOutput(measurements, []string{"net", "hostsystem"}),
		"is_normal":                 false,
		"is_normalized_smart_error": false,
		"is_raw_smart_error":        noOutput(measurements, []string{"sai_disk_smart"}),
		"is_system_error":           noOutput(measurements, []string{"sys", "system", "hostsystem"}),
	}

	errorCount := 0
	for k, v := range fields {
		saiAgentFields[k] = v
		if v != false {
			errorCount = errorCount + 1
		}
	}

	if errorCount != 0 {
		saiAgentFields["is_error"] = true
	} else {
		saiAgentFields["is_error"] = false
	}

	for k, v := range d.TelegrafConfig.Tags {
		saiAgentTags[k] = v
	}

	t := time.Now()
	m, err := metric.New("sai_agent", saiAgentTags, saiAgentFields, t)
	if err != nil {
		return err
	}

	*metrics = append(*metrics, m)

	return nil
}

func noOutput(measurementMap map[string]int, names []string) bool {
	for _, name := range names {
		if measurementMap[name] > 0 {
			return false
		}
	}
	return true
}
