package util

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os/exec"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/influxdata/telegraf/internal"
)

const (
	cmdTimeoutSecond    = 5 * time.Second
	sshCmdTimeoutSecond = 10 * time.Second
)

type FindRegexpMatchAndSetType struct {
	TargetToSet *string
	Regexp      *regexp.Regexp
}

func CommandExist(cmd ...string) (string, bool) {
	for _, v := range cmd {
		path, _ := exec.LookPath(v)
		if len(path) <= 0 {
			return v, false
		}
	}
	return "", true
}

// Calculate the checksum for a given key
func GenHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func isRoot() bool {
	u, err := user.Current()
	if err != nil {
		return false
	}
	return (u.Uid == "0")
}

func CheckCmdRootPermission(cmd string) (bool, error) {
	if isRoot() {
		return true, nil
	}
	out, err := executecmdwithtimeoutInternal(true, cmdTimeoutSecond, "sudo", "-l")
	if err != nil {
		return false, fmt.Errorf("Cannot check command root permission for %s. %s(%s)", cmd, err.Error(), string(out))
	}
	if strings.Contains(string(out), cmd) {
		return true, nil
	} else {
		if strings.Contains(string(out), "ALL") {
			return true, nil
		} else {
			return false, fmt.Errorf("Cannot have permission to execute %s", cmd)
		}
	}
}

func ExecuteSshCmdwithtimeout(addr string, user string, pw string, cmdStr string) ([]byte, error) {
	args := append([]string{"-p", pw, "ssh", "-oStrictHostKeyChecking=no", fmt.Sprintf("%s@%s", user, addr), cmdStr, "2>/dev/null"})
	e := exec.Command("sshpass", args...)
	out, err := internal.CombinedOutputTimeout(e, sshCmdTimeoutSecond)
	if err != nil {
		return out, fmt.Errorf("%s got %s", cmdStr, err.Error())
	}

	return out, nil
}

func executecmdwithtimeoutInternal(withsudo bool, timeout time.Duration, cmd string, args ...string) ([]byte, error) {
	if c, found := CommandExist(cmd); !found {
		err := fmt.Errorf("Cannot find command %s", c)
		return nil, err
	}
	str := cmd + " " + strings.Join(args, " ")
	if withsudo && !isRoot() {
		str = "sudo " + str
	}
	e := exec.Command("bash", "-c", str)
	out, err := internal.CombinedOutputTimeout(e, timeout)
	if err != nil {
		return out, fmt.Errorf("%s got %s", str, err.Error())
	}

	return out, nil
}

func ExecuteSudoCmdWithTimeout(cmd string, args ...string) ([]byte, error) {
	return executecmdwithtimeoutInternal(true, cmdTimeoutSecond, cmd, args...)
}

func ExecuteCmdWithTimeout(cmd string, args ...string) ([]byte, error) {
	return executecmdwithtimeoutInternal(false, cmdTimeoutSecond, cmd, args...)
}

func FindRegexpMatchAndSet(txt string, pattern []*FindRegexpMatchAndSetType) {
	// init target to empty
	for _, p := range pattern {
		*p.TargetToSet = ""
	}

	// find matched pattern
	for _, line := range strings.Split(txt, "\n") {
		for _, p := range pattern {
			if f := p.Regexp.FindStringSubmatch(line); len(f) > 1 {
				*p.TargetToSet = f[1]
			}
		}
	}
}

func FindRegexpMatchAndAppend(line string, target *[]string, pattern *regexp.Regexp) {
	strArray := pattern.FindStringSubmatch(line)
	if len(strArray) > 1 {
		if target != nil {
			*target = append(*target, strArray[1])
		}
	}
}

func GetCmdPathInOsPath(cmd string) (string, error) {
	path, _ := exec.LookPath(cmd)
	if len(path) == 0 {
		return "", fmt.Errorf("Cannot find %s in PATH", cmd)
	}
	return path, nil
}

func CheckCmdPath(cmdPath string) error {
	if len(cmdPath) > 0 {
		// test if smartctl is valid
		_, err := ExecuteCmdWithTimeout("ls", cmdPath)
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("Path is not specified")
}

func FindIndexOfText(txt []byte, seps [][]byte) int {
	var min = len(txt)
	var hit = false
	for _, sep := range seps {
		index := bytes.Index(txt, sep)
		if index == -1 {
			continue
		}
		hit = true
		if index < min {
			min = index
		}
	}
	if !hit {
		return -1
	}
	return min
}
