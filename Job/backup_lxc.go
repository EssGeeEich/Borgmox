package Job

import (
	"Borgmox/BorgCLI"
	"Borgmox/ProxmoxCLI"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func (s *JobData) runLxcBackup(bjd BackupJobData, js BackupJobSettings) error {
	switch js.LxcMode {
	case LXCBKP_Image:
		BackupSettings := ProxmoxCLI.StartImageBackupSettings{
			Compression: ProxmoxCLI.DontCompress,
			Mode:        ProxmoxCLI.Snapshot,
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

		archiveName := genArchiveName(js.ArchivePrefix, "tar", bjd.Info, time.Now())

		var cmdRunAll *exec.Cmd
		if cmdRunAll, err = BorgCLI.CreateArchiveExec(js.BorgRepository, archiveName, ArchiveSettings, cmdBackup); err != nil {
			return err
		}

		cmdRunAll.Stdout = os.Stdout
		cmdRunAll.Stderr = os.Stderr

		log.Printf("Now backing up LXC %v (%v)", bjd.Info.Name, bjd.Info.VMID)
		if err := cmdRunAll.Run(); err != nil {
			return err
		}

		return nil

	default:
		return fmt.Errorf("unimplemented backup method for LXCs: %v", string(js.LxcMode))
	}
}