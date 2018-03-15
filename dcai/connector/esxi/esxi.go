package esxi

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf/dcai/hardware/disk"
	"github.com/influxdata/telegraf/dcai/util"
)

const (
	defaultSmartctlPath = "/opt/smartmontools/smartctl"
)

var (
	wwnExp       = regexp.MustCompile("^(naa\\.\\w.+)$")
	iSSASExp     = regexp.MustCompile("^\\s*Is SAS:\\s*(\\w.+)\\s*$")
	hwidExp      = regexp.MustCompile("^\\s*UUID:\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)\\s*0x([[:alnum:]]+)$")
	iSOfflineExp = regexp.MustCompile("^\\s*Is Offline:\\s*(\\w.+)\\s*$")
)

type EsxiConnector struct {
	Address      string
	Username     string
	Password     string
	SmartctlPath string
}

func NewEsxiConnector(addr string, user string, pw string, path string) (*EsxiConnector, error) {
	esxi := new(EsxiConnector)
	esxi.Address = addr
	esxi.Username = user
	esxi.Password = pw
	esxi.SmartctlPath = path

	if esxi.SmartctlPath == "" {
		// try to locate smartctl
		out, err := util.ExecuteSshCmdwithtimeout(addr, user, pw, "which smartctl")
		if err != nil {
			// cannot find smartctl. Try default path
			esxi.SmartctlPath = defaultSmartctlPath
		} else {
			paths := strings.Split(string(out), "\n")
			if len(paths) < 2 {
				return nil, fmt.Errorf("cannot find smartctl")
			}
			esxi.SmartctlPath = strings.TrimSpace(paths[0])
		}
	}

	// Test if smartctl is valid
	_, err := util.ExecuteSshCmdwithtimeout(addr, user, pw, "ls "+esxi.SmartctlPath)
	if err != nil {
		return nil, fmt.Errorf("Invalid smartctl path")
	}

	return esxi, nil
}

func (ec *EsxiConnector) GetDiskPathList() ([]*disk.DiskHeaderType, error) {
	out, err := util.ExecuteSshCmdwithtimeout(ec.Address, ec.Username, ec.Password, "esxcli storage core device list | grep -v '.\\+naa\\.\\|.\\+t10\\.'")
	if err != nil {
		return nil, err
	}

	var (
		dhs          []*disk.DiskHeaderType
		naatxt       []byte
		naa          string
		issas        string
		isoffline    string
		nextnaastart int
	)

	seps := [][]byte{[]byte("naa."), []byte("t10.")}
	naastart := util.FindIndexOfText(out, seps)
	if naastart == -1 {
		return nil, fmt.Errorf("Cannot find any storage device from esxcli")
	}
	for nextnaastart != -1 {
		nextnaastart = util.FindIndexOfText(out[naastart+1:], seps)
		if nextnaastart != -1 {
			naatxt = out[naastart : naastart+1+nextnaastart]
		} else {
			naatxt = out[naastart:]
		}
		naastart = naastart + 1 + nextnaastart
		m := []*util.FindRegexpMatchAndSetType{
			{&naa, wwnExp},
			{&issas, iSSASExp},
			{&isoffline, iSOfflineExp},
		}

		util.FindRegexpMatchAndSet(string(naatxt), m)
		if naa == "" || isoffline == "true" {
			continue
		}
		if issas == "false" {
			dhs = append(dhs, disk.NewDiskHeaderType("/dev/disks/"+naa, "sat"))
		} else {
			dhs = append(dhs, disk.NewDiskHeaderType("/dev/disks/"+naa, "scsi"))
		}
	}

	return dhs, nil
}

func (ec *EsxiConnector) GetDiskSmartRawOutput(dh *disk.DiskHeaderType) string {

	// ignore error for trying to parse smartctl output even error code is not 0
	out, _ := util.ExecuteSshCmdwithtimeout(ec.Address, ec.Username, ec.Password, ec.SmartctlPath+" -xa --format=old -n never -d "+dh.Devtype+" "+dh.Devpath)
	return string(out)
}

func (ec *EsxiConnector) GetHostname() (string, error) {
	out, err := util.ExecuteSshCmdwithtimeout(ec.Address, ec.Username, ec.Password, "hostname")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (ec *EsxiConnector) GetHWID() (string, error) {
	out, err := util.ExecuteSshCmdwithtimeout(ec.Address, ec.Username, ec.Password, "esxcli hardware platform get | grep UUID")
	if err != nil {
		return "", err
	}
	str := strings.TrimSpace(string(out))
	id := hwidExp.FindStringSubmatch(str)
	if len(id) < 17 {
		return "", fmt.Errorf("cannot parse %s", str)
	}

	hwid := fmt.Sprintf("%02s%02s%02s%02s-%02s%02s-%02s%02s-%02s%02s-%02s%02s%02s%02s%02s%02s", id[1], id[2], id[3], id[4], id[5], id[6], id[7], id[8], id[9], id[10], id[11], id[12], id[13], id[14], id[15], id[16])

	return hwid, nil
}
