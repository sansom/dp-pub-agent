package mem

import (
	"github.com/influxdata/telegraf/dcai/testutil"
	"testing"
)

var (
	fakeDmidecodeCmdPath = "/fakeDmidecodeCmdPath"
	input                = `# dmidecode 3.0
Getting SMBIOS data from sysfs.
SMBIOS 2.5 present.

Handle 0x1000, DMI type 16, 15 bytes
Physical Memory Array
        Location: System Board Or Motherboard
        Use: System Memory
        Error Correction Type: Multi-bit ECC
        Maximum Capacity: 65280 MB
        Error Information Handle: Not Provided
        Number Of Devices: 8

Handle 0x1100, DMI type 17, 28 bytes
Memory Device
        Array Handle: 0x1000
        Error Information Handle: Not Provided
        Total Width: 72 bits
        Data Width: 64 bits
        Size: 4096 MB
        Form Factor: FB-DIMM
        Set: 1
        Locator: DIMM1
        Bank Locator: Not Specified
        Type: DDR2 FB-DIMM
        Type Detail: Synchronous
        Speed: 667 MHz
        Manufacturer: 802C7FB3802C
        Serial Number: D9152D22
        Asset Tag: 0C0744
        Part Number: 36HTF51272F667E1D4
        Rank: 2

Handle 0x1101, DMI type 17, 28 bytes
Memory Device
        Array Handle: 0x1000
        Error Information Handle: Not Provided
        Total Width: 72 bits
        Data Width: 64 bits
        Size: 4096 MB
        Form Factor: FB-DIMM
        Set: 1
        Locator: DIMM2
        Bank Locator: Not Specified
        Type: DDR2 FB-DIMM
        Type Detail: Synchronous
        Speed: 667 MHz
        Manufacturer: 802C7FB3802C
        Serial Number: D9152D1D
        Asset Tag: 0C0744
        Part Number: 36HTF51272F667E1D4
        Rank: 2

Handle 0x1102, DMI type 17, 28 bytes
Memory Device
        Array Handle: 0x1000
        Error Information Handle: Not Provided
        Total Width: 72 bits
        Data Width: 64 bits
        Size: 4096 MB
        Form Factor: FB-DIMM
        Set: 2
        Locator: DIMM3
        Bank Locator: Not Specified
        Type: DDR2 FB-DIMM
        Type Detail: Synchronous
        Speed: 667 MHz
        Manufacturer: 802C7FB3802C
        Serial Number: D9152D1A
        Asset Tag: 0C0744
        Part Number: 36HTF51272F667E1D4
        Rank: 2

Handle 0x1103, DMI type 17, 28 bytes
Memory Device
        Array Handle: 0x1000
        Error Information Handle: Not Provided
        Total Width: 72 bits
        Data Width: 64 bits
        Size: 4096 MB
        Form Factor: FB-DIMM
        Set: 2
        Locator: DIMM4
        Bank Locator: Not Specified
        Type: DDR2 FB-DIMM
        Type Detail: Synchronous
        Speed: 667 MHz
        Manufacturer: 802C7FB3802C
        Serial Number: D9152D1E
        Asset Tag: 0C0744
        Part Number: 36HTF51272F667E1D4
        Rank: 2

Handle 0x1104, DMI type 17, 28 bytes
Memory Device
        Array Handle: 0x1000
        Error Information Handle: Not Provided
        Total Width: 72 bits
        Data Width: 64 bits
        Size: 4096 MB
        Form Factor: FB-DIMM
        Set: 3
        Locator: DIMM5
        Bank Locator: Not Specified
        Type: DDR2 FB-DIMM
        Type Detail: Synchronous
        Speed: 667 MHz
        Manufacturer: 802C7FB3802C
        Serial Number: D9152D1F
        Asset Tag: 0C0744
        Part Number: 36HTF51272F667E1D4
        Rank: 2

Handle 0x1105, DMI type 17, 28 bytes
Memory Device
        Array Handle: 0x1000
        Error Information Handle: Not Provided
        Total Width: 72 bits
        Data Width: 64 bits
        Size: 4096 MB
        Form Factor: FB-DIMM
        Set: 3
        Locator: DIMM6
        Bank Locator: Not Specified
        Type: DDR2 FB-DIMM
        Type Detail: Synchronous
        Speed: 667 MHz
        Manufacturer: 802C7FB3802C
        Serial Number: D9152D20
        Asset Tag: 0C0744
        Part Number: 36HTF51272F667E1D4
        Rank: 2

Handle 0x1106, DMI type 17, 28 bytes
Memory Device
        Array Handle: 0x1000
        Error Information Handle: Not Provided
        Total Width: 72 bits
        Data Width: 64 bits
        Size: 4096 MB
        Form Factor: FB-DIMM
        Set: 4
        Locator: DIMM7
        Bank Locator: Not Specified
        Type: DDR2 FB-DIMM
        Type Detail: Synchronous
        Speed: 667 MHz
        Manufacturer: 0D9B8010802C
        Serial Number: 00000000
        Asset Tag: 000930
        Part Number:
        Rank: 2

Handle 0x1107, DMI type 17, 28 bytes
Memory Device
        Array Handle: 0x1000
        Error Information Handle: Not Provided
        Total Width: 72 bits
        Data Width: 64 bits
        Size: 4096 MB
        Form Factor: FB-DIMM
        Set: 4
        Locator: DIMM8
        Bank Locator: Not Specified
        Type: DDR2 FB-DIMM
        Type Detail: Synchronous
        Speed: 667 MHz
        Manufacturer: 05F785510000
        Serial Number: 00000000
        Asset Tag: 000000
        Part Number:
        Rank: 2

`
	mems = []*MemInfo{
		&MemInfo{"802C7FB3802C", "D9152D22", "0C0744", "36HTF51272F667E1D4", "DDR2 FB-DIMM", "Synchronous", "667 MHz", "DIMM1", "4096"},
		&MemInfo{"802C7FB3802C", "D9152D1D", "0C0744", "36HTF51272F667E1D4", "DDR2 FB-DIMM", "Synchronous", "667 MHz", "DIMM2", "4096"},
		&MemInfo{"802C7FB3802C", "D9152D1A", "0C0744", "36HTF51272F667E1D4", "DDR2 FB-DIMM", "Synchronous", "667 MHz", "DIMM3", "4096"},
		&MemInfo{"802C7FB3802C", "D9152D1E", "0C0744", "36HTF51272F667E1D4", "DDR2 FB-DIMM", "Synchronous", "667 MHz", "DIMM4", "4096"},
		&MemInfo{"802C7FB3802C", "D9152D1F", "0C0744", "36HTF51272F667E1D4", "DDR2 FB-DIMM", "Synchronous", "667 MHz", "DIMM5", "4096"},
		&MemInfo{"802C7FB3802C", "D9152D20", "0C0744", "36HTF51272F667E1D4", "DDR2 FB-DIMM", "Synchronous", "667 MHz", "DIMM6", "4096"},
		&MemInfo{"0D9B8010802C", "00000000", "000930", "", "DDR2 FB-DIMM", "Synchronous", "667 MHz", "DIMM7", "4096"},
		&MemInfo{"05F785510000", "00000000", "000000", "", "DDR2 FB-DIMM", "Synchronous", "667 MHz", "DIMM8", "4096"},
	}
)

func fakeExecCommand(cmd string, args ...string) ([]byte, error) {
	return []byte(input), nil
}

func fakeCheckCommandPermission(cmd string) (bool, error) {
	return true, nil
}

func cmdReturnNoMemoryData(cmd string, args ...string) ([]byte, error) {
	var output = `# dmidecode 3.0
Getting SMBIOS data from sysfs.
SMBIOS 2.5 present.
`
	return []byte(output), nil
}

func TestNewAllMeminfoWithoutData(t *testing.T) {
	execcmd = cmdReturnNoMemoryData
	checkCmdPermission = fakeCheckCommandPermission
	_, err := NewAllMemInfo(fakeDmidecodeCmdPath)
	if err != nil {
		if err.Error() != "Cannot find any memory device" {
			t.Errorf("NewAllCpuInfo return error (%s)", err)
		}
	} else {
		t.Errorf("NewAllCpuInfo should return error")
	}
}

func TestNewAllMemInfo(t *testing.T) {
	execcmd = fakeExecCommand
	checkCmdPermission = fakeCheckCommandPermission
	funcout, err := NewAllMemInfo(fakeDmidecodeCmdPath)
	if err != nil {
		t.Errorf("NewAllCpuInfo return error (%s)", err)
		return
	}

	testutil.CompareVar(t, funcout, mems)
}
