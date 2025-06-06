package main

import (
	"borgmox/BorgCLI"
	"borgmox/Job"
	"borgmox/ProxmoxCLI"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
	"github.com/pelletier/go-toml/v2"
)

func runMain() error {
	var jobData Job.JobData

	dontBackup := flag.Bool("no-backup", false, "disables backing up any VM/LXC, useful for only running prune jobs")
	dontPrune := flag.Bool("no-prune", false, "disables all prune jobs, useful for only running backup jobs")
	outputSampleToml := flag.Bool("stdout-sample-toml", false, "disables all processing and prints a sample toml file")

	flag.Parse()

	if *outputSampleToml {
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
					TargetServer: "",
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
						Compact:      true,
						KeepWithin:   "15d",
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
			os.Stdout.WriteString("\n")
			os.Exit(0)
		}
	}

	if len(flag.Args()) != 1 {
		return fmt.Errorf("usage: %s [input.toml]", os.Args[0])
	}

	if jobFile, err := os.ReadFile(flag.Args()[0]); err != nil {
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

	r := jobData.RunJob(Job.JobOptions{
		DontBackup: *dontBackup,
		DontPrune:  *dontPrune,
	})

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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
