package Job

import "Borgmox/ProxmoxCLI"

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
	Frequency       NotificationFrequency
	TargetServer    string
	Topic           string
	SuccessPriority NotificationPriority
	FailurePriority NotificationPriority
}

type JobResult struct {
	Error            error
	SucceededBackups map[uint64]struct{}
	FailedBackups    map[uint64]error
}

type BackupJobData struct {
	Info ProxmoxCLI.MachineInfo
}

type BackupJobSettings struct {
	VmPool         []string
	ArchivePrefix  string
	BorgRepository string
	VmMode         VMBackupMode
	LxcMode        LXCBackupMode
	Notification   NotificationSettings
}

type JobConfigurations struct {
	Jobs map[string]BackupJobSettings
}
