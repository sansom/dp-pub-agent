package smart

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/dcai"
	"github.com/influxdata/telegraf/dcai/event"
	"github.com/influxdata/telegraf/dcai/hardware/disk"
	"github.com/influxdata/telegraf/dcai/topology/host"
	"github.com/influxdata/telegraf/dcai/type"
	"github.com/influxdata/telegraf/dcai/util"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const SMARTCTL_CMD_TIMEOUT_SECOND = 10

var (
	execCommand             = util.ExecuteCmdWithTimeout
	checkSmartctlPermission = util.CheckCmdRootPermission
)

type Smart struct {
	Path       string
	Nocheck    string
	Attributes bool
	Excludes   []string
	Devices    []string
	UseSudo    bool
}

var sampleConfig = `
  ## Optionally specify the path to the smartctl executable
  # path = "/usr/bin/smartctl"
  #
  ## Skip checking disks in this power mode. Defaults to
  ## "standby" to not wake up disks that have stoped rotating.
  ## See --nocheck in the man pages for smartctl.
  ## smartctl version 5.41 and 5.42 have faulty detection of
  ## power mode and might require changing this value to
  ## "never" depending on your disks.
  ## Defaults to "never"
  ##
  # nocheck = "standby"
  #
`

func (m *Smart) SampleConfig() string {
	return sampleConfig
}

func (m *Smart) Description() string {
	return "Read metrics from storage devices supporting S.M.A.R.T."
}

func (m *Smart) Gather(acc telegraf.Accumulator) error {
	var err error
	if len(m.Path) == 0 {
		m.Path, err = util.GetCmdPathInOsPath("smartctl")
	} else {
		err = util.CheckCmdPath(m.Path)
	}
	if err != nil {
		return err
	}

	if _, err := checkSmartctlPermission("smartctl"); err != nil {
		return err
	}

	a, err := dcai.GetDcaiAgent()
	if err != nil {
		return err
	}
	h, err := dcai.FetchAgentHostConfig(a.Agenttype, a.TelegrafConfig.Agent.DmidecodePath)
	if err != nil {
		return err
	}

	err = m.getAttributes(acc, a.GetSaiClusterDomainId(), h)
	if err != nil {
		return err
	}
	return nil
}

/*// Wrap with sudo
func sudo(sudo bool, command string, args ...string) *exec.Cmd {
	if sudo {
		return execCommand("sudo", append([]string{"-n", command}, args...)...)
	}

	return execCommand(command, args...)
}

// Scan for S.M.A.R.T. devices
func (m *Smart) scan() ([]string, error) {

	cmd := sudo(m.UseSudo, m.Path, "--scan-open")
	out, err := internal.CombinedOutputTimeout(cmd, time.Second*5)
	if err != nil {
		return []string{}, fmt.Errorf("failed to run command %s: %s", strings.Join(cmd.Args, " "), err)
	}

	devices := []string{}
	for _, line := range strings.Split(string(out), "\n") {
		dev := strings.Split(line, "#")
		if dev[0] == "" {
			continue
		}
		if len(dev) > 1 && !excludedDev(m.Excludes, strings.TrimSpace(dev[0])) {
			devices = append(devices, strings.TrimSpace(dev[0]))
		}
	}
	return devices, nil
}*/

func excludedDev(excludes []string, deviceLine string) bool {
	device := strings.Split(deviceLine, " ")
	if len(device) != 0 {
		for _, exclude := range excludes {
			if device[0] == exclude {
				return true
			}
		}
	}
	return false
}

// Get info and attributes for each S.M.A.R.T. device
func (m *Smart) getAttributes(acc telegraf.Accumulator, saiClDomainID string, h host.HostConfig) error {
	devices, err := h.GetDisks(m.Path)
	if err != nil {
		return err
	}

	for _, device := range devices {
		gatherDisk(acc, saiClDomainID, h, m.UseSudo, m.Attributes, m.Path, m.Nocheck, device)
	}
	return nil
}

// Command line parse errors are denoted by the exit code having the 0 bit set.
// All other errors are drive/communication errors and should be ignored.
func exitStatus(err error) (int, error) {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
	}
	return 0, err
}

func gatherDisk(acc telegraf.Accumulator, saiClDomainID string, h host.HostConfig, usesudo, attributes bool, path, nockeck string, device *disk.DiskInfo) {

	// smartctl 5.41 & 5.42 have are broken regarding handling of --nocheck/-n

	disk.CollectSmartMetricsBySmartctlOutput(acc, saiClDomainID, h.DomainID(), device.Header, device.SmartctlOutput)

	event.SendMetricsMonitoring(acc, h, saiClDomainID, fmt.Sprintf("1 point(s) of %s was written to DB", device.GetName()), dcaitype.EventTitleSmartDataSent, dcaitype.LogLevelInfo)

	disk.CollectSaiDiskBySmartctlOutput(acc, saiClDomainID, h.DomainID(), device.Header, device.SmartctlOutput)
}

func parseRawValue(rawVal string) (int64, error) {

	// Integer
	if i, err := strconv.ParseInt(rawVal, 10, 64); err == nil {
		return i, nil
	}

	// Duration: 65h+33m+09.259s
	unit := regexp.MustCompile("^(.*)([hms])$")
	parts := strings.Split(rawVal, "+")
	if len(parts) == 0 {
		return 0, fmt.Errorf("Couldn't parse RAW_VALUE '%s'", rawVal)
	}

	duration := int64(0)
	for _, part := range parts {
		timePart := unit.FindStringSubmatch(part)
		if len(timePart) == 0 {
			continue
		}
		switch timePart[2] {
		case "h":
			duration += parseInt(timePart[1]) * int64(3600)
		case "m":
			duration += parseInt(timePart[1]) * int64(60)
		case "s":
			// drop fractions of seconds
			duration += parseInt(strings.Split(timePart[1], ".")[0])
		default:
			// Unknown, ignore
		}
	}
	return duration, nil
}

func parseInt(str string) int64 {
	if i, err := strconv.ParseInt(str, 10, 64); err == nil {
		return i
	}
	return 0
}

func init() {
	m := Smart{}
	m.Nocheck = "never"
	m.Attributes = true

	inputs.Add("smart", func() telegraf.Input {
		return &m
	})
}
