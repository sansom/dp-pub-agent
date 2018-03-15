package saphana

import (
	"database/sql"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGather(t *testing.T) {
	if os.Getenv("CIRCLE_PROJECT_REPONAME") != "localTest" {
		t.Skip("Skipping CI testing due to race conditions")
	}
	acc := testutil.Accumulator{}

	sap := &SapHana{
		Server:   "127.0.0.1",
		Port:     "39013",
		Username: "root",
		Password: "password",
	}

	// err := sap.getHostResource()
	// if err != nil {
	// 	require.NoError(t, err)
	// }

	// err := sap.getHostNetwork()
	// if err != nil {
	// 	require.NoError(t, err)
	// }

	// err := sap.getDisk()
	// if err != nil {
	// 	require.NoError(t, err)
	// }

	err := sap.Gather(&acc)
	if err != nil {
		require.NoError(t, err)
	}
	assert.True(t, acc.HasMeasurement("sap_host_resource"))
	assert.True(t, acc.HasMeasurement("sap_host_network"))
	assert.True(t, acc.HasMeasurement("sap_disk"))
	assert.True(t, acc.HasField("sap_host_resource", "total_cpu_idle_time"))
	assert.True(t, acc.HasField("sap_host_network", "tcp_segments_received"))
	assert.True(t, acc.HasField("sap_disk", "used_size"))

}

func TestDBPing(t *testing.T) {
	if os.Getenv("CIRCLE_PROJECT_REPONAME") != "localTest" {
		t.Skip("Skipping CI testing due to race conditions")
	}
	sap := &SapHana{
		Server:   "127.0.0.1",
		Port:     "39013",
		Username: "root",
		Password: "password",
	}

	db, err := sql.Open("hdb", sap.getDbUrl())
	require.NoError(t, err)

	defer db.Close()

	if err := db.Ping(); err != nil {
		require.NoError(t, err)
	}
}

func TestGetDatabase(t *testing.T) {
	if os.Getenv("CIRCLE_PROJECT_REPONAME") != "localTest" {
		t.Skip("Skipping CI testing due to race conditions")
	}
	sap := &SapHana{
		Server:   "127.0.0.1",
		Port:     "39013",
		Username: "root",
		Password: "password",
	}

	db, err := sql.Open("hdb", sap.getDbUrl())
	require.NoError(t, err)

	defer db.Close()

	mdbs := make([]*M_Database_View, 0)

	// Query the M_DATABASE view
	// var ver string
	// err = db.QueryRow(`SELECT host FROM m_host_resource_utilization`).Scan(&ver)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("SAP HANA " + ver)

	rows, err := db.Query("Select * from M_DATABASE")
	require.NoError(t, err)

	defer rows.Close()

	for rows.Next() {
		mdb := new(M_Database_View)
		err := rows.Scan(&mdb.System_id,
			&mdb.Database_name,
			&mdb.Host,
			&mdb.Start_time,
			&mdb.Version,
			&mdb.Usage)
		if err != nil {
			require.NoError(t, err)
		}
		mdbs = append(mdbs, mdb)
	}
	// fmt.Println(mdbs)
}
