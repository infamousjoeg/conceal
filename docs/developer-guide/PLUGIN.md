# Writing a Migration Plugin

Plugins extend `conceal migrate` to upload secrets elsewhere. Each plugin is its
own Go binary built with [go-plugin](https://github.com/hashicorp/go-plugin).

1. Copy `plugins/template` to a new directory.
2. Fill in the `Migrator` methods, especially `Put` which stores a secret.
3. Build your plugin:
   ```bash
   go build -o conceal-migrate-myplugin .
   ```
4. Move the binary to `$XDG_CONFIG_HOME/conceal/plugins`.
5. Run `conceal provider` to check it appears.

A detailed example lives in `plugins/aws-secretsmanager`.
