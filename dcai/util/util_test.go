package util

import (
	"strings"
	"testing"
)

func TestExecuteCmdWithTimeout(t *testing.T) {
	cmd := "sleep"
	timeoutvalue := "1"
	_, err := ExecuteCmdWithTimeout(cmd, timeoutvalue)
	if err != nil {
		t.Errorf("Command does not execeed timeout value (%v)", cmdTimeoutSecond)
	}
	timeoutvalue = "6"
	_, err = ExecuteCmdWithTimeout(cmd, timeoutvalue)
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Command '%s %s' should got timeout (%v)", cmd, timeoutvalue, cmdTimeoutSecond)
	}
}

func TestExecuteSudoCmdWithTimeout(t *testing.T) {
	cmd := "sleep"
	timeoutvalue := "1"
	_, err := ExecuteCmdWithTimeout(cmd, timeoutvalue)
	if err != nil {
		t.Errorf("Command does not execeed timeout value (%v)", cmdTimeoutSecond)
	}
	timeoutvalue = "6"
	_, err = ExecuteCmdWithTimeout(cmd, timeoutvalue)
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Command '%s %s' should got timeout (%v)", cmd, timeoutvalue, cmdTimeoutSecond)
	}
}

func TestCommandExist(t *testing.T) {
	_, existed := CommandExist("no_this_command")
	if existed != false {
		t.Errorf("Command no_this_command should not existed")
	}
}
