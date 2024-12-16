package Job

import (
	"Borgmox/ProxmoxCLI"
	"fmt"
	"strconv"
	"time"
)

func genArchiveName(archivePrefix string, archiveExtension string, machineInfo ProxmoxCLI.MachineInfo, ts time.Time) string {
	var archiveName string
	if archivePrefix != "" {
		archiveName += archivePrefix + "-"
	}
	archiveName += string(machineInfo.Type) + "-" + strconv.FormatUint(machineInfo.VMID, 10) + "-"
	ts = ts.UTC()
	archiveName += fmt.Sprintf("%d_%02d_%02d-%02d_%02d_%02d", ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second())
	if archiveExtension != "" {
		archiveName += "." + archiveExtension
	}
	return archiveName
}
