package main

import (
	"Borgmox/BorgCLI"
	"Borgmox/ProxmoxCLI"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/go-version"
)

type VMBackupMode string
type LXCBackupMode string

const (
	VMBKP_Image VMBackupMode = "image"

	LXCBKP_Image LXCBackupMode = "image"
)

func genArchiveName(archivePrefix string, archiveExtension string, machineInfo ProxmoxCLI.MachineInfo, ts time.Time) string {
	var archiveName string
	if archivePrefix != "" {
		archiveName += archivePrefix + "-"
	}
	archiveName += string(machineInfo.Type) + "-" + string(machineInfo.VMID) + "-"
	ts = ts.UTC()
	archiveName += fmt.Sprintf("%d_%02d_%02d-%02d_%02d_%02d", ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second())
	if archiveExtension != "" {
		archiveName += "." + archiveExtension
	}
	return archiveName
}

func runMain() error {
	var VM_Mode VMBackupMode
	var LXC_Mode LXCBackupMode
	var Archive_Prefix string
	var Repository string

	{
		var VM_Mode_Input string
		var LXC_Mode_Input string
		flag.StringVar(&VM_Mode_Input, "vm_mode", "image", "backup method for proxmox VMs (image)")
		flag.StringVar(&LXC_Mode_Input, "lxc_mode", "image", "backup method for proxmox LXCs (image)")
		flag.StringVar(&Archive_Prefix, "prefix", "", "prefix for the archive name. if empty, will use the hostname (name of the node).")
		flag.StringVar(&Repository, "repo", "", "repository path")

		flag.Parse()

		switch VM_Mode_Input {
		case "image":
			VM_Mode = VMBKP_Image
		default:
			return fmt.Errorf("invalid vm_mode: %v", VM_Mode_Input)
		}

		switch LXC_Mode_Input {
		case "image":
			LXC_Mode = LXCBKP_Image
		default:
			return fmt.Errorf("invalid lxc_mode: %v", VM_Mode_Input)
		}

		if Archive_Prefix == "" {
			Archive_Prefix, _ = os.Hostname()
		}

		if Repository == "" {
			return fmt.Errorf("invalid repo: %v", Repository)
		}
	}

	proxmoxVer, err := ProxmoxCLI.GetVersion()
	if err != nil {
		return fmt.Errorf("cannot verify proxmox version: %w", err)
	}

	if targetMinimumVersion, err := version.NewVersion("8.2.8"); err != nil {
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

	machines := make(map[uint64](ProxmoxCLI.MachineInfo), 64)

	{
		skippedMachines := make(map[uint64]interface{}, 64)
		var newMachines []ProxmoxCLI.MachineInfo

		for _, proxmox_pool := range flag.Args() {
			if newMachines, err = ProxmoxCLI.GetMachinesByPool(proxmox_pool); err != nil {
				return fmt.Errorf("cannot receive Proxmox machines with pool %v: %w", proxmox_pool, err)
			}

			// Array of VMs to associative map of VMs.
			// Avoid duplicate backups of VMs that appear in two different pools.

			for _, machine := range newMachines {
				switch machine.Type {
				case ProxmoxCLI.LXC:
					machines[machine.VMID] = machine
				case ProxmoxCLI.VM:
					machines[machine.VMID] = machine
				default:
					if _, ok := skippedMachines[machine.VMID]; !ok {
						skippedMachines[machine.VMID] = struct{}{}
						log.Printf("Invalid machine type '%v' for VMID %v, skipping.", string(machine.Type), machine.VMID)
					}
				}

			}
		}
	}

	pruneableMachines := make(map[uint64]bool, len(machines))

	// Run the effective backup job
	for id, machine := range machines {
		pruneableMachines[id] = false

		switch machine.Type {
		case ProxmoxCLI.VM:
			switch VM_Mode {
			case VMBKP_Image:
				BackupSettings := ProxmoxCLI.StartImageBackupSettings{
					Compression: ProxmoxCLI.DontCompress,
					Mode:        ProxmoxCLI.Snapshot,
				}
				ArchiveSettings := BorgCLI.CreateArchiveSettings{
					Compression: "auto,zlib",
				}

				var cmdBackup *exec.Cmd
				if cmdBackup, err = ProxmoxCLI.StartImageBackup(machine.VMID, BackupSettings); err != nil {
					log.Printf("Error backing up VM, skipping. VMID %v: %v", machine.VMID, err)
					continue
				}

				archiveName := genArchiveName(Archive_Prefix, "vma", machine, time.Now())

				var cmdRunAll *exec.Cmd
				if cmdRunAll, err = BorgCLI.CreateArchiveExec(Repository, archiveName, ArchiveSettings, cmdBackup); err != nil {
					log.Printf("Error backing up VM, skipping. VMID %v: %v", machine.VMID, err)
					continue
				}

				log.Printf("Now backing up VM %v (%v)", machine.Name, machine.VMID)
				if out, err := cmdRunAll.Output(); err != nil {
					log.Printf("Error backing up VM. VMID %v: %v", machine.VMID, err)
					log.Printf("Output: %v", string(out))
					continue
				}

				pruneableMachines[id] = true
			default:
				return fmt.Errorf("unimplemented backup method for VMs: %v", string(VM_Mode))
			}
		case ProxmoxCLI.LXC:
			switch LXC_Mode {
			case LXCBKP_Image:
				BackupSettings := ProxmoxCLI.StartImageBackupSettings{
					Compression: ProxmoxCLI.DontCompress,
					Mode:        ProxmoxCLI.Snapshot,
				}
				ArchiveSettings := BorgCLI.CreateArchiveSettings{
					Compression: "auto,zlib",
				}

				var cmdBackup *exec.Cmd
				if cmdBackup, err = ProxmoxCLI.StartImageBackup(machine.VMID, BackupSettings); err != nil {
					log.Printf("Error backing up VM, skipping. VMID %v: %v", machine.VMID, err)
					continue
				}

				archiveName := genArchiveName(Archive_Prefix, "tar", machine, time.Now())

				var cmdRunAll *exec.Cmd
				if cmdRunAll, err = BorgCLI.CreateArchiveExec(Repository, archiveName, ArchiveSettings, cmdBackup); err != nil {
					log.Printf("Error backing up VM, skipping. VMID %v: %v", machine.VMID, err)
					continue
				}

				log.Printf("Now backing up VM %v (%v)", machine.Name, machine.VMID)
				if out, err := cmdRunAll.Output(); err != nil {
					log.Printf("Error backing up VM. VMID %v: %v", machine.VMID, err)
					log.Printf("Output: %v", string(out))
					continue
				}

				pruneableMachines[id] = true
			default:
				return fmt.Errorf("unimplemented backup method for LXCs: %v", string(LXC_Mode))
			}
		}
	}

	return nil
}

func main() {
	if err := runMain(); err != nil {
		log.Fatal(err)
	}
}
