package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func defaultProviderDirectory() string {
	if path := os.Getenv("SUMMON_PROVIDER_PATH"); path != "" {
		return path
	}
	if runtime.GOOS == "windows" {
		base := os.Getenv("ProgramW6432")
		if base == "" {
			base = os.Getenv("ProgramFiles")
		}
		if base == "" {
			base = `C:\Program Files`
		}
		return filepath.Join(base, "Cyberark Conjur", "Summon", "Providers")
	}
	return "/usr/local/lib/summon"
}

func newInstallCommand() *cobra.Command {
	command := &cobra.Command{Use: "install", Short: "Install a native Summon provider executable", Args: cobra.NoArgs}
	command.Flags().String("dir", "", "Provider directory (default: SUMMON_PROVIDER_PATH or Summon's standard directory)")
	command.Flags().Bool("force", false, "Replace an existing regular provider file")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		directory, _ := cmd.Flags().GetString("dir")
		if directory == "" {
			directory = defaultProviderDirectory()
		}
		force, _ := cmd.Flags().GetBool("force")
		source, err := os.Executable()
		if err != nil {
			return err
		}
		name := "conceal_summon"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		target, err := installProvider(source, directory, name, force)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Provider installed: %s\nSet Summon's --provider option to this path.\n", target)
		return err
	}
	return command
}

func installProvider(source, directory, name string, force bool) (string, error) {
	directory, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", fmt.Errorf("create provider directory (use --dir for a writable location): %w", err)
	}
	target := filepath.Join(directory, name)
	input, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer func() { _ = input.Close() }()
	sourceInfo, err := input.Stat()
	if err != nil {
		return "", err
	}
	if existing, statErr := os.Lstat(target); statErr == nil {
		if !force {
			return "", fmt.Errorf("provider already exists at %s; use --force to replace it", target)
		}
		if !existing.Mode().IsRegular() || os.SameFile(sourceInfo, existing) {
			return "", fmt.Errorf("refusing to replace a non-regular file or the running executable")
		}
	} else if !os.IsNotExist(statErr) {
		return "", statErr
	}
	output, err := os.CreateTemp(directory, ".conceal-provider-*")
	if err != nil {
		return "", err
	}
	temp := output.Name()
	defer func() { _ = os.Remove(temp) }()
	_, copyErr := io.Copy(output, input)
	if copyErr == nil {
		copyErr = output.Chmod(0755)
	}
	if copyErr == nil {
		copyErr = output.Sync()
	}
	closeErr := output.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	if force {
		err = os.Rename(temp, target)
	} else {
		// A same-directory hard link installs without overwriting a concurrent file.
		err = os.Link(temp, target)
	}
	if err != nil {
		return "", fmt.Errorf("install provider: %w", err)
	}
	return target, nil
}
