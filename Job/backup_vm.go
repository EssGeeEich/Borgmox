package Job

import (
	"Borgmox/BorgCLI"
	"Borgmox/ProxmoxCLI"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func (s *JobData) runVmBackup(jobName string, bjd BackupJobData, js BackupJobSettings) error {
	switch js.VmMode {
	case VMBKP_Image:
		BackupSettings := ProxmoxCLI.StartImageBackupSettings{
			Compression: ProxmoxCLI.DontCompress,
			Mode:        ProxmoxCLI.Snapshot,
			AdditionalArgs: []string{
				"--job-id",
				"borgmox-vm-vmid_" + strconv.FormatUint(bjd.Info.VMID, 10) + "-id_" + removeSpaces(bjd.Info.ID) + "-job_" + removeSpaces(jobName),
			},
		}
		ArchiveSettings := BorgCLI.CreateArchiveSettings{
			Compression: "auto,zlib",
			AdditionalArgs: []string{
				"--progress",
			},
		}

		var cmdBackup *exec.Cmd
		var err error
		if cmdBackup, err = ProxmoxCLI.StartImageBackup(bjd.Info.VMID, BackupSettings); err != nil {
			return err
		}

		archivePrefix := genArchivePrefix(js.ArchivePrefix, bjd.Info)
		archiveName := genArchiveName(archivePrefix, time.Now(), "vma")

		var cmdRunAll *exec.Cmd
		if cmdRunAll, err = BorgCLI.CreateArchiveExec(js.Borg, archiveName, ArchiveSettings, cmdBackup); err != nil {
			return err
		}

		cmdRunAll.Stdout = os.Stdout
		cmdRunAll.Stderr = os.Stderr

		log.Printf("Now backing up VM %v (%v)", bjd.Info.Name, bjd.Info.VMID)
		if err := cmdRunAll.Run(); err != nil {
			return err
		}

		return nil

	default:
		return fmt.Errorf("unimplemented backup method for LXCs: %v", string(js.LxcMode))
	}
}
