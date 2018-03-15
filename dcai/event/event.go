package event

import (
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/dcai"
	"github.com/influxdata/telegraf/dcai/topology/host"
	"github.com/influxdata/telegraf/dcai/type"
	"github.com/influxdata/telegraf/metric"
)

// SendFirstAgentHeartbeat send first heartbeat of Agent
func SendFirstAgentHeartbeat(acc telegraf.Accumulator) error {

	d, err := dcai.GetDcaiAgent()
	if err != nil {
		return err
	}

	outputs := []string{}
	inputs := []string{}
	tabOut := make(map[string]bool)
	tabIn := make(map[string]bool)
	for _, output := range d.TelegrafConfig.OutputNames() {
		if !tabOut[output] {
			tabOut[output] = true
			outputs = append(outputs, output)
		}
	}
	for _, input := range d.TelegrafConfig.InputNames() {
		if !tabIn[input] {
			tabIn[input] = true
			inputs = append(inputs, input)
		}
	}
	outputPlugins := strings.Join(outputs, " ")
	inputPlugins := strings.Join(inputs, " ")
	interval := int64(d.TelegrafConfig.Agent.Interval.Duration / time.Second)
	details := fmt.Sprintf("Agent query-interval-in-second: %d, Plugin outputs: %s, Plugin inputs: %s", interval, outputPlugins, inputPlugins)

	m, err := createSaiEventMetric(d, dcaitype.EventTypeFirstAgentHeartbeat, dcaitype.EventTitleAgentStarted, dcaitype.LogLevelInfo, details)
	if err != nil {
		return err
	}

	acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())

	return nil
}

func AddIntervalAgentHeartbeatMetric(metrics *[]telegraf.Metric) error {

	d, err := dcai.GetDcaiAgent()
	if err != nil {
		return err
	}

	interval := int64(d.TelegrafConfig.Agent.Interval.Duration / time.Second)
	details := fmt.Sprintf("Agent query-interval-in-second: %d", interval)

	m, err := createSaiEventMetric(d, dcaitype.EventTypeIntervalAgentHeartbeat, dcaitype.EventTitleAgentAlived, dcaitype.LogLevelInfo, details)
	if err != nil {
		return err
	}

	*metrics = append(*metrics, m)

	return nil
}

// SendMetricsMonitoring send metrics monitoring of Agent
func SendMetricsMonitoring(
	acc telegraf.Accumulator,
	host host.HostConfig,
	saiClusterDomainId string,
	details string,
	title dcaitype.EventTitle,
	level dcaitype.LogLevel,
) error {
	d, err := dcai.GetDcaiAgent()
	if err != nil {
		return err
	}

	m, err := createSaiEventMetric(d, dcaitype.EventTypeMetricsMonitoring, title, level, details)
	if err != nil {
		return err
	}

	acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())

	return nil
}

func createSaiEventMetric(
	d *dcai.DcaiAgent,
	eventType dcaitype.EventType,
	title dcaitype.EventTitle,
	level dcaitype.LogLevel,
	details string,
) (telegraf.Metric, error) {
	h, err := dcai.FetchAgentHostConfig(d.Agenttype, d.TelegrafConfig.Agent.DmidecodePath)
	if err != nil {
		return nil, err
	}
	t := time.Now()
	timestamp := t.UnixNano() / int64(time.Millisecond)

	saiEventTags := map[string]string{}
	saiEventFields := make(map[string]interface{})
	for k, v := range d.TelegrafConfig.Tags {
		saiEventTags[k] = v
	}
	saiEventFields["build_number"] = d.GetTelegrafVersion()
	saiEventFields["cluster_domain_id"] = d.GetSaiClusterDomainId()
	saiEventFields["details"] = details
	saiEventFields["event_level"] = level.String()
	saiEventFields["event_type"] = eventType.String()
	saiEventFields["host_ip"] = h.IPv4s()
	saiEventFields["host_ipv6"] = h.IPv6s()
	saiEventFields["timestamp"] = timestamp
	saiEventFields["title"] = title.String()
	saiEventFields["host_domain_id"] = h.HWID()

	return metric.New("sai_event", saiEventTags, saiEventFields, t)
}
