package main

import (
	"Borgmox/BorgCLI"
	"Borgmox/Job"
	"Borgmox/ProxmoxCLI"
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
	"github.com/pelletier/go-toml/v2"
)

func runMain() error {
	var jobData Job.JobData

	if len(os.Args) != 2 {
		return fmt.Errorf("usage: %s [input.toml]", os.Args[0])
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
	jobData.RunJob()

	return nil
}

func main() {
	if err := runMain(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
