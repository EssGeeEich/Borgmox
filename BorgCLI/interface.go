package BorgCLI

import (
	"fmt"
	"os/exec"
	"regexp"

	gv "github.com/hashicorp/go-version"
)

var regexBorgVersion *regexp.Regexp

type CreateArchiveSettings struct {
	Compression    string
	Comment        string
	AdditionalArgs []string
}

func CreateArchiveExec(Repo string, ArchiveName string, Settings CreateArchiveSettings, cmdSource *exec.Cmd) (*exec.Cmd, error) {
	Settings.AdditionalArgs = append(Settings.AdditionalArgs, "--content-from-command")
	cmd, err := CreateArchive(Repo, ArchiveName, Settings)
	if err != nil {
		return nil, err
	}

	cmd.Args = append(cmd.Args, "--")
	cmd.Args = append(cmd.Args, cmdSource.Args...)
	return cmd, nil
}

func CreateArchive(Repo string, ArchiveName string, Settings CreateArchiveSettings) (*exec.Cmd, error) {
	args := []string{
		"create",
		"--files-cache",
		"disabled",
	}

	args = append(args, "--stdin-name", ArchiveName)

	if Settings.Compression != "" {
		args = append(args, "--compression", Settings.Compression)
	}
	if Settings.Comment != "" {
		args = append(args, "--comment", Settings.Comment)
	}
	args = append(args, Settings.AdditionalArgs...)
	args = append(args, Repo+"::"+ArchiveName)

	cmd := exec.Command("borg", args...)
	if cmd.Err != nil {
		return nil, fmt.Errorf("archive creation process failed: %v", cmd.Err)
	}
	return cmd, nil
}

func GetVersion() (*gv.Version, error) {
	cmd := exec.Command("borg", "-V")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("borg returned an error: %w", err)
	} else {
		if regexBorgVersion == nil {
			regexBorgVersion = regexp.MustCompile(`borg.*([\d\.]+).*`)
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
