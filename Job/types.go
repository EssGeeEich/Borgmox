package Job

import (
	"Borgmox/BorgCLI"
	"Borgmox/ProxmoxCLI"
)

type VMBackupMode string
type LXCBackupMode string

const (
	VMBKP_Image VMBackupMode = "image"

	LXCBKP_Image LXCBackupMode = "image"
)

type NotificationFrequency string
type NotificationPriority string

const (
	NF_Never             NotificationFrequency = "never"
	NF_EveryVmFinished   NotificationFrequency = "every vm"
	NF_EntireJobFinished NotificationFrequency = "single job"

	NP_Max      NotificationPriority = "max"
	NP_Urgent   NotificationPriority = "urgent"
	NP_High     NotificationPriority = "high"
	NP_Default  NotificationPriority = "default"
	NP_Low      NotificationPriority = "low"
	NP_Min      NotificationPriority = "min"
	NP_Disabled NotificationPriority = "off"
)

type NotificationTargetInfo struct {
	Frequency          NotificationFrequency
	SuccessPriority    NotificationPriority
	FailurePriority    NotificationPriority
	SuccessEmailTarget string
	FailureEmailTarget string
}

type NotificationSettings struct {
	BackupTargetInfo NotificationTargetInfo
	PruneTargetInfo  NotificationTargetInfo

	TargetServer string
	AuthUser     string
	AuthPassword string
	Topic        string
}

type BackupJobData struct {
	Info ProxmoxCLI.MachineInfo
}

type BackupJobSettings struct {
	ArchivePrefix string
	VmPool        []string
	VmMode        VMBackupMode
	LxcMode       LXCBackupMode
	Notification  NotificationSettings
	Borg          BorgCLI.BorgSettings
}

type JobResult struct {
	Error            error
	SucceededBackups map[uint64]struct{}
	FailedBackups    map[uint64]error
	SucceededPrunes  map[uint64]struct{}
	FailedPrunes     map[uint64]error
}

type JobConfigurations struct {
	Jobs map[string]BackupJobSettings
}
