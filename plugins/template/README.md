# Plugin Template

Use this directory as a starting point for a new migration plugin.
Implement the `Migrator` interface and build your binary:

```bash
go build -o conceal-migrate-myplugin .
```

Move the binary into your plugin directory and run `conceal provider` to verify
it loads.
