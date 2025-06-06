package Job

import (
	"borgmox/BorgCLI"
	"log"
	"os"
	"os/exec"
)

func (s *JobData) runPrune(bjd BackupJobData, js BackupJobSettings) error {
	archivePrefix := genArchivePrefix(js.ArchivePrefix, bjd.Info)
	var cmdRunAll *exec.Cmd
	var err error

	if cmdRunAll, err = BorgCLI.PruneByPrefix(js.Borg, archivePrefix); err != nil {
		return err
	}

	cmdRunAll.Stdout = os.Stdout
	cmdRunAll.Stderr = os.Stderr

	log.Printf("Now pruning archives for VM/LXC %v (%v)", bjd.Info.Name, bjd.Info.VMID)
	if err := cmdRunAll.Run(); err != nil {
		return err
	}

	return nil
}

func (s *JobData) runCompact(js BackupJobSettings) error {
	var cmdRunAll *exec.Cmd
	var err error

	if cmdRunAll, err = BorgCLI.Compact(js.Borg); err != nil {
		return err
	}

	cmdRunAll.Stdout = os.Stdout
	cmdRunAll.Stderr = os.Stderr

	log.Printf("Now compacting borg repository...")
	if err := cmdRunAll.Run(); err != nil {
		return err
	}

	return nil
}
