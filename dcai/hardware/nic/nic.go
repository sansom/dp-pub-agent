package nic

import (
	"github.com/influxdata/telegraf/dcai/util"
	"regexp"
	"strings"
)

var (
	execcmd = util.ExecuteCmdWithTimeout
	nicNameRegexp			= regexp.MustCompile("^\\d+:\\s+([\\w@]+):.*$")
	nicMACRegexp			= regexp.MustCompile("\\s*link/ether\\s+([a-zA-Z0-9]{2}:[a-zA-Z0-9]{2}:[a-zA-Z0-9]{2}:[a-zA-Z0-9]{2}:[a-zA-Z0-9]{2}:[a-zA-Z0-9]{2}).*$")
	nicIPv4Regexp			= regexp.MustCompile("\\s*inet\\s+([0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+)/.*")
	nicIPv6Regexp			= regexp.MustCompile("\\s*inet6\\s+([0-9a-fA-F:]+)/.*")
)

type NetworkInfo struct {
	Name  string
	MACs  []string
	IPv4s []string
	IPv6s []string
}

func NewNetworkInfo(name string, macs []string, ipv4s []string, ipv6s []string) (*NetworkInfo, error) {

	n := new(NetworkInfo)
	n.Name = name
	n.MACs = macs
	n.IPv4s = ipv4s
	n.IPv6s = ipv6s

	return n, nil
}

func NewAllNetworkInfo() ([]*NetworkInfo, error) {
	var (
		nic      *NetworkInfo
		allNics  []*NetworkInfo
		out      []byte
		line     string
		strArray []string
		err      error
	)

	if out, err = execcmd("ip", "addr"); err != nil {
		return nil, err
	}

	for _, line = range strings.Split(string(out), "\n") {
		strArray = nicNameRegexp.FindStringSubmatch(line)
		if len(strArray) > 1 {
			// skip lo device
			if strArray[1] == "lo" {
				continue
			}
			if nic != nil && nic.Name != "" {
				allNics = append(allNics, nic)
			}
			nic = new(NetworkInfo)
			nic.Name = strArray[1]
		}

		if nic != nil {
			util.FindRegexpMatchAndAppend(line, &nic.MACs, nicMACRegexp)
			util.FindRegexpMatchAndAppend(line, &nic.IPv4s, nicIPv4Regexp)
			util.FindRegexpMatchAndAppend(line, &nic.IPv6s, nicIPv6Regexp)
		}
	}
	allNics = append(allNics, nic)

	return allNics, nil
}
