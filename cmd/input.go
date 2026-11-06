package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const maxInputSize = 1024 * 1024

func terminalInput(cmd *cobra.Command) (*os.File, bool) {
	file, ok := cmd.InOrStdin().(*os.File)
	return file, ok && term.IsTerminal(int(file.Fd()))
}

func secretName(cmd *cobra.Command, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], validateSecretName(args[0])
	}
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
	_, terminal := terminalInput(cmd)
	if nonInteractive || !terminal {
		return "", fmt.Errorf("secret name is required when input is noninteractive")
	}
	if _, err := fmt.Fprint(cmd.ErrOrStderr(), "Secret name: "); err != nil {
		return "", err
	}
	var name string
	if _, err := fmt.Fscanln(cmd.InOrStdin(), &name); err != nil {
		return "", fmt.Errorf("read secret name: %w", err)
	}
	return name, validateSecretName(name)
}

func readSecretValue(cmd *cobra.Command) ([]byte, error) {
	file, terminal := terminalInput(cmd)
	if terminal {
		nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
		if nonInteractive {
			return nil, fmt.Errorf("pipe the secret value when using --non-interactive")
		}
		if _, err := fmt.Fprint(cmd.ErrOrStderr(), "Secret value: "); err != nil {
			return nil, err
		}
		value, err := term.ReadPassword(int(file.Fd()))
		_, _ = fmt.Fprintln(cmd.ErrOrStderr())
		if err != nil {
			clear(value)
		}
		return value, err
	}
	value, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), maxInputSize+1))
	if err != nil {
		clear(value)
		return nil, fmt.Errorf("read secret input: %w", err)
	}
	if len(value) > maxInputSize {
		clear(value)
		return nil, fmt.Errorf("secret input exceeds 1 MiB")
	}
	raw, _ := cmd.Flags().GetBool("raw")
	if !raw && bytes.HasSuffix(value, []byte("\n")) {
		value = value[:len(value)-1]
		value = bytes.TrimSuffix(value, []byte("\r"))
	}
	return value, nil
}

func writeSecretCommand(store secretStore, update bool) *cobra.Command {
	name, description := "set", "Add a secret from a secure prompt or stdin"
	if update {
		name, description = "update", "Update an existing secret from a secure prompt or stdin"
	}
	command := &cobra.Command{Use: name + " [secret-name]", Short: description, Args: cobra.MaximumNArgs(1)}
	if !update {
		command.Aliases = []string{"add", "create"}
	}
	command.Flags().Bool("raw", false, "Preserve stdin exactly, including its final newline")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		id, err := secretName(cmd, args)
		if err != nil {
			return err
		}
		value, err := readSecretValue(cmd)
		if err != nil {
			return err
		}
		defer clear(value)
		if update {
			err = store.UpdateSecret(id, value)
		} else {
			err = store.AddSecret(id, value)
		}
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "Secret saved.")
		return err
	}
	return command
}
