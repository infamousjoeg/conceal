# Plugin Template

Use this directory as a starting point for writing a new migration plugin. Implement the `Migrator` interface and build your binary:

```bash
go build -o conceal-migrate-myplugin .
```

Copy the binary into your Conceal plugin directory and run `conceal provider` to verify it loads.
