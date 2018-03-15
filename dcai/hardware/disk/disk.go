package disk

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/dcai/type"
	"github.com/influxdata/telegraf/dcai/util"
)

const (
	ARECA_MAX_ENCLOSURE_NUM = 2
	ARECA_MAX_SLOT_NUM      = 16
	CCISS_MAX_SLOT_NUM      = 128
)

var (
	sudoExeccmd        = util.ExecuteSudoCmdWithTimeout
	checkCmdPermission = util.CheckCmdRootPermission
	execcmd            = util.ExecuteCmdWithTimeout

	//Name
	megaraidDevice = regexp.MustCompile("^.*megaraid,([0-9]+)$")
	ccissDevice    = regexp.MustCompile("^.*cciss,([0-9]+)$")
	arecaDevice    = regexp.MustCompile("^.*areca,([0-9]+)/([0-9]+)$")
	//FirmwareVersion
	firmwareVersion = regexp.MustCompile("^Firmware Version:\\s+(.*)$")
	revision        = regexp.MustCompile("^Revision:\\s+(.*)$")
	//Model
	modelInInfo = regexp.MustCompile("^Device Model:\\s+(.*)$")
	product     = regexp.MustCompile("^Product:\\s+(.*)$")
	//SataVersion
	sataVersion = regexp.MustCompile("^SATA Version is:\\s+(.*)$")
	//SectorSize
	sectorSizes      = regexp.MustCompile("^Sector Size[s]?:\\s+(.*)$")
	logicalBlockSize = regexp.MustCompile("^Logical block size:\\s+(.*)$")
	//Type
	rotationRate = regexp.MustCompile("^Rotation Rate:\\s+(.*)$")
	//WWN
	wwnInInfo     = regexp.MustCompile("^LU WWN Device Id:\\s+(.*)$")
	logicalUnitId = regexp.MustCompile("^Logical Unit id:\\s+0x(.*)$")
	//SerialNumber
	serialInInfo = regexp.MustCompile("^Serial [nN]umber:\\s+(.*)$")
	//Size
	userCapacity = regexp.MustCompile("^User Capacity:\\s+[0-9,]+\\s+bytes\\s+\\[(.*)\\]$")
	//SmartHealthStatus
	smartOverallHealth = regexp.MustCompile("^SMART overall-health self-assessment test result:\\s+(\\w+).*$")
	smartHealthStatus  = regexp.MustCompile("^SMART Health Status:\\s+(\\w+).*$")
	//Vendor
	modelFamily = regexp.MustCompile("^Model Family:\\s+(.*)$")
	vendor      = regexp.MustCompile("^Vendor:\\s+(.*)$")
	//TransportProtocol
	transportProtocol = regexp.MustCompile("^Transport protocol:\\s+(.*)$")
	//Device type
	peripheralDeviceType = regexp.MustCompile("^Device type:\\s+(.*)$")
	// Current Drive Temperature:     37 C
	currentDriveTemperature = regexp.MustCompile("^Current.+Temperature:\\s+([0-9]+).*$")
	// Drive Trip Temperature:        68 C
	driveTripTemperature = regexp.MustCompile("^Drive Trip Temperature:\\s+([0-9]+).*$")
	// Elements in grown defect list: 0
	elementsInGrownDefectList = regexp.MustCompile("^Elements in grown defect list:\\s+([0-9]+)$")

	//    Accumulated power on time, hours:minutes 16389:51 [983391 minutes]
	powerOnHoursPatterns = []*regexp.Regexp{
		regexp.MustCompile("^\\s*Accumulated power on time,\\shours:minutes\\s([0-9]+):[0-9]+\\s.[0-9]+\\sminutes.*$"),
		regexp.MustCompile("^\\s*number of hours powered up\\s*=\\s*([0-9]+)[.0-9]*$"),
	}
	attribute = regexp.MustCompile("^\\s*([0-9]+)\\s+\\S+\\s+0x[0-9a-z]+\\s+[0-9]+\\s+[0-9]+\\s+[0-9]+\\s+\\S+\\s+\\w+\\s+[-\\w]+\\s+([\\w\\+\\.]+).*$")

	//       Errors Corrected by           Total   Correction     Gigabytes    Total
	//           ECC          rereads/    errors   algorithm      processed    uncorrected
	//       fast | delayed   rewrites  corrected  invocations   [10^9 bytes]  errors
	// read:    397365        0         0    397365     397365         78.328           0
	// write:        0        0         0         0          0         86.715           0
	errorCounterLog = regexp.MustCompile("^(\\w+)\\W\\s+([0-9]+)\\s+([0-9]+)\\s+([0-9]+)\\s+([0-9]+)\\s+([0-9]+)\\s+([0-9.]+)\\s+([0-9]+)$")

	smartHealthStatusTypesSATA = map[int]string{
		0: "UNKNOWN",
		1: "PASSED",
		2: "FAILED",
	}

	smartHealthStatusTypesSAS = map[int]string{
		0: "UNKNOWN",
		1: "OK",
		3: "WARNING",
		4: "CRITICAL",
	}
)

type DiskInfo struct {
	Header            *DiskHeaderType
	Name              string
	SmartctlOutput    string
	Status            dcaitype.DiskStatusType
	Type              dcaitype.DiskType
	PeripheralDevtype string
	FirmwareVersion   string
	Model             string
	SataVersion       string
	SectorSize        string
	WWN               string
	SerialNumber      string
	Size              string
	SmartHealthStatus string
	Vendor            string
	TransportProtocol string
}

func NewDiskInfo(name string, wwn string, vendor string, model string, fmver string, sataver string, sectorsize string, sn string, size string, t dcaitype.DiskType, status dcaitype.DiskStatusType, smartstatus string, transprotocol string) (*DiskInfo, error) {

	d := new(DiskInfo)
	d.Name = name
	d.WWN = wwn
	d.Vendor = vendor
	d.Model = model
	d.FirmwareVersion = fmver
	d.SataVersion = sataver
	d.SectorSize = sectorsize
	d.SerialNumber = sn
	d.Size = size
	d.Type = t
	d.Status = status
	d.SmartHealthStatus = smartstatus
	d.TransportProtocol = transprotocol

	return d, nil
}

func (d *DiskInfo) GetName() string {
	if d.Header != nil {
		return d.Header.GetDiskName()
	}
	return d.Name
}

func (d *DiskInfo) DomainID() string {
	return d.WWN
}

func lookupCode(t string, m map[int]string) int {
	for k, v := range m {
		if t == v {
			return k
		}
	}
	return 0
}

func diskTypeMatch(sataversion string, transportprotocol string, rotationrate string) dcaitype.DiskType {

	var disktype dcaitype.DiskType

	if strings.Contains(rotationrate, "rpm") {

		disktype = dcaitype.DiskTypeHDD
		if strings.Contains(sataversion, "SATA") {
			disktype = dcaitype.DiskTypeHDDSATA
		}
		if strings.Contains(transportprotocol, "SAS") {
			disktype = dcaitype.DiskTypeHDDSAS
		}
		return disktype
	}

	if strings.Contains(rotationrate, "Solid") {

		disktype = dcaitype.DiskTypeSSD
		if strings.Contains(sataversion, "SATA") {
			disktype = dcaitype.DiskTypeSSDSATA
		}
		if strings.Contains(transportprotocol, "SAS") {
			disktype = dcaitype.DiskTypeSSDSAS
		}
		return disktype
	}

	disktype = dcaitype.DiskTypeUnknown

	return disktype
}

type DiskHeaderType struct {
	Devname string
	Devpath string
	Devtype string
}

func NewDiskHeaderType(path string, devtype string) *DiskHeaderType {

	dh := new(DiskHeaderType)
	dh.Devpath = path
	dh.Devtype = devtype
	path_nodes := strings.Split(string(path), "/")
	dh.Devname = path_nodes[len(path_nodes)-1]
	return dh
}

func NewDiskHeaderFromSmartctlScan(scanresult string) *DiskHeaderType {
	diskHeaderStr := strings.Split(scanresult, "#")
	if diskHeaderStr[0] != "" {
		dev := strings.Split(diskHeaderStr[0], "-d")
		if len(dev) > 1 {
			return NewDiskHeaderType(strings.TrimSpace(dev[0]), strings.TrimSpace(dev[1]))
		} else {
			return NewDiskHeaderType(strings.TrimSpace(dev[0]), "scsi")
		}
	}
	return nil
}

func (dh *DiskHeaderType) NewLocalDiskInfoBySmartctl(smartctlPath string) (*DiskInfo, error) {
	if _, err := checkCmdPermission("smartctl"); err != nil {
		return nil, err
	}

	args := []string{"--xall", "--format=old", "-n", "never"}
	args = append(args, dh.Devpath)
	args = append(args, "-d")
	args = append(args, dh.Devtype)
	disktxt, _ := sudoExeccmd(smartctlPath, args...)

	// try to parse the smartctl output even the smartctl output is not complete
	return NewDiskInfoBySmartctlOutput(dh, string(disktxt))
}

func (dh *DiskHeaderType) GetDiskName() string {
	megaraid := megaraidDevice.FindStringSubmatch(dh.Devtype)
	if len(megaraid) > 1 {
		return "MegaraidDisk-" + megaraid[1]
	}

	ccissdisk := ccissDevice.FindStringSubmatch(dh.Devtype)
	if len(ccissdisk) > 1 {
		return "HP-" + ccissdisk[1]
	}

	arecadisk := arecaDevice.FindStringSubmatch(dh.Devtype)
	if len(arecadisk) > 2 {
		return "Areca-" + arecadisk[1] + "-" + arecadisk[2]
	}

	return dh.Devname
}

func NewDiskInfoBySmartctlOutput(dh *DiskHeaderType, smartctlOutput string) (*DiskInfo, error) {
	var rotationrate string

	disk := new(DiskInfo)

	disk.Header = dh
	disk.Name = dh.GetDiskName()
	disk.SmartctlOutput = smartctlOutput

	m := []*util.FindRegexpMatchAndSetType{
		{&disk.FirmwareVersion, firmwareVersion},
		{&disk.FirmwareVersion, revision},
		{&disk.Model, modelInInfo},
		{&disk.Model, product},
		{&disk.SataVersion, sataVersion},
		{&disk.SectorSize, sectorSizes},
		{&disk.SectorSize, logicalBlockSize},
		{&rotationrate, rotationRate},
		{&disk.WWN, wwnInInfo},
		{&disk.WWN, logicalUnitId},
		{&disk.SerialNumber, serialInInfo},
		{&disk.Size, userCapacity},
		{&disk.SmartHealthStatus, smartOverallHealth},
		{&disk.SmartHealthStatus, smartHealthStatus},
		{&disk.Vendor, modelFamily},
		{&disk.Vendor, vendor},
		{&disk.TransportProtocol, transportProtocol},
		{&disk.PeripheralDevtype, peripheralDeviceType},
	}
	util.FindRegexpMatchAndSet(smartctlOutput, m)

	if codeSATA := lookupCode(disk.SmartHealthStatus, smartHealthStatusTypesSATA); codeSATA != 0 {
		disk.Status = dcaitype.DiskStatusType(codeSATA)
	} else {
		codeSAS := lookupCode(disk.SmartHealthStatus, smartHealthStatusTypesSAS)
		disk.Status = dcaitype.DiskStatusType(codeSAS)
	}

	if strings.Contains(disk.WWN, " ") {
		disk.WWN = strings.Replace(disk.WWN, " ", "", -1)
	}

	disk.Type = diskTypeMatch(disk.SataVersion, disk.TransportProtocol, rotationrate)

	return disk, nil
}

func CollectSaiDiskBySmartctlOutput(acc telegraf.Accumulator, saiClusterDomainId string, hostDomainId string, dh *DiskHeaderType, smartctlOutput string) error {
	d, err := NewDiskInfoBySmartctlOutput(dh, smartctlOutput)
	if err != nil {
		return err
	}
	CreateSaiDiskDataPoint(acc, d.GetName(), saiClusterDomainId, d.Status, d.Type, d.FirmwareVersion, hostDomainId, d.Model, d.SataVersion, d.SectorSize, d.SerialNumber, d.Size, d.SmartHealthStatus, d.TransportProtocol, d.Vendor, d.WWN)
	return nil
}

func CollectSmartMetricsBySmartctlOutput(acc telegraf.Accumulator, saiClusterDomainId string, hostDomainId string, dh *DiskHeaderType, smartctlOutput string) error {

	//Start to collect data to form sai_disk_smart from telegraf
	saidisksmarttags := map[string]string{}
	saidisksmartfields := make(map[string]interface{})

	saidisksmarttags["disk_name"] = dh.GetDiskName()

	saidisksmartfields["host_domain_id"] = hostDomainId
	//saidisksmartfields["exit_status"] = exitStatus
	saidisksmartfields["cluster_domain_id"] = saiClusterDomainId

	for _, line := range strings.Split(smartctlOutput, "\n") {
		wwn := wwnInInfo.FindStringSubmatch(line)
		if len(wwn) > 1 {
			saidisksmarttags["disk_wwn"] = strings.Replace(wwn[1], " ", "", -1)
		}

		wwnSAS := logicalUnitId.FindStringSubmatch(line)
		if len(wwnSAS) > 1 {
			saidisksmarttags["disk_wwn"] = wwnSAS[1]
		}

		saidisksmarttags["primary_key"] = saiClusterDomainId + "-" + hostDomainId + "-" + saidisksmarttags["disk_wwn"]
		saidisksmarttags["disk_domain_id"] = saidisksmarttags["disk_wwn"]

		currenttemperature := currentDriveTemperature.FindStringSubmatch(line)
		if len(currenttemperature) > 1 {
			if i, err := strconv.ParseInt(currenttemperature[1], 10, 64); err == nil {
				saidisksmartfields["CurrentDriveTemperature_raw"] = i
			}
		}

		triptemperature := driveTripTemperature.FindStringSubmatch(line)
		if len(triptemperature) > 1 {
			if i, err := strconv.ParseInt(triptemperature[1], 10, 64); err == nil {
				saidisksmartfields["DriveTripTemperature_raw"] = i
			}
		}

		elementsinlist := elementsInGrownDefectList.FindStringSubmatch(line)
		if len(elementsinlist) > 1 {
			if i, err := strconv.ParseInt(elementsinlist[1], 10, 64); err == nil {
				saidisksmartfields["ElementsInGrownDefectList_raw"] = i
			}
		}

		for _, p := range powerOnHoursPatterns {
			powerOnHoursSAS := p.FindStringSubmatch(line)
			if len(powerOnHoursSAS) > 1 {
				if i, err := strconv.ParseInt(powerOnHoursSAS[1], 10, 64); err == nil {
					saidisksmartfields["9_raw"] = i
					continue
				}
			}
		}
		/*		powerOnHoursSAS := powerOnHours.FindStringSubmatch(line)
				if len(powerOnHoursSAS) > 1 {
					if i, err := strconv.ParseInt(powerOnHoursSAS[1], 10, 64); err == nil {
						saidisksmartfields["9_raw"] = i
					}
				}
		*/
		attr := attribute.FindStringSubmatch(line)

		if len(attr) > 1 {
			if val, err := parseRawValue(attr[2]); err == nil {
				saidisksmartfields[attr[1]+"_raw"] = val
			}
		}

		errorcounterlog := errorCounterLog.FindStringSubmatch(line)

		if len(errorcounterlog) > 1 {

			if errorcounterlog[1] == "read" {

				if i, err := strconv.ParseInt(errorcounterlog[2], 10, 64); err == nil {
					saidisksmartfields["ErrorsCorrectedbyECCFastRead_raw"] = i
				}
				if i, err := strconv.ParseInt(errorcounterlog[3], 10, 64); err == nil {
					saidisksmartfields["ErrorsCorrectedbyECCDelayedRead_raw"] = i
				}
				if i, err := strconv.ParseInt(errorcounterlog[4], 10, 64); err == nil {
					saidisksmartfields["ErrorCorrectedByRereadsRewritesRead_raw"] = i
				}
				if i, err := strconv.ParseInt(errorcounterlog[5], 10, 64); err == nil {
					saidisksmartfields["TotalErrorsCorrectedRead_raw"] = i
				}
				if i, err := strconv.ParseInt(errorcounterlog[6], 10, 64); err == nil {
					saidisksmartfields["CorrectionAlgorithmInvocationsRead_raw"] = i
				}
				if i, err := strconv.ParseFloat(errorcounterlog[7], 64); err == nil {
					saidisksmartfields["GigaBytesProcessedRead_raw"] = i
				}
				if i, err := strconv.ParseInt(errorcounterlog[8], 10, 64); err == nil {
					saidisksmartfields["TotalUncorrectedErrorsRead_raw"] = i
				}
			}

			if errorcounterlog[1] == "write" {

				if i, err := strconv.ParseInt(errorcounterlog[2], 10, 64); err == nil {
					saidisksmartfields["ErrorsCorrectedbyECCFastWrite_raw"] = i
				}
				if i, err := strconv.ParseInt(errorcounterlog[3], 10, 64); err == nil {
					saidisksmartfields["ErrorsCorrectedbyECCDelayedWrite_raw"] = i
				}
				if i, err := strconv.ParseInt(errorcounterlog[4], 10, 64); err == nil {
					saidisksmartfields["ErrorCorrectedByRereadsRewritesWrite_raw"] = i
				}
				if i, err := strconv.ParseInt(errorcounterlog[5], 10, 64); err == nil {
					saidisksmartfields["TotalErrorsCorrectedWrite_raw"] = i
				}
				if i, err := strconv.ParseInt(errorcounterlog[6], 10, 64); err == nil {
					saidisksmartfields["CorrectionAlgorithmInvocationsWrite_raw"] = i
				}
				if i, err := strconv.ParseFloat(errorcounterlog[7], 64); err == nil {
					saidisksmartfields["GigaBytesProcessedWrite_raw"] = i
				}
				if i, err := strconv.ParseInt(errorcounterlog[8], 10, 64); err == nil {
					saidisksmartfields["TotalUncorrectedErrorsWrite_raw"] = i
				}
			}
		}
	}

	acc.AddFields("sai_disk_smart", saidisksmartfields, saidisksmarttags)

	return nil
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

func getLocalDiskListByDevNode(smartctlPath string, devnodePrefix string) ([]*DiskHeaderType, error) {
	var disklist []*DiskHeaderType

	out, err := execcmd("ls", fmt.Sprintf("/dev/%s*", devnodePrefix))
	if err != nil {
		return nil, err
	}
	for _, dev := range strings.Split(string(out), "\n") {
		if len(dev) == 0 {
			continue
		}
		dh := NewDiskHeaderType(strings.TrimSpace(dev), "scsi")
		d, _ := dh.NewLocalDiskInfoBySmartctl(smartctlPath)
		if d == nil {
			continue
		}

		// continue to probe possible disks event d's information is not complete due to smartctl err
		switch d.PeripheralDevtype {
		case "storage array":
			if d.Vendor == "HP" {
				disklist = append(disklist, generateDisklistByDevtype(dev, "cciss")...)
			}
		case "disk":
			disklist = append(disklist, dh)
		default:
			if d.Vendor == "Areca" {
				disklist = append(disklist, generateDisklistByDevtype(dev, "areca")...)
			}
		}
	}
	return disklist, nil
}

func generateDisklistByDevtype(devpath string, devtype string) []*DiskHeaderType {
	var disklist []*DiskHeaderType

	if devtype == "cciss" {
		for i := 0; i < CCISS_MAX_SLOT_NUM; i++ {
			dh := NewDiskHeaderType(devpath, fmt.Sprintf("cciss,%d", i))
			disklist = append(disklist, dh)
		}
	} else if devtype == "areca" {
		for i := 1; i <= ARECA_MAX_ENCLOSURE_NUM; i++ {
			for j := 1; j <= ARECA_MAX_SLOT_NUM; j++ {
				dh := NewDiskHeaderType(devpath, fmt.Sprintf("areca,%d/%d", j, i))
				disklist = append(disklist, dh)
			}
		}
	}
	return disklist
}

func getLocalDiskListBySmartctlScan(smartctlPath string) ([]*DiskHeaderType, error) {
	var (
		out      []byte
		err      error
		disklist []*DiskHeaderType
	)

	if _, err := checkCmdPermission("smartctl"); err != nil {
		return nil, err
	}

	if out, err = sudoExeccmd(smartctlPath, "--scan-open"); err != nil {
		return nil, err
	}

	for _, line := range strings.Split(string(out), "\n") {
		dh := NewDiskHeaderFromSmartctlScan(line)
		if dh != nil {
			disklist = append(disklist, dh)
		}
	}
	return disklist, nil

}

func isSATADisk(t dcaitype.DiskType) bool {
	if t == dcaitype.DiskTypeSSDSATA ||
		t == dcaitype.DiskTypeHDDSATA {
		return true
	}
	return false
}

func isSASDisk(t dcaitype.DiskType) bool {
	if t == dcaitype.DiskTypeSSDSAS ||
		t == dcaitype.DiskTypeHDDSAS {
		return true
	}
	return false
}

func GetLocalDisks(smartctlPath string) ([]*DiskInfo, error) {
	var (
		disks          []*DiskInfo
		diskMapBySN    map[string]*DiskInfo
		diskHeaderList []*DiskHeaderType
		list           []*DiskHeaderType
		d              *DiskInfo
		err            error
	)

	if err = util.CheckCmdPath(smartctlPath); err != nil {
		return nil, err
	}

	list, err = getLocalDiskListBySmartctlScan(smartctlPath)
	if err == nil {
		diskHeaderList = append(diskHeaderList, list...)
	}

	list, err = getLocalDiskListByDevNode(smartctlPath, "sg")
	if err == nil {
		diskHeaderList = append(diskHeaderList, list...)
	}

	// in case that sdx and sgx are the same disk, filter out disks with the same WWN
	// smartctl --info will not have WWN. So use sn as the key
	diskMapBySN = make(map[string]*DiskInfo)

	for _, diskHeader := range diskHeaderList {
		d, err = diskHeader.NewLocalDiskInfoBySmartctl(smartctlPath)
		if err == nil {
			if !IsValidDisk(d) {
				continue
			}
			_, snExisted := diskMapBySN[d.WWN]
			if !snExisted {
				diskMapBySN[d.WWN] = d
				disks = append(disks, d)
			}
		}
	}

	return disks, nil

}

func IsValidDisk(d *DiskInfo) bool {
	if d.WWN != "" && (isSASDisk(d.Type) || isSATADisk(d.Type)) {
		return true
	}
	return false
}

// CreateSaiDiskDataPoint create a data point of sai_disk
func CreateSaiDiskDataPoint(
	acc telegraf.Accumulator,
	diskName string,
	saiClusterDomainId string,
	diskStatus dcaitype.DiskStatusType,
	diskType dcaitype.DiskType,
	firmwareVersion string,
	hostDomainID string,
	model string,
	sataVersion string,
	sectorSize string,
	SerialNumber string,
	size string,
	smartHealthStatus string,
	transportProtocol string,
	vendor string,
	diskWWN string,
) {
	diskTags := map[string]string{}
	diskFields := make(map[string]interface{})
	diskTags["disk_name"] = diskName
	diskTags["disk_domain_id"] = diskWWN
	diskTags["primary_key"] = saiClusterDomainId + "-" + hostDomainID + "-" + diskTags["disk_domain_id"]
	diskTags["disk_wwn"] = diskWWN
	diskFields["cluster_domain_id"] = saiClusterDomainId
	diskFields["disk_status"] = int(diskStatus)
	diskFields["disk_type"] = int(diskType)
	diskFields["firmware_version"] = firmwareVersion
	diskFields["host_domain_id"] = hostDomainID
	diskFields["model"] = model
	diskFields["sata_version"] = sataVersion
	diskFields["sector_size"] = sectorSize
	diskFields["serial_number"] = SerialNumber
	diskFields["size"] = size
	diskFields["smart_health_status"] = smartHealthStatus
	diskFields["transport_protocol"] = transportProtocol
	diskFields["vendor"] = vendor
	acc.AddFields("sai_disk", diskFields, diskTags)
}
