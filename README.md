# Conceal

Conceal is a small command line tool that saves secrets like passwords or API keys in your operating system's secret store. It works on macOS, Windows and Linux.

Secrets are stored in:

- **macOS** – the Keychain
- **Windows** – Credential Manager
- **Linux** – the keyring provided by `libsecret` (for example `gnome-keyring`)

On Linux you may need to install and run a secret service such as `gnome-keyring` before using Conceal.

## Installation

### Homebrew (macOS)

```bash
brew tap infamousjoeg/tap
brew install conceal
```

### Manual download

1. Download the latest release from the [GitHub releases page](https://github.com/infamousjoeg/conceal/releases).
2. Place the `conceal` executable somewhere in your `PATH` (for example `~/bin`).

## Basic usage

Set a secret:

```bash
conceal set my/secret
# or provide the value via a pipe
echo "value" | conceal set my/secret
```

Update a secret:

```bash
echo "new" | conceal update my/secret
```

Retrieve a secret:

```bash
conceal get my/secret            # copy to clipboard for 15 seconds
conceal get my/secret --stdout   # print the value so it can be piped
```

List all stored secrets:

```bash
conceal list
```

Remove a secret:

```bash
conceal unset my/secret
```

Use Conceal as a [Summon](https://cyberark.github.io/summon) provider:

```bash
conceal summon install
```

## Managing the secret store

Conceal uses your operating system's built in store. Normally there is nothing else you need to do. On Linux you may need to install `libsecret` and run a secret service like `gnome-keyring`. Conceal will report an error if it cannot reach the secret service.

## More information

For development guidelines see the [docs directory](docs/). Bugs and pull requests are welcome!

Conceal is released under the [Apache 2.0](LICENSE) license.
