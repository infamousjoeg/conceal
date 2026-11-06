package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/infamousjoeg/conceal/internal/redact"
	"github.com/spf13/cobra"
)

var environmentName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type binding struct{ name, secret string }

func parseBindings(values []string) ([]binding, error) {
	result := make([]binding, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		name, id, ok := strings.Cut(value, "=")
		if !ok || !environmentName.MatchString(name) {
			return nil, fmt.Errorf("--env must be NAME=secret-name with a valid environment variable name")
		}
		if err := validateSecretName(id); err != nil {
			return nil, err
		}
		key := name
		if runtime.GOOS == "windows" {
			key = strings.ToUpper(key)
		}
		if seen[key] {
			return nil, fmt.Errorf("duplicate environment variable %q", name)
		}
		seen[key] = true
		result = append(result, binding{name, id})
	}
	return result, nil
}

func newRunCommand(store secretStore) *cobra.Command {
	command := &cobra.Command{
		Use:   "run --env NAME=secret-name [--env ...] -- command [args...]",
		Short: "Run a trusted process with selected secrets; mask exact values in output",
		Args:  cobra.MinimumNArgs(1),
	}
	command.Flags().StringArray("env", nil, "Map an environment variable to a stored secret (repeatable)")
	command.Flags().Bool("clean-env", false, "Do not inherit the parent environment")
	command.Flags().Bool("no-redact", false, "Disable exact-value masking of child stdout and stderr")
	command.Flags().Duration("timeout", 0, "Stop the direct child after this duration (0 means no timeout)")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Context().Err() != nil {
			return &statusError{130, "child process interrupted"}
		}
		if cmd.ArgsLenAtDash() != 0 {
			return fmt.Errorf("separate the child command with --")
		}
		values, _ := cmd.Flags().GetStringArray("env")
		if len(values) == 0 {
			return fmt.Errorf("at least one --env NAME=secret-name is required")
		}
		bindings, err := parseBindings(values)
		if err != nil {
			return err
		}
		timeout, _ := cmd.Flags().GetDuration("timeout")
		if timeout < 0 {
			return fmt.Errorf("--timeout cannot be negative")
		}
		path, err := exec.LookPath(args[0])
		if err != nil {
			return fmt.Errorf("find child executable: %w", err)
		}
		var secrets [][]byte
		defer func() {
			for _, value := range secrets {
				clear(value)
			}
		}()
		clean, _ := cmd.Flags().GetBool("clean-env")
		env := []string{}
		if !clean {
			env = os.Environ()
		}
		for _, binding := range bindings {
			value, err := store.ReadSecret(binding.secret)
			if err != nil {
				return fmt.Errorf("resolve secret %q: %w", binding.secret, err)
			}
			secrets = append(secrets, value)
			if strings.IndexByte(string(value), 0) >= 0 {
				return fmt.Errorf("secret %q contains NUL and cannot be an environment variable", binding.secret)
			}
			env = append(env, binding.name+"="+string(value))
		}
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		child := exec.CommandContext(ctx, path, args[1:]...)
		child.Env = env
		child.Stdin = cmd.InOrStdin()
		child.Stdout = cmd.OutOrStdout()
		child.Stderr = cmd.ErrOrStderr()
		child.WaitDelay = 2 * time.Second
		noRedact, _ := cmd.Flags().GetBool("no-redact")
		var stdout, stderr *redact.Writer
		if !noRedact {
			stdout = redact.New(child.Stdout, secrets)
			stderr = redact.New(child.Stderr, secrets)
			child.Stdout, child.Stderr = stdout, stderr
		}
		err = child.Run()
		if stdout != nil {
			if flushErr := errors.Join(stdout.Close(), stderr.Close()); flushErr != nil {
				return fmt.Errorf("write redacted child output: %w", flushErr)
			}
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return &statusError{124, "child process timed out"}
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return &statusError{130, "child process interrupted"}
		}
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			code := exit.ExitCode()
			if code < 0 {
				code = 1
			}
			return &statusError{code: code}
		}
		return err
	}
	return command
}
