package migrate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
)

// Manager loads Migrator plugins from the plugin directory.

type Manager struct {
	Dir string
}

func pluginDir() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "conceal", "plugins")
}

func NewManager() *Manager {
	return &Manager{Dir: pluginDir()}
}

func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.Dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "conceal-migrate-") {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func (m *Manager) Install(src string) error {
	if err := os.MkdirAll(m.Dir, 0o755); err != nil {
		return err
	}
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	dst := filepath.Join(m.Dir, filepath.Base(src))
	if err := os.WriteFile(dst, in, 0o755); err != nil {
		return err
	}
	return nil
}

func (m *Manager) load(name string) (*plugin.Client, Migrator, error) {
	path := filepath.Join(m.Dir, name)
	if _, err := os.Stat(path); err != nil {
		return nil, nil, err
	}
	handshake := plugin.HandshakeConfig{ProtocolVersion: 1}
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshake,
		Cmd:              execCommand(path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		Logger:           hclog.NewNullLogger(),
	})
	rpcClient, err := client.Client()
	if err != nil {
		return nil, nil, err
	}
	raw, err := rpcClient.Dispense("migrator")
	if err != nil {
		client.Kill()
		return nil, nil, err
	}
	mig := raw.(Migrator)
	return client, mig, nil
}

func execCommand(path string) *exec.Cmd {
	return exec.Command(path)
}

func (m *Manager) Configure(ctx context.Context, name string, in io.Reader, out io.Writer) error {
	cl, mig, err := m.load(name)
	if err != nil {
		return err
	}
	defer cl.Kill()
	return mig.Configure(ctx, in, out)
}

func (m *Manager) Migrate(ctx context.Context, name string, src map[string][]byte, rate int, cont bool, out io.Writer) error {
	cl, mig, err := m.load(name)
	if err != nil {
		return err
	}
	defer cl.Kill()
	return m.MigrateWith(ctx, mig, src, rate, cont, out)
}

func (m *Manager) MigrateWith(ctx context.Context, mig Migrator, src map[string][]byte, rate int, cont bool, out io.Writer) error {
	if rate <= 0 {
		rate = mig.DefaultRate()
	}
	tb := time.NewTicker(time.Second / time.Duration(rate))
	defer tb.Stop()
	type failure struct {
		key string
		err error
	}
	var fails []failure
	for k, v := range src {
		<-tb.C
		if err := mig.Put(ctx, k, v, nil); err != nil {
			fails = append(fails, failure{k, err})
			if !cont {
				break
			}
		}
	}
	if len(fails) > 0 {
		_, _ = fmt.Fprintln(out, "Failures:")
		for _, f := range fails {
			_, _ = fmt.Fprintf(out, "%s: %v\n", f.key, f.err)
		}
		return fmt.Errorf("%d secrets failed", len(fails))
	}
	return nil
}
