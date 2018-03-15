package disk

import (
	"testing"

	"github.com/influxdata/telegraf/dcai/testutil"
	"github.com/influxdata/telegraf/dcai/type"
)

type DiskDataTestcase struct {
	InputScanOpenData  string
	InputDiskMockData  string
	ExpectedOutputData *DiskInfo
}

type DiskTypeTestcase struct {
	SataVersion       string
	TransportProtocol string
	RotationRate      string
	ExpectedDiskType  dcaitype.DiskType
}

var (
	inputScanOpenData = []string{
		`/dev/sda -d sat`,
		`/dev/bus/2 -d megaraid,1`,
		`/dev/bus/2 -d sat+megaraid,0`,
	}

	inputDiskMockData = []string{
		`
smartctl 6.5 2016-01-24 r4214 [x86_64-linux-4.13.0-26-generic] (local build)
Copyright (C) 2002-16, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Model Family:     Western Digital Blue Mobile
Device Model:     WDC WD10JPCX-24UE4T0
Serial Number:    WD-WXL1A96K9KA7
LU WWN Device Id: 5 0014ee 60711e39c
Firmware Version: 01.01A01
User Capacity:    1,000,204,886,016 bytes [1.00 TB]
Sector Sizes:     512 bytes logical, 4096 bytes physical
Rotation Rate:    5400 rpm
Device is:        In smartctl database [for details use: -P show]
ATA Version is:   ACS-2 (minor revision not indicated)
SATA Version is:  SATA 3.0, 6.0 Gb/s (current: 6.0 Gb/s)
Local Time is:    Fri Jan 12 10:15:35 2018 CST
SMART support is: Available - device has SMART capability.
SMART support is: Enabled

=== START OF READ SMART DATA SECTION ===
SMART overall-health self-assessment test result: PASSED

`,
		`
smartctl 6.5 2016-01-24 r4214 [x86_64-linux-4.4.0-79-generic] (local build)
Copyright (C) 2002-16, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Vendor:               SEAGATE
Product:              ST3300657SS
Revision:             ES66
User Capacity:        300,000,000,000 bytes [300 GB]
Logical block size:   512 bytes
Rotation Rate:        15000 rpm
Form Factor:          3.5 inches
Logical Unit id:      0x5000c5005f50e6ab
Serial number:        6SJ6PWDK
Device type:          disk
Transport protocol:   SAS (SPL-3)
Local Time is:        Sun Dec 17 23:26:24 2017 PST
SMART support is:     Available - device has SMART capability.
SMART support is:     Enabled
Temperature Warning:  Disabled or Not Supported

=== START OF READ SMART DATA SECTION ===
SMART Health Status: OK

`,
		`
smartctl 6.5 2016-01-24 r4214 [x86_64-linux-4.4.0-79-generic] (local build)
Copyright (C) 2002-16, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Model Family:     Intel 730 and DC S35x0/3610/3700 Series SSDs
Device Model:     INTEL SSDSC2BP480G4
Serial Number:    BTJR51660055480BGN
LU WWN Device Id: 5 5cd2e4 04b7ee0e9
Firmware Version: L2010420
User Capacity:    480,103,981,056 bytes [480 GB]
Sector Size:      512 bytes logical/physical
Rotation Rate:    Solid State Device
Form Factor:      2.5 inches
Device is:        In smartctl database [for details use: -P show]
ATA Version is:   ATA8-ACS T13/1699-D revision 4
SATA Version is:  SATA 2.6, 6.0 Gb/s (current: 1.5 Gb/s)
Local Time is:    Mon Dec 18 01:00:54 2017 PST
SMART support is: Available - device has SMART capability.
SMART support is: Enabled

=== START OF READ SMART DATA SECTION ===
SMART Status not supported: ATA return descriptor not supported by controller firmware
SMART overall-health self-assessment test result: PASSED
Warning: This result is based on an Attribute check.

`,
	}

	disks = []DiskDataTestcase{
		DiskDataTestcase{
			inputScanOpenData[0],
			inputDiskMockData[0],
			&DiskInfo{&DiskHeaderType{"sda", "/dev/sda", "sat"}, "sda", inputDiskMockData[0], dcaitype.DiskStatusGood, dcaitype.DiskTypeHDDSATA, "", "01.01A01", "WDC WD10JPCX-24UE4T0", "SATA 3.0, 6.0 Gb/s (current: 6.0 Gb/s)", "512 bytes logical, 4096 bytes physical", "50014ee60711e39c", "WD-WXL1A96K9KA7", "1.00 TB", "PASSED", "Western Digital Blue Mobile", ""},
		},
		DiskDataTestcase{
			inputScanOpenData[1],
			inputDiskMockData[1],
			&DiskInfo{&DiskHeaderType{"2", "/dev/bus/2", "megaraid,1"}, "MegaraidDisk-1", inputDiskMockData[1], dcaitype.DiskStatusGood, dcaitype.DiskTypeHDDSAS, "disk", "ES66", "ST3300657SS", "", "512 bytes", "5000c5005f50e6ab", "6SJ6PWDK", "300 GB", "OK", "SEAGATE", "SAS (SPL-3)"},
		},
		DiskDataTestcase{
			inputScanOpenData[2],
			inputDiskMockData[2],
			&DiskInfo{&DiskHeaderType{"2", "/dev/bus/2", "sat+megaraid,0"}, "MegaraidDisk-0", inputDiskMockData[2], dcaitype.DiskStatusGood, dcaitype.DiskTypeSSDSATA, "", "L2010420", "INTEL SSDSC2BP480G4", "SATA 2.6, 6.0 Gb/s (current: 1.5 Gb/s)", "512 bytes logical/physical", "55cd2e404b7ee0e9", "BTJR51660055480BGN", "480 GB", "PASSED", "Intel 730 and DC S35x0/3610/3700 Series SSDs", ""},
		},
	}

	disktypes = []DiskTypeTestcase{
		DiskTypeTestcase{"SATA 3.0, 6.0 Gb/s (current: 6.0 Gb/s)", "", "5400 rpm", dcaitype.DiskTypeHDDSATA},
		DiskTypeTestcase{"", "SAS (SPL-3)", "5400 rpm", dcaitype.DiskTypeHDDSAS},
		DiskTypeTestcase{"SATA 3.0, 6.0 Gb/s (current: 6.0 Gb/s)", "", "Solid State Device", dcaitype.DiskTypeSSDSATA},
		DiskTypeTestcase{"", "SAS (SPL-3)", "Solid State Device", dcaitype.DiskTypeSSDSAS},
	}
)

func TestNewAllDiskSmartInfo(t *testing.T) {

	for _, disk := range disks {
		var (
			d   *DiskInfo
			err error
		)
		dh := NewDiskHeaderFromSmartctlScan(disk.InputScanOpenData)
		if dh != nil {
			d, err = NewDiskInfoBySmartctlOutput(dh, disk.InputDiskMockData)
			if err != nil {
				t.Errorf("NewAllDiskSmartInfo return error (%s)", err)
			}
		}
		testutil.CompareVar(t, d, disk.ExpectedOutputData)
	}
}

func TestDiskTypeCodeMatch(t *testing.T) {

	for _, disktype := range disktypes {

		funcout := diskTypeMatch(disktype.SataVersion, disktype.TransportProtocol, disktype.RotationRate)

		testutil.CompareVar(t, funcout, disktype.ExpectedDiskType)
	}
}
