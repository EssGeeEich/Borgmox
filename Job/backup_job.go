package Job

import (
	"Borgmox/ProxmoxCLI"
	"fmt"
	"log"
	"sort"
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
		}

		// Run the backups of all requested VMs, sorting by VMID.
		keys := sortedMapKeys(machines)

		for _, key := range keys {
			machine := machines[key]

			switch machine.Info.Type {
			case ProxmoxCLI.VM:
				if err := s.runVmBackup(machine, jobSettings); err != nil {
					result.FailedBackups[machine.Info.VMID] = err
					if jobSettings.Notification.Frequency == NF_EveryVmFinished {
						s.sendFailureNotification(jobSettings, "VM backup failed!", fmt.Sprintf("VM VMID %v: Backup failed!\n%v", machine.Info.VMID, err.Error()), []string{})
					}

				} else {
					result.SucceededBackups[machine.Info.VMID] = struct{}{}
					if jobSettings.Notification.Frequency == NF_EveryVmFinished {
						s.sendSuccessNotification(jobSettings, "VM backup completed!", fmt.Sprintf("VM VMID %v: Backup completed!", machine.Info.VMID), []string{})
					}
				}
			case ProxmoxCLI.LXC:
				if err := s.runLxcBackup(machine, jobSettings); err != nil {
					result.FailedBackups[machine.Info.VMID] = err
					if jobSettings.Notification.Frequency == NF_EveryVmFinished {
						s.sendFailureNotification(jobSettings, "LXC backup failed!", fmt.Sprintf("LXC VMID %v: Backup failed!\n%v", machine.Info.VMID, err.Error()), []string{})
					}
				} else {
					result.SucceededBackups[machine.Info.VMID] = struct{}{}
					if jobSettings.Notification.Frequency == NF_EveryVmFinished {
						s.sendSuccessNotification(jobSettings, "LXC backup completed!", fmt.Sprintf("LXC VMID %v: Backup completed!", machine.Info.VMID), []string{})
					}
				}
			}
		}

		// TODO: Prune?

		jobResults[jobName] = result

		if jobSettings.Notification.Frequency == NF_EntireJobFinished {
			if len(result.FailedBackups) > 0 && len(result.SucceededBackups) > 0 {
				s.sendFailureNotification(jobSettings, "Backup Job incomplete!", "Some VM/LXC backup jobs failed!", []string{})
			} else if len(result.FailedBackups) > 0 {
				s.sendFailureNotification(jobSettings, "Backup Job failed!", "All VM/LXC backup jobs failed!", []string{})
			} else if len(result.SucceededBackups) > 0 {
				s.sendSuccessNotification(jobSettings, "Backup Job completed!", "All VM/LXC backup jobs succeeded!", []string{})
			}
		}

	}

	return jobResults
}