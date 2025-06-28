# Writing a Migration Plugin

Plugins extend `conceal migrate` by implementing the `Migrator` interface. Each plugin is a standalone Go program built with [go-plugin](https://github.com/hashicorp/go-plugin).

1. Copy `plugins/template` as a starting point.
2. Implement the `Put` method to upload secrets to your backend.
3. Build the plugin:

```bash
cd myplugin
go build -o conceal-migrate-myplugin
```

4. Move the binary into `$XDG_CONFIG_HOME/conceal/plugins` and run `conceal provider` to verify it is detected.

See `plugins/aws-secretsmanager` for a complete example.
