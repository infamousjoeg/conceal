# Conceal <!-- omit in toc -->

Conceal is a command-line utility that eases the interaction between developer and OSX Keychain Access. It is the open-source companion to [Summon](https://cyberark.github.io/summon) as every secret added using this tool into Keychain is added using Summon-compliant formatting.

[![](https://github.com/infamousjoeg/conceal/workflows/Go/badge.svg?branch=master)](https://github.com/infamousjoeg/conceal/actions?query=workflow%3AGo) [![](https://img.shields.io/github/downloads/infamousjoeg/conceal/latest/total?color=blue&label=Download%20Latest%20Release&logo=github)](https://github.com/infamousjoeg/conceal/releases/latest)

## Table of Contents <!-- omit in toc -->
- [Requirements](#requirements)
- [Installation](#installation)
  - [Homebrew (MacOS)](#homebrew-macos)
  - [Manual](#manual)
- [Usage](#usage)
  - [Add a secret](#add-a-secret)
  - [Get a secret value](#get-a-secret-value)
  - [List Summon secrets](#list-summon-secrets)
  - [Update a secret](#update-a-secret)
  - [Remove a secret](#remove-a-secret)
  - [Display Help](#display-help)
  - [Display Version](#display-version)
- [keychain Package](#keychain-package)
  - [Usage](#usage-1)
    - [func AddSecret](#func-addsecret)
    - [func UpdateSecret](#func-updatesecret)
    - [func DeleteSecret](#func-deletesecret)
    - [func ListSecrets](#func-listsecrets)
    - [func SecretExists](#func-secretexists)
- [clipboard Package](#clipboard-package)
  - [Usage](#usage-2)
    - [func Secret](#func-secret)
    - [func SetupCloseHandler](#func-setupclosehandler)
- [Maintainer](#maintainer)
- [Contributions](#contributions)
- [License](#license)

## Requirements

* MacOS

## Installation

### Homebrew (MacOS)

```shell
brew tap infamousjoeg/tap
brew install conceal
```

### Manual

1. Download the latest release available at [GitHub Releases](https://github.com/infamousjoeg/go-conceal/releases).
2. Move the `conceal` executable file to a directory in your `PATH`. (I use `~/bin`.)
3. In Terminal, run the following command to make sure it's in your `PATH`: \
   `$ conceal`

## Usage

### Add a secret

`$ conceal set dockerhub/token`

To add a secret to Keychain, call `conceal` and use the `set` command to pass the account name to add. You will be immediately prompted to provide a secret value in a secure manner.

### Get a secret value

`$ conceal get dockerhub/token`

To retrieve a secret from Keychain, call `conceal` and use the `get` command to pass the account name to retrieve from. The secret value will be added to your clipboard for 15 seconds.

### List Summon secrets

`$ conceal list`

To list all secrets associated with Summon in Keychain, call `conceal` and use the `list` command to list all accounts present.

To filter the list further, pipe to `grep` like this `$ conceal list | grep dockerhub/`.

### Update a secret

`$ conceal update dockerhub/token`

To update a secret existing in Keychain, call `conceal` and use the `update` command to pass the account name to update. You will be immediately prompted to provide a secret value in a secure manner.

### Remove a secret

`$ conceal unset dockerhub/token`

To remove a secret that was added for Summon, call `conceal` and use the `unset` command to pass the account name to remove.

### Display Help

`$ conceal help`

To display the help message, just call `conceal help`.

`$ conceal help [COMMAND]`

To display the help message for a specific command, just call `conceal help` and provide the command name, such as `set` or `get`.

### Display Version

`$ conceal version`

To display the current version, call `conceal` with the `version` command.

## keychain Package

```go
import "github.com/infamousjoeg/conceal/pkg/conceal/keychain"
```

### Usage

#### func AddSecret

```go
func AddSecret(secretID string, secret []byte)
```
AddSecret is a non-return function that adds the secret and secret value to
keychain.

#### func UpdateSecret

```go
func UpdateSecret(secretID string, secret []byte)
```
UpdateSecret is a non-return function that updates an existing secret's secret value.

#### func DeleteSecret

```go
func DeleteSecret(secretID string)
```
DeleteSecret is a non-return function that removes the secret from keychain

#### func ListSecrets

```go
func ListSecrets() []string
```
ListSecrets is a string array function that returns all secrets in keychain with
the label `summon`.

#### func SecretExists

```go
func SecretExists(secretID string) bool
```
SecretExists is a boolean function to verify a secret is present in keychain

## clipboard Package

```go
import "github.com/infamousjoeg/conceal/pkg/conceal/clipboard"
```

### Usage

#### func Secret

```go
func Secret(secret string)
```
Secret is a non-return function that adds content to the host clipboard that
persists for 15 seconds. If a signal interrupt is detected, the content is
immediately cleared.

#### func SetupCloseHandler

```go
func SetupCloseHandler()
```
SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
program if it receives an interrupt from the OS. We then handle this by calling
our clean up procedure and exiting the program.

## Maintainer

[@infamousjoeg](https://github.com/infamousjoeg)

[![Buy me a coffee][buymeacoffee-shield]][buymeacoffee]

[buymeacoffee]: https://www.buymeacoffee.com/infamousjoeg
[buymeacoffee-shield]: https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png

## Contributions

Pull Requests are currently being accepted.  Please read and follow the guidelines laid out in [CONTRIBUTING.md]().

## License

[Apache 2.0](LICENSE)
