package main

import (
	"Borgmox/BorgCLI"
	"Borgmox/Job"
	"Borgmox/ProxmoxCLI"
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
	"github.com/pelletier/go-toml/v2"
)

func runMain() error {
	var jobData Job.JobData

	//if len(os.Args) != 2 {
	//	return fmt.Errorf("usage: %s [input.toml]", os.Args[0])
	//}

	//if os.Args[1] == "--stdout-sample-toml" || true {
	if true {
		jobData.BackupJobs = map[string]Job.BackupJobSettings{
			"My Job": {
				ArchivePrefix: "",
				VmPool:        []string{"my_proxmox_vm_pool", "my_proxmox_lxc_pool"},
				VmMode:        Job.VMBKP_Image,
				LxcMode:       Job.LXCBKP_Image,
				Notification: Job.NotificationSettings{
					BackupTargetInfo: Job.NotificationTargetInfo{
						Frequency:          Job.NF_EntireJobFinished,
						SuccessPriority:    Job.NP_Default,
						FailurePriority:    Job.NP_High,
						SuccessEmailTarget: "",
						FailureEmailTarget: "",
					},
					PruneTargetInfo: Job.NotificationTargetInfo{
						Frequency:          Job.NF_EveryVmFinished,
						SuccessPriority:    Job.NP_Default,
						FailurePriority:    Job.NP_High,
						SuccessEmailTarget: "",
						FailureEmailTarget: "",
					},
					TargetServer: "https://ntfy.my.local",
					AuthUser:     "my_user_or_empty_for_access_token",
					AuthPassword: "my_user_password_or_token",
					Topic:        "MyNotificationTopic",
				},
				Borg: BorgCLI.BorgSettings{
					Repository: "ssh://my_borg_repo",
					RemotePath: "/my/remote/borg/path/if/needed/or/empty",
					Passphrase: "my-borg-passphrase",
					Prune: BorgCLI.BorgPruneSettings{
						Enabled:      false,
						KeepWithin:   "15D",
						KeepLast:     10,
						KeepMinutely: 0,
						KeepHourly:   0,
						KeepDaily:    0,
						KeepWeekly:   8,
						KeepMonthly:  12,
						KeepYearly:   10,
					},
				},
			},
		}
		if data, err := toml.Marshal(jobData); err != nil {
			return fmt.Errorf("couldn't encode empty toml template: %w", err)
		} else {
			os.Stdout.Write(data)
			os.Exit(0)
		}
	}

	if jobFile, err := os.ReadFile(os.Args[1]); err != nil {
		return fmt.Errorf("couldn't open toml input file: %w", err)
	} else if err := toml.Unmarshal(jobFile, &jobData); err != nil {
		return fmt.Errorf("couldn't decode toml input file: %w", err)
	}

	proxmoxVer, err := ProxmoxCLI.GetVersion()
	if err != nil {
		return fmt.Errorf("cannot verify proxmox version: %w", err)
	}

	if targetMinimumVersion, err := version.NewVersion("8.0.0"); err != nil {
		return fmt.Errorf("cannot compare proxmox version; semver error: %w", err)
	} else if proxmoxVer.LessThan(targetMinimumVersion) {
		return fmt.Errorf("current proxmox version: %v, minimum version required: %v", proxmoxVer.Original(), targetMinimumVersion.Original())
	}

	borgVer, err := BorgCLI.GetVersion()
	if err != nil {
		return fmt.Errorf("cannot verify borg version: %w", err)
	}

	if targetMinimumVersion, err := version.NewVersion("1.2.4"); err != nil {
		return fmt.Errorf("cannot compare borg version; semver error: %w", err)
	} else if borgVer.LessThan(targetMinimumVersion) {
		return fmt.Errorf("current borg version: %v, minimum version required: %v", borgVer.Original(), targetMinimumVersion.Original())
	}

	// Run the effective backup job
	operationError := errors.New("operation failed")

	r := jobData.RunJob()
	for _, val := range r {
		if val.Error != nil {
			return operationError
		}
		if len(val.FailedBackups) > 0 {
			return operationError
		}
		if len(val.FailedPrunes) > 0 {
			return operationError
		}
	}

	return nil
}

func main() {
	if err := runMain(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
