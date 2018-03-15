package dmidecode

import (
	"github.com/influxdata/telegraf/dcai/testutil"
	"testing"
)

var (
	fakeDmidecodeCmdPath = "/fakeDmidecodeCmdPath"
	systeminput          = `# dmidecode 3.0
Getting SMBIOS data from sysfs.
SMBIOS 2.5 present.

Handle 0x0100, DMI type 1, 27 bytes
System Information
        Manufacturer: Dell Inc.
        Product Name: PowerEdge 1950
        Version: Not Specified
        Serial Number: D46ZMF1
        UUID: 44454C4C-3400-1036-805A-C4C04F4D4631
        Wake-up Type: Power Switch
        SKU Number: Not Specified
        Family: Not Specified

Handle 0x0C00, DMI type 12, 5 bytes
System Configuration Options
        Option 1: NVRAM_CLR:  Clear user settable NVRAM areas and set defaults
        Option 2: PWRD_EN:  Close to enable password

Handle 0x2000, DMI type 32, 11 bytes
System Boot Information
        Status: No errors detected`

	baseboardinput = `# dmidecode 3.0
Getting SMBIOS data from sysfs.
SMBIOS 2.5 present.

Handle 0x0200, DMI type 2, 9 bytes
Base Board Information
        Manufacturer: Dell Inc.
        Product Name: 0TT740
        Version: A00
        Serial Number: ..CN6970281N6242.
        Asset Tag: Not Specified

Handle 0x0A00, DMI type 10, 10 bytes
On Board Device 1 Information
        Type: Video
        Status: Enabled
        Description: Embedded ATI ES1000 Video
On Board Device 2 Information
        Type: Ethernet
        Status: Enabled
        Description: Embedded Broadcom 5708 NIC 1
On Board Device 3 Information
        Type: Ethernet
        Status: Enabled
        Description: Embedded Broadcom 5708 NIC 2

Handle 0x2900, DMI type 41, 11 bytes
Onboard Device
        Reference Designation: Embedded NIC 1
        Type: Ethernet
        Status: Enabled
        Type Instance: 1
        Bus Address: 0000:03:00.0

Handle 0x2901, DMI type 41, 11 bytes
Onboard Device
        Reference Designation: Embedded NIC 2
        Type: Ethernet
        Status: Enabled
        Type Instance: 2
        Bus Address: 0000:07:00.0

`
	dmisystem    = &DmiSystemInfo{"Dell Inc.", "PowerEdge 1950", "D46ZMF1"}
	dmibaseboard = &DmiBaseboardInfo{"Dell Inc.", "0TT740", "..CN6970281N6242."}
)

func fakeExecCommand(cmd string, args ...string) ([]byte, error) {
	switch args[0] {
	case "-t system":
		return []byte(systeminput), nil
	case "-t baseboard":
		return []byte(baseboardinput), nil
	default:
		return nil, nil
	}
}

func fakeCheckCommandPermission(cmd string) (bool, error) {
	return true, nil
}

func TestNewDmiSystemInfo(t *testing.T) {
	execcmd = fakeExecCommand
	execCmdCheckPermission = fakeCheckCommandPermission
	dmisys, err := newDmiSystemInfo(fakeDmidecodeCmdPath)
	if err != nil {
		t.Errorf("newDmiSystemInfo return error (%s)", err)
	}

	testutil.CompareVar(t, dmisys, dmisystem)
}

func TestNewDmiBaseboardInfo(t *testing.T) {
	execcmd = fakeExecCommand
	execCmdCheckPermission = fakeCheckCommandPermission
	dmibb, err := newDmiBaseboardInfo(fakeDmidecodeCmdPath)
	if err != nil {
		t.Errorf("newDmiSystemInfo return error (%s)", err)
	}

	testutil.CompareVar(t, dmibb, dmibaseboard)
}
