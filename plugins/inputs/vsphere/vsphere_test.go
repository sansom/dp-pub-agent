package vsphere

import (
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnection(t *testing.T) {
	if os.Getenv("CIRCLE_PROJECT_REPONAME") != "localTest" {
		t.Skip("Skipping CI testing due to race conditions")
	}
	var acc testutil.Accumulator

	v := &VSphere{
		Server:   "127.0.0.1",
		Username: "root",
		Password: "password",
		Insecure: true,
	}
	err := v.Gather(&acc)
	if err != nil {
		require.NoError(t, err)
	}

}
