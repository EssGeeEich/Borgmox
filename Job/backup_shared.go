package Job

import (
	"Borgmox/ProxmoxCLI"
	"fmt"
	"os"
	"strconv"
	"time"
)

var cachedHostname string

func genArchivePrefix(hostname string, machineInfo ProxmoxCLI.MachineInfo) string {
	var prefix string

	// Hostname empty => System Hostname
	if hostname == "" {
		if cachedHostname == "" {
			cachedHostname, _ = os.Hostname()
		}
		hostname = cachedHostname
	}

	if hostname != "" {
		prefix = hostname + "-"
	}

	return prefix + string(machineInfo.Type) + "-" + strconv.FormatUint(machineInfo.VMID, 10) + "-"
}

func genArchiveName(prefix string, ts time.Time, archiveExtension string) string {
	archiveName := prefix

	ts = ts.UTC()
	archiveName += fmt.Sprintf("%d_%02d_%02d-%02d_%02d_%02d", ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second())
	if archiveExtension != "" {
		archiveName += "." + archiveExtension
	}
	return archiveName
}
