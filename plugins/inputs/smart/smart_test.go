package smart

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/influxdata/telegraf/dcai/hardware/disk"
	"github.com/influxdata/telegraf/dcai/topology/cluster"
	"github.com/influxdata/telegraf/dcai/topology/host"
	"github.com/influxdata/telegraf/dcai/topology/host/linux"
	"github.com/influxdata/telegraf/dcai/type"
	"github.com/influxdata/telegraf/testutil"
)

const (
	notNull         = "NOTNULL"
	saiDiskSmart    = "sai_disk_smart"
	primaryKey      = "primary_key"
	clusterDomainID = "cluster_domain_id"
	hostDomainID    = "host_domain_id"
	diskDomainID    = "disk_domain_id"
)

var (
	mockHost = &linux.LinuxHostConfig{
		Name:   "testHost",
		OSType: dcaitype.OSLinux,
	}
	mockCluster = &cluster.ClusterConfig{
		Name:        "testCluster",
		ClusterType: dcaitype.ClusterDefaultCluster,
		Hosts:       []host.HostConfig{mockHost},
	}
	mockScanData          string
	mockInfoAttributeData string
	// int64(2091)
	int64Number = regexp.MustCompile("^int64\\((.*)\\)$")
	// float64(3640.962)
	float64Number = regexp.MustCompile("^float64\\((.*)\\)$")
	dirPath       = "./diskmockData"
	filePath      string
)

type ExpectedOutputDataStruct struct {
	Fields map[string]interface{} `json:"fields"`
	Tags   map[string]string      `json:"tags"`
}

func TestGatherAttributes(t *testing.T) {

	fileList, err := getFileList(dirPath)
	if err != nil {
		t.Errorf("%s\n", err)
		return
	}

	for _, filePathLocal := range fileList {

		checkSmartctlPermission = fakeCheckCommandPermission
		s := &Smart{
			Path:       "smartctl",
			Attributes: true,
		}

		var acc testutil.Accumulator
		var expectedOutputData *ExpectedOutputDataStruct
		var err error
		var scanstr string
		var smartctlOut string

		filePath = filePathLocal
		scanstr, smartctlOut, expectedOutputData, err = readMockData(filePath)
		if err != nil {
			t.Errorf("%s\n", err)
			return
		}

		dh := disk.NewDiskHeaderFromSmartctlScan(scanstr)
		d, _ := disk.NewDiskInfoBySmartctlOutput(dh, smartctlOut)
		disks := []*disk.DiskInfo{d}
		mockHost.SetDisks(disks)

		s.getAttributes(&acc, "dpCluster", mockHost)

		//		require.NoError(t, err)

		typeChange(expectedOutputData.Fields)

		for key, value := range expectedOutputData.Fields {
			if value == notNull {
				if acc.HasField(saiDiskSmart, key) {
					if i, _ := acc.StringField(saiDiskSmart, key); i != "" {
						expectedOutputData.Fields[key] = i
					}
				}
			}
		}

		for key, value := range expectedOutputData.Tags {
			if value == notNull {
				if acc.HasTag(saiDiskSmart, key) {
					if i := acc.TagValue(saiDiskSmart, key); i != "" {
						expectedOutputData.Tags[key] = i
					}
				}
			}
		}

		acc.AssertContainsFields(t, saiDiskSmart, expectedOutputData.Fields)

		acc.AssertContainsTaggedFields(t, saiDiskSmart, expectedOutputData.Fields, expectedOutputData.Tags)
	}
}

func TestPrimaryKey(t *testing.T) {

	fileList, err := getFileList(dirPath)
	if err != nil {
		t.Errorf("%s\n", err)
		return
	}

	for _, filePathLocal := range fileList {

		checkSmartctlPermission = fakeCheckCommandPermission
		s := &Smart{
			Path:       "smartctl",
			Attributes: true,
		}

		var acc testutil.Accumulator
		var smartctlOut string

		filePath = filePathLocal
		scanstr, smartctlOut, _, err := readMockData(filePath)
		if err != nil {
			t.Errorf("%s\n", err)
			return
		}

		dh := disk.NewDiskHeaderFromSmartctlScan(scanstr)
		d, _ := disk.NewDiskInfoBySmartctlOutput(dh, smartctlOut)
		disks := []*disk.DiskInfo{d}
		mockHost.SetDisks(disks)

		s.getAttributes(&acc, "dpCluster", mockHost)

		primarykey := acc.TagValue(saiDiskSmart, primaryKey)
		clusterdomainid, _ := acc.StringField(saiDiskSmart, clusterDomainID)
		hostdomainid, _ := acc.StringField(saiDiskSmart, hostDomainID)
		diskdomainid := acc.TagValue(saiDiskSmart, diskDomainID)

		expectedPK := clusterdomainid + "-" + hostdomainid + "-" + diskdomainid

		if primarykey != expectedPK {
			t.Errorf("Primary key does not match the combination rule. Received %s != Expected %s", primarykey, expectedPK)
			return
		}
	}
}

func fakeCheckCommandPermission(cmd string) (bool, error) {
	return true, nil
}

func readMockData(filename string) (string, string, *ExpectedOutputDataStruct, error) {

	fileopen, err := os.Open(filename)
	if err != nil {
		return "", "", nil, err
	}

	defer func() {
		if err := fileopen.Close(); err != nil {
			panic(err)
		}
	}()

	readData := bufio.NewReader(fileopen)

	mockData, err := ioutil.ReadAll(readData)
	if err != nil {
		return "", "", nil, err
	}

	mockDataArray := bytes.Split(mockData, []byte("$~~@#!#@~~$\n"))
	//    fmt.Printf("%d\n%s\n", len(mockDataArray), mockData)

	readScanData := string(mockDataArray[1])
	readInfoAttributeData := string(mockDataArray[2])
	outExpectedData := mockDataArray[3]
	//    fmt.Printf("1 --> %s\n2 --> %s\n3 --> %s\n", readScanData, readInfoAttributeData, outExpectedData)

	var expectedOutputData *ExpectedOutputDataStruct
	err = json.Unmarshal(outExpectedData, &expectedOutputData)
	if err != nil {
		return "", "", nil, err
	}

	return readScanData, readInfoAttributeData, expectedOutputData, nil
}

func typeChange(fields map[string]interface{}) {

	for key, value := range fields {

		int64Value := int64Number.FindStringSubmatch(value.(string))
		if len(int64Value) > 1 {
			if i, err := strconv.ParseInt(int64Value[1], 10, 64); err == nil {
				fields[key] = i
			}
		}

		float64Value := float64Number.FindStringSubmatch(value.(string))
		if len(float64Value) > 1 {
			if i, err := strconv.ParseFloat(float64Value[1], 64); err == nil {
				fields[key] = i
			}
		}
	}
}

func getFileList(path string) ([]string, error) {

	var filePathArray []string
	err := filepath.Walk(

		path, func(path string, file os.FileInfo, err error) error {
			if file == nil {
				return err
			}
			if file.IsDir() {
				return nil
			}
			filePathArray = append(filePathArray, path)
			return nil
		})
	if err != nil {
		return nil, err
	}
	return filePathArray, nil
}
