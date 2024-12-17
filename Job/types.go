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

type NotificationSettings struct {
	Frequency          NotificationFrequency
	TargetServer       string
	AuthUser           string
	AuthPassword       string
	Topic              string
	SuccessPriority    NotificationPriority
	FailurePriority    NotificationPriority
	SuccessEmailTarget string
	FailureEmailTarget string
}

type JobResult struct {
	Error            error
	SucceededBackups map[uint64]string
	FailedBackups    map[uint64]error
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

type JobConfigurations struct {
	Jobs map[string]BackupJobSettings
}
