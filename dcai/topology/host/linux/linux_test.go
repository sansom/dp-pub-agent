package linux

import (
	"github.com/influxdata/telegraf/dcai/hardware/linux/dmidecode"
	"github.com/influxdata/telegraf/dcai/hardware/nic"
	"github.com/influxdata/telegraf/dcai/testutil"
	"github.com/influxdata/telegraf/dcai/type"
	"testing"
)

var (
	osInfoInput = `
NAME="CentOS Linux"
VERSION="7 (Core)"
ID="centos"
ID_LIKE="rhel fedora"
VERSION_ID="7"
PRETTY_NAME="CentOS Linux 7 (Core)"
ANSI_COLOR="0;31"
CPE_NAME="cpe:/o:centos:centos:7"
HOME_URL="https://www.centos.org/"
BUG_REPORT_URL="https://bugs.centos.org/"

CENTOS_MANTISBT_PROJECT="CentOS-7"
CENTOS_MANTISBT_PROJECT_VERSION="7"
REDHAT_SUPPORT_PRODUCT="centos"
REDHAT_SUPPORT_PRODUCT_VERSION="7"

`
	expectedOsInfoOutput = LinuxHostConfig{
		Name:      "SomeHostName",
		OSType:    dcaitype.OSLinux,
		OSName:    "CentOS Linux",
		OSVersion: "7 (Core)",
		DmiInfo: &dmidecode.HostDmiInfo{
			Baseboard: &dmidecode.DmiBaseboardInfo{
				Manufacturer: "Intel Corporation",
				ProductName:  "440BX Desktop Reference Platform",
				SerialNumber: "None",
			},
			System: &dmidecode.DmiSystemInfo{
				Manufacturer: "VMware, Inc.",
				ProductName:  "VMware Virtual Platform",
				SerialNumber: "VMware-56 4d ee 09 48 78 cc df-d4 bc 03 82 75 cf 08 81",
			},
		},
		NICs: []*nic.NetworkInfo{
			&nic.NetworkInfo{
				Name:  "ens33",
				MACs:  []string{"00:0c:29:cf:08:81"},
				IPv4s: []string{"172.31.86.110"},
				IPv6s: []string{"fe80::65d5:aedc:ee1d:ecc8"}},
		},
	}

	expectedHWID = "fc2ba9fc7d7f94f3413a0b9ddb6e73d5"
)

func fakeExecCommand(cmd string, args ...string) ([]byte, error) {
	return []byte(osInfoInput), nil
}

func TestFetchLinuxOSInfo(t *testing.T) {
	execcmd = fakeExecCommand

	h := new(LinuxHostConfig)
	err := h.FetchLinuxOSInfo()
	if err != nil {
		t.Errorf("FetchLinuxOSInfo return error (%s)", err)
	}

	testutil.CompareVar(t, h.OSType, expectedOsInfoOutput.OSType)
	testutil.CompareVar(t, h.OSName, expectedOsInfoOutput.OSName)
	testutil.CompareVar(t, h.OSVersion, expectedOsInfoOutput.OSVersion)
}

func TestHWID(t *testing.T) {
	testutil.CompareVar(t, expectedOsInfoOutput.HWID(), expectedHWID)
}

func TestDomainID(t *testing.T) {
	testutil.CompareVar(t, expectedOsInfoOutput.DomainID(), expectedHWID)
}
