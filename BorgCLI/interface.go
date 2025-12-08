package BorgCLI

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	gv "github.com/hashicorp/go-version"
)

var regexBorgVersion *regexp.Regexp

type CreateArchiveSettings struct {
	Compression    string
	Comment        string
	AdditionalArgs []string
}

func CreateArchiveExec(settings BorgSettings, ArchiveName string, Settings CreateArchiveSettings, cmdSource *exec.Cmd) (*exec.Cmd, error) {
	Settings.AdditionalArgs = append(Settings.AdditionalArgs, "--content-from-command")
	cmd, err := CreateArchive(settings, ArchiveName, Settings)
	if err != nil {
		return nil, err
	}

	cmd.Args = append(cmd.Args, "--")
	cmd.Args = append(cmd.Args, cmdSource.Args...)
	return cmd, nil
}

func CreateArchive(settings BorgSettings, ArchiveName string, Settings CreateArchiveSettings) (*exec.Cmd, error) {
	args := []string{
		"create",
		"--files-cache",
		"disabled",
	}

	if settings.RemotePath != "" {
		args = append(args, "--remote-path", settings.RemotePath)
	}

	args = append(args, "--stdin-name", ArchiveName)

	if Settings.Compression != "" {
		args = append(args, "--compression", Settings.Compression)
	}
	if Settings.Comment != "" {
		args = append(args, "--comment", Settings.Comment)
	}
	args = append(args, Settings.AdditionalArgs...)
	args = append(args, settings.Repository+"::"+ArchiveName)

	cmd := exec.Command("borg", args...)
	if cmd.Err != nil {
		return nil, fmt.Errorf("archive creation process failed: %v", cmd.Err)
	}
	cmd.Env = os.Environ()

	// Will be escaped by cmd.Exec
	cmd.Env = append(cmd.Env, "BORG_PASSPHRASE="+settings.Passphrase)

	return cmd, nil
}

func PruneByPrefix(settings BorgSettings, ArchivePrefix string) (*exec.Cmd, error) {
	if !settings.Prune.Enabled {
		return nil, errors.New("prune is disabled in the current borg configuration")
	}

	args := []string{
		"prune",
	}

	if settings.RemotePath != "" {
		args = append(args, "--remote-path", settings.RemotePath)
	}

	if settings.Prune.KeepWithin != "" {
		args = append(args, "--keep-within", settings.Prune.KeepWithin)
	}
	if settings.Prune.KeepLast > 0 {
		args = append(args, "--keep-last", strconv.FormatUint(settings.Prune.KeepLast, 10))
	}
	if settings.Prune.KeepMinutely > 0 {
		args = append(args, "--keep-minutely", strconv.FormatUint(settings.Prune.KeepMinutely, 10))
	}
	if settings.Prune.KeepHourly > 0 {
		args = append(args, "--keep-hourly", strconv.FormatUint(settings.Prune.KeepHourly, 10))
	}
	if settings.Prune.KeepDaily > 0 {
		args = append(args, "--keep-daily", strconv.FormatUint(settings.Prune.KeepDaily, 10))
	}
	if settings.Prune.KeepWeekly > 0 {
		args = append(args, "--keep-weekly", strconv.FormatUint(settings.Prune.KeepWeekly, 10))
	}
	if settings.Prune.KeepMonthly > 0 {
		args = append(args, "--keep-monthly", strconv.FormatUint(settings.Prune.KeepMonthly, 10))
	}
	if settings.Prune.KeepYearly > 0 {
		args = append(args, "--keep-yearly", strconv.FormatUint(settings.Prune.KeepYearly, 10))
	}

	args = append(args, "--glob-archives", ArchivePrefix+"*")
	args = append(args, settings.Repository)

	cmd := exec.Command("borg", args...)
	if cmd.Err != nil {
		return nil, fmt.Errorf("archive pruning process failed: %v", cmd.Err)
	}
	cmd.Env = os.Environ()
	// Will be escaped by cmd.Exec
	cmd.Env = append(cmd.Env, "BORG_PASSPHRASE="+settings.Passphrase)
	return cmd, nil
}

func Compact(settings BorgSettings) (*exec.Cmd, error) {
	if !settings.Prune.Compact {
		return nil, errors.New("compact is disabled in the current borg configuration")
	}

	args := []string{
		"compact",
	}

	if settings.RemotePath != "" {
		args = append(args, "--remote-path", settings.RemotePath)
	}

	args = append(args, settings.Repository)

	cmd := exec.Command("borg", args...)
	if cmd.Err != nil {
		return nil, fmt.Errorf("archive compacting process failed: %v", cmd.Err)
	}
	cmd.Env = os.Environ()
	// Will be escaped by cmd.Exec
	cmd.Env = append(cmd.Env, "BORG_PASSPHRASE="+settings.Passphrase)
	return cmd, nil
}

func GetVersion() (*gv.Version, error) {
	cmd := exec.Command("borg", "-V")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("borg returned an error: %w", err)
	} else {
		if regexBorgVersion == nil {
			regexBorgVersion = regexp.MustCompile(`borg\s*([\d\.]+).*`)
		}
		submatches := regexBorgVersion.FindStringSubmatch(string(output))
		if submatches == nil || len(submatches) != 2 {
			return nil, fmt.Errorf("borg returned a non-parsable string: %v", string(output))
		}

		if ver, err := gv.NewVersion(submatches[1]); err != nil {
			return nil, fmt.Errorf("couldn't parse borg version number: %w", err)
		} else {
			return ver, nil
		}
	}
}
