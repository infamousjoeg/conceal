package main

import (
	"context"
	"io"
	"net/rpc"

	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/infamousjoeg/conceal/internal/migrate"
)

// myMigrator demonstrates the Migrator interface.
type myMigrator struct{}

func (m *myMigrator) Name() string                                                     { return "example" }
func (m *myMigrator) Configure(ctx context.Context, in io.Reader, out io.Writer) error { return nil }
func (m *myMigrator) DefaultRate() int                                                 { return 10 }
func (m *myMigrator) Put(ctx context.Context, key string, value []byte, meta map[string]string) error {
	// TODO: write secret to your backend
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

type MigratorRPCClient struct{ client *rpc.Client }

func (c *MigratorRPCClient) Name() string {
	var s string
	_ = c.client.Call("MigratorRPC.Name", struct{}{}, &s)
	return s
}
func (c *MigratorRPCClient) Configure(ctx context.Context, in io.Reader, out io.Writer) error {
	return nil
}
func (c *MigratorRPCClient) DefaultRate() int {
	var i int
	_ = c.client.Call("MigratorRPC.DefaultRate", struct{}{}, &i)
	return i
}
func (c *MigratorRPCClient) Put(ctx context.Context, key string, val []byte, meta map[string]string) error {
	return c.client.Call("MigratorRPC.Put", struct {
		Key  string
		Val  []byte
		Meta map[string]string
	}{key, val, meta}, nil)
}

type MigratorPlugin struct{ Impl migrate.Migrator }

func (p *MigratorPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &MigratorRPC{p.Impl}, nil
}
func (p *MigratorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &MigratorRPCClient{client: c}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{ProtocolVersion: 1},
		Plugins: map[string]plugin.Plugin{
			"migrator": &MigratorPlugin{Impl: &myMigrator{}},
		},
		Logger: hclog.NewNullLogger(),
	})
}
