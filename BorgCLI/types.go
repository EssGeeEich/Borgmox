package BorgCLI

type BorgPruneSettings struct {
	Enabled      bool
	KeepWithin   string
	KeepLast     uint64
	KeepMinutely uint64
	KeepHourly   uint64
	KeepDaily    uint64
	KeepWeekly   uint64
	KeepMonthly  uint64
	KeepYearly   uint64
}

type BorgSettings struct {
	Repository string
	RemotePath string
	Passphrase string

	Prune BorgPruneSettings
}
