package dmidecode

import (
	"fmt"
	"github.com/influxdata/telegraf/dcai/util"
	"regexp"
)

var (
	execcmd                  = util.ExecuteSudoCmdWithTimeout
	execCmdCheckPermission   = util.CheckCmdRootPermission
	systemManufacturerRegexp = regexp.MustCompile("\\s*Manufacturer:\\s*(.*)\\s*$")
	systemProductNameRegexp  = regexp.MustCompile("\\s*Product Name:\\s*(.*)\\s*$")
	systemSerialNumberRegexp = regexp.MustCompile("\\s*Serial Number:\\s*(.*)\\s*$")

	baseboardManufacturerRegexp = regexp.MustCompile("\\s*Manufacturer:\\s*(.*)\\s*$")
	baseboardProductNameRegexp  = regexp.MustCompile("\\s*Product Name:\\s*(.*)\\s*$")
	baseboardSerialNumberRegexp = regexp.MustCompile("\\s*Serial Number:\\s*(.*)\\s*$")
)

type HostDmiInfo struct {
	Baseboard *DmiBaseboardInfo
	System    *DmiSystemInfo
}

type DmiBaseboardInfo struct {
	Manufacturer string
	ProductName  string
	SerialNumber string
}

type DmiSystemInfo struct {
	Manufacturer string
	ProductName  string
	SerialNumber string
}

func (d *DmiBaseboardInfo) String() string {
	return fmt.Sprintf("%s%s%s", d.Manufacturer, d.ProductName, d.SerialNumber)
}

func (d *DmiSystemInfo) String() string {
	return fmt.Sprintf("%s%s%s", d.Manufacturer, d.ProductName, d.SerialNumber)
}

func (h *HostDmiInfo) String() string {
	if h != nil {
		return fmt.Sprintf("%s%s", h.System, h.Baseboard)
	} else {
		return ""
	}
}

func NewHostDmiInfo(dmidecodePath string) (*HostDmiInfo, error) {
	b, err := newDmiBaseboardInfo(dmidecodePath)
	if err != nil {
		return nil, err
	}
	s, err := newDmiSystemInfo(dmidecodePath)
	if err != nil {
		return nil, err
	}
	dmi := HostDmiInfo{Baseboard: b, System: s}

	return &dmi, nil
}

func newDmiSystemInfo(dmidecodePath string) (*DmiSystemInfo, error) {
	if _, err := execCmdCheckPermission("dmidecode"); err != nil {
		return nil, err
	}
	out, err := execcmd(dmidecodePath, "-t system")
	if err != nil {
		return nil, err
	}

	system := new(DmiSystemInfo)
	m := []*util.FindRegexpMatchAndSetType{
		{&system.Manufacturer, systemManufacturerRegexp},
		{&system.ProductName, systemProductNameRegexp},
		{&system.SerialNumber, systemSerialNumberRegexp},
	}
	util.FindRegexpMatchAndSet(string(out), m)
	return system, nil
}

func newDmiBaseboardInfo(dmidecodePath string) (*DmiBaseboardInfo, error) {
	if _, err := execCmdCheckPermission("dmidecode"); err != nil {
		return nil, err
	}
	out, err := execcmd(dmidecodePath, "-t baseboard")
	if err != nil {
		return nil, err
	}

	baseboard := new(DmiBaseboardInfo)
	m := []*util.FindRegexpMatchAndSetType{
		{&baseboard.Manufacturer, baseboardManufacturerRegexp},
		{&baseboard.ProductName, baseboardProductNameRegexp},
		{&baseboard.SerialNumber, baseboardSerialNumberRegexp},
	}
	util.FindRegexpMatchAndSet(string(out), m)
	return baseboard, nil
}
