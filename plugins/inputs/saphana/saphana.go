package saphana

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/SAP/go-hdb/driver"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type SapHana struct {
	Server   string `toml:"server"`
	Port     string `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// M_Database_View holds the columns of the M_Database view
type M_Database_View struct {
	System_id     string
	Database_name string
	Host          string
	Start_time    string
	Version       string
	Usage         string
}

type MHostResource struct {
	Host                              string
	FreePhysicalMemory                int64
	UsedPhysicalMemory                int64
	FreeSwapSpace                     int64
	UsedSwapSpace                     int64
	AllocationLimit                   int64
	InstanceTotalMemeoryUsedSize      int64
	InstanceTotalMemeoryPeakUsedSize  int64
	InstanceTotalMemeoryAllocatedSize int64
	InstanceCodeSize                  int64
	InstanceSharedMemoryAllocatedSize int64
	TotalCPUUserTime                  int64
	TotalCPUSystemTime                int64
	TotalCPUWioTime                   int64
	TotalCPUIdleTime                  int64
	UtcTimestamp                      time.Time
}

type MHostNetwork struct {
	Host                       string
	TCP_SEGMENTS_RECEIVED      int64
	TCP_SEGMENTS_SENT_OUT      int64
	TCP_SEGMENTS_RETRANSMITTED int64
	TCP_BAD_SEGMENTS_RECEIVED  int64
}

type MDisk struct {
	DISK_ID         int
	DEVICE_ID       int64
	Host            string
	PATH            string
	SUBPATH         string
	FILESYSTEM_TYPE string
	USAGE_TYPE      string
	TOTAL_SIZE      int64
	USED_SIZE       int64
}

type MServiceNetworkIO struct {
	SENDER_HOST      string
	SENDER_PORT      int
	RECEIVER_HOST    string
	RECEIVER_PORT    int
	SEND_SIZE        int64
	RECEIVE_SIZE     int64
	SEND_DURATION    int64
	RECEIVE_DURATION int64
	REQUEST_COUNT    int64
}

var sampleConfig = `
  ## Sap Hana host
	server = "saphana.Ip"
	port = "39013"
  username = "root"
  password = "saphana"
`

func (s *SapHana) Description() string {
	return "Collect metrics from Sap Hana"
}

func (s *SapHana) SampleConfig() string {
	return sampleConfig
}

func (s *SapHana) getDbUrl() string {
	return "hdb://" + s.Username + ":" + s.Password + "@" + s.Server + ":" + s.Port
}

func (s *SapHana) getHostResource(acc telegraf.Accumulator, db *sql.DB) error {

	// Query the M_HOST_RESOURCE_UTILIZATION view
	rows, err := db.Query("select HOST, FREE_PHYSICAL_MEMORY, USED_PHYSICAL_MEMORY, FREE_SWAP_SPACE, USED_SWAP_SPACE, ALLOCATION_LIMIT, INSTANCE_TOTAL_MEMORY_USED_SIZE, INSTANCE_TOTAL_MEMORY_PEAK_USED_SIZE, INSTANCE_TOTAL_MEMORY_ALLOCATED_SIZE, INSTANCE_CODE_SIZE, INSTANCE_SHARED_MEMORY_ALLOCATED_SIZE, TOTAL_CPU_USER_TIME, TOTAL_CPU_SYSTEM_TIME, TOTAL_CPU_WIO_TIME, TOTAL_CPU_IDLE_TIME, UTC_TIMESTAMP from M_HOST_RESOURCE_UTILIZATION")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		mdb := new(MHostResource)
		records := make(map[string]interface{})
		tags := make(map[string]string)

		err := rows.Scan(&mdb.Host, &mdb.FreePhysicalMemory, &mdb.UsedPhysicalMemory, &mdb.FreeSwapSpace, &mdb.UsedSwapSpace, &mdb.AllocationLimit, &mdb.InstanceTotalMemeoryUsedSize, &mdb.InstanceTotalMemeoryPeakUsedSize, &mdb.InstanceTotalMemeoryAllocatedSize, &mdb.InstanceCodeSize, &mdb.InstanceSharedMemoryAllocatedSize, &mdb.TotalCPUUserTime, &mdb.TotalCPUSystemTime, &mdb.TotalCPUWioTime, &mdb.TotalCPUIdleTime, &mdb.UtcTimestamp)

		if err != nil {
			return err
		}
		tags["host"] = mdb.Host

		records["free_physical_memory"] = mdb.FreePhysicalMemory
		records["used_physical_memory"] = mdb.UsedPhysicalMemory
		records["free_swap_space"] = mdb.FreeSwapSpace
		records["used_swap_space"] = mdb.UsedSwapSpace
		records["allocation_limit"] = mdb.AllocationLimit
		records["instance_total_memeory_used_size"] = mdb.InstanceTotalMemeoryUsedSize
		records["instance_total_memeory_peak_used_size"] = mdb.InstanceTotalMemeoryPeakUsedSize
		records["instance_total_memeory_allocated_size"] = mdb.InstanceTotalMemeoryAllocatedSize
		records["instance_code_size"] = mdb.InstanceCodeSize
		records["instance_shared_memory_allocated_size"] = mdb.InstanceSharedMemoryAllocatedSize
		records["total_cpu_user_time"] = mdb.TotalCPUUserTime
		records["total_cpu_system_time"] = mdb.TotalCPUSystemTime
		records["total_cpu_wio_time"] = mdb.TotalCPUWioTime
		records["total_cpu_idle_time"] = mdb.TotalCPUIdleTime

		acc.AddFields("sap_host_resource", records, tags)
	}

	return nil
}

func (s *SapHana) getHostNetwork(acc telegraf.Accumulator, db *sql.DB) error {

	// Query the M_HOST_NETWORK_STATISTICS view
	rows, err := db.Query("select HOST, TCP_SEGMENTS_RECEIVED, TCP_SEGMENTS_SENT_OUT, TCP_SEGMENTS_RETRANSMITTED, TCP_BAD_SEGMENTS_RECEIVED from M_HOST_NETWORK_STATISTICS")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		mdb := new(MHostNetwork)
		records := make(map[string]interface{})
		tags := make(map[string]string)

		err := rows.Scan(&mdb.Host, &mdb.TCP_SEGMENTS_RECEIVED, &mdb.TCP_SEGMENTS_SENT_OUT, &mdb.TCP_SEGMENTS_RETRANSMITTED, &mdb.TCP_BAD_SEGMENTS_RECEIVED)

		if err != nil {
			return err
		}
		tags["host"] = mdb.Host

		records["tcp_segments_received"] = mdb.TCP_SEGMENTS_RECEIVED
		records["tcp_segments_sent_out"] = mdb.TCP_SEGMENTS_SENT_OUT
		records["tcp_segments_retransmitted"] = mdb.TCP_SEGMENTS_RETRANSMITTED
		records["tcp_segments_received"] = mdb.TCP_BAD_SEGMENTS_RECEIVED

		acc.AddFields("sap_host_network", records, tags)
	}
	return nil
}

func (s *SapHana) getServiceNetworkIO(acc telegraf.Accumulator, db *sql.DB) error {

	// Query the M_HOST_NETWORK_STATISTICS view
	rows, err := db.Query("select SENDER_HOST, SENDER_PORT, RECEIVER_HOST, RECEIVER_PORT, SEND_SIZE, RECEIVE_SIZE, SEND_DURATION, RECEIVE_DURATION, REQUEST_COUNT from M_SERVICE_NETWORK_IO")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		mdb := new(MServiceNetworkIO)
		records := make(map[string]interface{})
		tags := make(map[string]string)

		err := rows.Scan(&mdb.SENDER_HOST, &mdb.SENDER_PORT, &mdb.RECEIVER_HOST, &mdb.RECEIVER_PORT, &mdb.SEND_SIZE, &mdb.RECEIVE_SIZE, &mdb.SEND_DURATION, &mdb.RECEIVE_DURATION, &mdb.REQUEST_COUNT)

		if err != nil {
			return err
		}
		tags["host"] = mdb.SENDER_HOST

		records["sender_host"] = mdb.SENDER_HOST
		records["sender_port"] = mdb.SENDER_PORT
		records["receiver_host"] = mdb.RECEIVER_HOST
		records["receiver_port"] = mdb.RECEIVER_PORT
		records["sender_size"] = mdb.SEND_SIZE
		records["receiver_size"] = mdb.RECEIVE_SIZE
		records["sender_duration"] = mdb.SEND_DURATION
		records["receiver_duration"] = mdb.RECEIVE_DURATION
		records["request_count"] = mdb.REQUEST_COUNT

		acc.AddFields("sap_service_network", records, tags)
	}
	return nil
}

func (s *SapHana) getDisk(acc telegraf.Accumulator, db *sql.DB) error {

	// Query the M_DISKS view
	rows, err := db.Query("select DISK_ID, DEVICE_ID, HOST, PATH, SUBPATH, FILESYSTEM_TYPE, USAGE_TYPE, TOTAL_SIZE, USED_SIZE from M_DISKS")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		mdb := new(MDisk)
		records := make(map[string]interface{})
		tags := make(map[string]string)

		err := rows.Scan(&mdb.DISK_ID, &mdb.DEVICE_ID, &mdb.Host, &mdb.PATH, &mdb.SUBPATH, &mdb.FILESYSTEM_TYPE, &mdb.USAGE_TYPE, &mdb.TOTAL_SIZE, &mdb.USED_SIZE)

		if err != nil {
			return err
		}
		tags["host"] = mdb.Host

		records["disk_id"] = mdb.DISK_ID
		records["device_id"] = mdb.DEVICE_ID
		records["path"] = mdb.PATH
		records["subpath"] = mdb.SUBPATH
		records["filesystem_type"] = mdb.FILESYSTEM_TYPE
		records["usage_type"] = mdb.USAGE_TYPE
		records["total_size"] = mdb.TOTAL_SIZE
		records["used_size"] = mdb.USED_SIZE

		acc.AddFields("sap_disk", records, tags)

	}
	return nil
}

func (s *SapHana) Gather(acc telegraf.Accumulator) error {
	db, err := sql.Open("hdb", s.getDbUrl())
	if err != nil {
		acc.AddError(fmt.Errorf("Failed to connect to the database: %s", err))
	}
	defer db.Close()

	err = s.getHostResource(acc, db)
	if err != nil {
		acc.AddError(fmt.Errorf("Cannot read Sap Hana host resources for: %s", err))
	}

	err = s.getServiceNetworkIO(acc, db)
	if err != nil {
		acc.AddError(fmt.Errorf("Cannot read Sap Hana service network for: %s", err))
	}

	err = s.getDisk(acc, db)
	if err != nil {
		acc.AddError(fmt.Errorf("Cannot read Sap Hana disk for: %s", err))
	}

	return nil
}

func init() {
	inputs.Add("saphana", func() telegraf.Input { return &SapHana{} })
}
