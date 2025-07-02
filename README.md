# Conceal

![Build](https://github.com/infamousjoeg/conceal/actions/workflows/go-test.yml/badge.svg)
![License](https://img.shields.io/github/license/infamousjoeg/conceal)
![Go Report](https://goreportcard.com/badge/github.com/infamousjoeg/conceal)

**Conceal** is a friendly CLI that stores secrets in your operating system's own
credential manager. It works on macOS Keychain, Windows Credential Manager and
Linux keyrings such as `gnome-keyring`. Secrets never touch disk and can be used
with [Summon](https://cyberark.github.io/summon) and other tools.

## Quickstart

```bash
# install from source
go install github.com/infamousjoeg/conceal@latest

# add a secret
echo "hunter2" | conceal set demo/password

# retrieve the value
conceal get demo/password --stdout
```

See the [installation guide](docs/user-guide/INSTALLATION.md) for packages and
binary downloads.

## Features

- Secure secret storage using the OS keychain
- Simple commands for set, get, update and list
- Piping support via `--stdout`
- Plugin system for migrating secrets to other providers
- Summon integration for applications

## Plugins

Plugins extend `conceal migrate` and live in
`$XDG_CONFIG_HOME/conceal/plugins`. Build the sample AWS plugin with:

```bash
cd plugins/aws-secretsmanager
go build -o conceal-migrate-aws-secretsmanager .
```

Move the binary to the plugin directory then run `conceal provider` to list it.
Full details are in [docs/MIGRATION.md](docs/MIGRATION.md).

## Documentation

- [User Guide](docs/user-guide/INSTALLATION.md)
- [Developer Guide](docs/developer-guide/CONTRIBUTING.md)
- [Architecture](docs/architecture/DESIGN.md)
- [Examples](docs/examples/basic-usage.md)

## Contributing

We welcome pull requests! Please read the
[contributing guide](docs/developer-guide/CONTRIBUTING.md) and run the tests with
`go test ./...` before submitting.

## License

Conceal is released under the [Apache 2.0](LICENSE) license.
