package main

import (
	"context"
	"io"
	"net/rpc"
	"sync"

	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/infamousjoeg/conceal/internal/migrate"
)

type memoryMigrator struct {
	store map[string][]byte
	mu    sync.Mutex
}

func (m *memoryMigrator) Name() string { return "aws-secretsmanager" }
func (m *memoryMigrator) Configure(ctx context.Context, in io.Reader, out io.Writer) error {
	return nil
}
func (m *memoryMigrator) DefaultRate() int { return 10 }
func (m *memoryMigrator) Put(ctx context.Context, key string, value []byte, meta map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.store == nil {
		m.store = make(map[string][]byte)
	}
	m.store[key] = append([]byte(nil), value...)
	return nil
}

type MigratorRPC struct{ impl migrate.Migrator }

func (s *MigratorRPC) Name(args struct{}, resp *string) error        { *resp = s.impl.Name(); return nil }
func (s *MigratorRPC) Configure(args struct{}, resp *struct{}) error { return nil }
func (s *MigratorRPC) DefaultRate(args struct{}, resp *int) error {
	*resp = s.impl.DefaultRate()
	return nil
}
func (s *MigratorRPC) Put(req struct {
	Key  string
	Val  []byte
	Meta map[string]string
}, resp *struct{}) error {
	return s.impl.Put(context.Background(), req.Key, req.Val, req.Meta)
}

type MigratorPlugin struct{ Impl migrate.Migrator }

func (p *MigratorPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &MigratorRPC{p.Impl}, nil
}
func (p *MigratorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &MigratorRPCClient{client: c}, nil
}

type MigratorRPCClient struct{ client *rpc.Client }

func (c *MigratorRPCClient) Name() string {
	var s string
	c.client.Call("MigratorRPC.Name", struct{}{}, &s)
	return s
}
func (c *MigratorRPCClient) Configure(ctx context.Context, in io.Reader, out io.Writer) error {
	return nil
}
func (c *MigratorRPCClient) DefaultRate() int {
	var i int
	c.client.Call("MigratorRPC.DefaultRate", struct{}{}, &i)
	return i
}
func (c *MigratorRPCClient) Put(ctx context.Context, key string, val []byte, meta map[string]string) error {
	return c.client.Call("MigratorRPC.Put", struct {
		Key  string
		Val  []byte
		Meta map[string]string
	}{key, val, meta}, nil)
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{ProtocolVersion: 1},
		Plugins: map[string]plugin.Plugin{
			"migrator": &MigratorPlugin{Impl: &memoryMigrator{}},
		},
		Logger: hclog.NewNullLogger(),
	})
}
