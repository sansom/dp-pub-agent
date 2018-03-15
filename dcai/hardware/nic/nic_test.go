package nic

import (
	"github.com/influxdata/telegraf/dcai/testutil"
	"testing"
)

var (
	input = `
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
2: eno1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP qlen 1000
    link/ether 00:1d:09:6b:b3:8e brd ff:ff:ff:ff:ff:ff
    inet 10.1.10.133/24 brd 10.1.10.255 scope global eno1
       valid_lft forever preferred_lft forever
    inet6 fe80::efdb:b26d:714a:55c/64 scope link
       valid_lft forever preferred_lft forever
3: eno2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP qlen 1000
    link/ether 00:1d:09:6b:b3:90 brd ff:ff:ff:ff:ff:ff
`

	nics = []*NetworkInfo{
		&NetworkInfo{"eno1", []string{"00:1d:09:6b:b3:8e"}, []string{"10.1.10.133"}, []string{"fe80::efdb:b26d:714a:55c"}},
		&NetworkInfo{"eno2", []string{"00:1d:09:6b:b3:90"}, []string{}, []string{}},
	}
)

func fakeExecCommand(cmd string, args ...string) ([]byte, error) {
	return []byte(input), nil
}

func TestNewAllNetworkInfo(t *testing.T) {
	execcmd = fakeExecCommand
	funcout, err := NewAllNetworkInfo()
	if err != nil {
		t.Errorf("NewAllNetworkInfo return error (%s)", err)
	}

	testutil.CompareVar(t, funcout, nics)
}
