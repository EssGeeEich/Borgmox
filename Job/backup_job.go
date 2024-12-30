package Job

import (
	"Borgmox/ProxmoxCLI"
	"fmt"
	"log"
	"sort"
	"strconv"
)

func sortedMapKeys(m map[uint64]BackupJobData) []uint64 {
	keys := make([]uint64, 0, len(m))

	for key := range m {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}

func (s *JobData) RunJob() map[string]JobResult {
	jobResults := make(map[string]JobResult, len(s.BackupJobs))

	skippedMachines := make(map[uint64]struct{}, 64)
	var err error

	// Look through all Backup Jobs
nextJob:
	for jobName, jobSettings := range s.BackupJobs {
		machines := make(map[uint64]BackupJobData, 64)
		var newMachines []ProxmoxCLI.MachineInfo

		// Look through all requested VM pools
		for _, vmPool := range jobSettings.VmPool {
			if newMachines, err = ProxmoxCLI.GetMachinesByPool(vmPool); err != nil {
				jobResults[jobName] = JobResult{
					Error: fmt.Errorf("cannot receive Proxmox machines with pool %v in Backup Job %v: %w", vmPool, jobName, err),
				}
				continue nextJob
			}

			// Array of VMs to associative map of VMs.
			// Avoid duplicate backups of VMs that appear in two different pools.

			for _, machine := range newMachines {
				switch machine.Type {
				case ProxmoxCLI.LXC:
					machines[machine.VMID] = BackupJobData{
						Info: machine,
					}
				case ProxmoxCLI.VM:
					machines[machine.VMID] = BackupJobData{
						Info: machine,
					}
				default:
					if _, ok := skippedMachines[machine.VMID]; !ok {
						skippedMachines[machine.VMID] = struct{}{}
						log.Printf("Invalid machine type '%v' for VMID %v, skipping.", string(machine.Type), machine.VMID)
					}
				}
			}
		}

		result := JobResult{
			SucceededBackups: make(map[uint64]struct{}, len(machines)),
			FailedBackups:    make(map[uint64]error, len(machines)),
			SucceededPrunes:  make(map[uint64]struct{}, len(machines)),
			FailedPrunes:     make(map[uint64]error, len(machines)),
		}

		// Run the backups of all requested VMs, sorting by VMID.
		keys := sortedMapKeys(machines)

		prunableMachines := []struct {
			Bjd BackupJobData
			Js  BackupJobSettings
		}{}

		for _, key := range keys {
			machine := machines[key]

			switch machine.Info.Type {
			case ProxmoxCLI.VM:
				if err := s.runVmBackup(jobName, machine, jobSettings); err != nil {
					result.FailedBackups[machine.Info.VMID] = err
					if jobSettings.Notification.BackupTargetInfo.Frequency == NF_EveryVmFinished {
						s.sendFailureNotification(jobSettings, jobSettings.Notification.BackupTargetInfo, "VM backup failed!", fmt.Sprintf("VM VMID %v: Backup failed!\n%v", machine.Info.VMID, err.Error()), []string{})
					}
				} else {
					result.SucceededBackups[machine.Info.VMID] = struct{}{}
					prunableMachines = append(prunableMachines, struct {
						Bjd BackupJobData
						Js  BackupJobSettings
					}{
						Bjd: machine,
						Js:  jobSettings,
					})

					if jobSettings.Notification.BackupTargetInfo.Frequency == NF_EveryVmFinished {
						s.sendSuccessNotification(jobSettings, jobSettings.Notification.BackupTargetInfo, "VM backup completed!", fmt.Sprintf("VM VMID %v: Backup completed!", machine.Info.VMID), []string{})
					}
				}
			case ProxmoxCLI.LXC:
				if err := s.runLxcBackup(jobName, machine, jobSettings); err != nil {
					result.FailedBackups[machine.Info.VMID] = err
					if jobSettings.Notification.BackupTargetInfo.Frequency == NF_EveryVmFinished {
						s.sendFailureNotification(jobSettings, jobSettings.Notification.BackupTargetInfo, "LXC backup failed!", fmt.Sprintf("LXC VMID %v: Backup failed!\n%v", machine.Info.VMID, err.Error()), []string{})
					}
				} else {
					result.SucceededBackups[machine.Info.VMID] = struct{}{}
					prunableMachines = append(prunableMachines, struct {
						Bjd BackupJobData
						Js  BackupJobSettings
					}{
						Bjd: machine,
						Js:  jobSettings,
					})

					if jobSettings.Notification.BackupTargetInfo.Frequency == NF_EveryVmFinished {
						s.sendSuccessNotification(jobSettings, jobSettings.Notification.BackupTargetInfo, "LXC backup completed!", fmt.Sprintf("LXC VMID %v: Backup completed!", machine.Info.VMID), []string{})
					}
				}
			}
		}

		// TODO: Prune?
		if jobSettings.Borg.Prune.Enabled {
			for _, pruneData := range prunableMachines {
				if err := s.runPrune(pruneData.Bjd, pruneData.Js); err != nil {
					result.FailedPrunes[pruneData.Bjd.Info.VMID] = err
					if jobSettings.Notification.PruneTargetInfo.Frequency == NF_EveryVmFinished {
						s.sendFailureNotification(pruneData.Js, jobSettings.Notification.PruneTargetInfo, "Archive prune failed!", fmt.Sprintf("VMID %v Archive: Prune failed!\n%v", pruneData.Bjd.Info.VMID, err.Error()), []string{})
					}
				} else {
					result.SucceededPrunes[pruneData.Bjd.Info.VMID] = struct{}{}
					if jobSettings.Notification.PruneTargetInfo.Frequency == NF_EveryVmFinished {
						s.sendSuccessNotification(pruneData.Js, jobSettings.Notification.PruneTargetInfo, "Archive prune completed!", fmt.Sprintf("VMID %v Archive: Prune completed!", pruneData.Bjd.Info.VMID), []string{})
					}
				}
			}
		}

		jobResults[jobName] = result

		if jobSettings.Notification.BackupTargetInfo.Frequency == NF_EntireJobFinished {
			if len(result.FailedBackups) > 0 && len(result.SucceededBackups) > 0 {
				strMessage := "Succeeded:\n"
				for vmid := range result.SucceededBackups {
					strMessage += "- " + strconv.FormatUint(vmid, 10) + "\n"
				}
				strMessage += "\nFailed:\n"
				for vmid, err := range result.FailedBackups {
					strMessage += "- " + strconv.FormatUint(vmid, 10) + " (" + err.Error() + ")\n"
				}
				target := jobSettings.Notification.BackupTargetInfo
				target.FailurePriority = highestPriority(target.FailurePriority, target.SuccessPriority)
				s.sendFailureNotification(jobSettings, target, "Backup Job incomplete!", "Some VM/LXC backup jobs failed!\n"+strMessage, []string{})
			} else if len(result.FailedBackups) > 0 {
				strMessage := "\n"
				for vmid, err := range result.FailedBackups {
					strMessage += "- " + strconv.FormatUint(vmid, 10) + " (" + err.Error() + ")\n"
				}
				s.sendFailureNotification(jobSettings, jobSettings.Notification.BackupTargetInfo, "Backup Job failed!", "All VM/LXC backup jobs failed!\n"+strMessage, []string{})
			} else if len(result.SucceededBackups) > 0 {
				s.sendSuccessNotification(jobSettings, jobSettings.Notification.BackupTargetInfo, "Backup Job completed!", "All VM/LXC backup jobs succeeded!", []string{})
			}
		}

		if jobSettings.Notification.PruneTargetInfo.Frequency == NF_EntireJobFinished {
			if len(result.FailedPrunes) > 0 && len(result.SucceededPrunes) > 0 {
				strMessage := "Succeeded:\n"
				for vmid := range result.SucceededPrunes {
					strMessage += "- " + strconv.FormatUint(vmid, 10) + "\n"
				}
				strMessage += "\nFailed:\n"
				for vmid, err := range result.FailedPrunes {
					strMessage += "- " + strconv.FormatUint(vmid, 10) + " (" + err.Error() + ")\n"
				}
				target := jobSettings.Notification.PruneTargetInfo
				target.FailurePriority = highestPriority(target.FailurePriority, target.SuccessPriority)
				s.sendFailureNotification(jobSettings, target, "Prune Job incomplete!", "Some VM/LXC prune jobs failed!\n"+strMessage, []string{})
			} else if len(result.FailedPrunes) > 0 {
				strMessage := "\n"
				for vmid, err := range result.FailedPrunes {
					strMessage += "- " + strconv.FormatUint(vmid, 10) + " (" + err.Error() + ")\n"
				}
				s.sendFailureNotification(jobSettings, jobSettings.Notification.PruneTargetInfo, "Prune Job failed!", "All VM/LXC prune jobs failed!\n"+strMessage, []string{})
			} else if len(result.SucceededPrunes) > 0 {
				s.sendSuccessNotification(jobSettings, jobSettings.Notification.PruneTargetInfo, "Prune Job completed!", "All VM/LXC prune jobs succeeded!", []string{})
			}
		}

	}

	return jobResults
}
