package ProxmoxCLI

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"

	gv "github.com/hashicorp/go-version"
)

type MachineType string
type MachineStatus string
type BackupMode string
type BackupCompression string

type BackupAcceptor func(*io.PipeReader) error

const (
	VM  MachineType = "qemu"
	LXC MachineType = "lxc"
	// OpenVZ  MachineType = "openvz"

	Running MachineStatus = "running"
	Stopped MachineStatus = "stopped"

	DontCompress BackupCompression = "0"
	GZipCompress BackupCompression = "gzip"
	LzoCompress  BackupCompression = "lzo"
	ZStdCompress BackupCompression = "zstd"

	Snapshot BackupMode = "snapshot"
	Stop     BackupMode = "stop"
	Suspend  BackupMode = "suspend"
)

type MachineInfo struct {
	ID     string        `json:"id"`
	Type   MachineType   `json:"type"`
	VMID   uint64        `json:"vmid,omitempty"`
	Name   string        `json:"name,omitempty"`
	Node   string        `json:"node,omitempty"`
	Status MachineStatus `json:"status,omitempty"`
}

type GetMachinesByPoolInfo struct {
	Comment string        `json:"comment"`
	PoolID  string        `json:"poolid"`
	Members []MachineInfo `json:"members"`
}

type StartImageBackupSettings struct {
	Compression    BackupCompression
	Mode           BackupMode
	AdditionalArgs []string
}

func GetMachinesByPool(Pool string) ([]MachineInfo, error) {
	cmd := exec.Command("pvesh", "get", "/pools/"+Pool, "--output-format=json")
	if output, err := cmd.Output(); err != nil {
		return nil, fmt.Errorf("pvesh returned an error: %w", err)
	} else {
		data := GetMachinesByPoolInfo{}
		if err = json.Unmarshal(output, &data); err != nil {
			return nil, fmt.Errorf("json decoding of pvesh data returned an error: %w", err)
		} else {
			return data.Members, nil
		}
	}
}

var regexPveVersion *regexp.Regexp

func GetVersion() (*gv.Version, error) {
	cmd := exec.Command("pveversion")
	if output, err := cmd.Output(); err != nil {
		return nil, fmt.Errorf("pveversion returned an error: %w", err)
	} else {
		if regexPveVersion == nil {
			regexPveVersion = regexp.MustCompile(`pve.*\/([\d\.]+)\/.*`)
		}
		submatches := regexPveVersion.FindStringSubmatch(string(output))
		if submatches == nil || len(submatches) != 2 {
			return nil, fmt.Errorf("pveversion returned a non-parsable string: %v", string(output))
		}

		if ver, err := gv.NewVersion(submatches[1]); err != nil {
			return nil, fmt.Errorf("couldn't parse pve version number: %w", err)
		} else {
			return ver, nil
		}
	}
}

func StartImageBackup(VMID uint64, Settings StartImageBackupSettings) (*exec.Cmd, error) {
	args := []string{
		strconv.FormatUint(VMID, 10),
		"--stdout",
		"--quiet",
	}
	if string(Settings.Mode) != "" {
		args = append(args, "--mode", string(Settings.Mode))
	}
	if string(Settings.Compression) != "" {
		args = append(args, "--compress", string(Settings.Compression))
	}
	args = append(args, Settings.AdditionalArgs...)

	cmd := exec.Command("vzdump", args...)
	if cmd.Err != nil {
		return nil, fmt.Errorf("image backup process failed: %v", cmd.Err)
	}

	return cmd, nil
}
