# Conceal <!-- omit in toc -->

Conceal is a command-line utility that eases the interaction between developer and OSX Keychain Access. It is the open-source companion to [Summon](https://cyberark.github.io/summon) as every secret added using this tool into Keychain is added using Summon-compliant formatting. If you don't plan on using Summon, it's still a great Keychain management tool.

[![](https://github.com/infamousjoeg/conceal/workflows/Go/badge.svg?branch=master)](https://github.com/infamousjoeg/conceal/actions?query=workflow%3AGo) [![](https://img.shields.io/github/downloads/infamousjoeg/conceal/latest/total?color=blue&label=Download%20Latest%20Release&logo=github)](https://github.com/infamousjoeg/conceal/releases/latest)

## Table of Contents <!-- omit in toc -->
- [Requirements](#requirements)
- [Installation](#installation)
  - [Homebrew (MacOS)](#homebrew-macos)
  - [Manual](#manual)
- [Usage](#usage)
  - [Add a secret](#add-a-secret)
  - [Update a secret](#update-a-secret)
  - [Get a secret value](#get-a-secret-value)
  - [List Summon secrets](#list-summon-secrets)
  - [Remove a secret](#remove-a-secret)
  - [Install Conceal as Summon Provider](#install-conceal-as-summon-provider)
  - [Show a secret](#show-a-secret)
  - [Display Help](#display-help)
  - [Display Version](#display-version)
- [keychain Package](#keychain-package)
  - [Usage](#usage-1)
    - [func  AddSecret](#func--addsecret)
    - [func  DeleteSecret](#func--deletesecret)
    - [func  ListSecrets](#func--listsecrets)
    - [func  SecretExists](#func--secretexists)
    - [func  UpdateSecret](#func--updatesecret)
    - [func  GetSecret](#func--getsecret)
- [clipboard Package](#clipboard-package)
  - [Usage](#usage-2)
    - [func  Secret](#func--secret)
    - [func  SetupCloseHandler](#func--setupclosehandler)
- [Concept](#concept)
  - [Why Choose Conceal for Your Secret Management Needs?](#why-choose-conceal-for-your-secret-management-needs)
    - [**Leverage Existing Tools**](#leverage-existing-tools)
    - [**Seamless Integration with Summon**](#seamless-integration-with-summon)
    - [**Establish Secure Coding Practices Early**](#establish-secure-coding-practices-early)
    - [**Avoid Technical Debt**](#avoid-technical-debt)
    - [**Cost-Effective Solution**](#cost-effective-solution)
  - [Key Features of Conceal](#key-features-of-conceal)
  - [How to Get Started with Conceal](#how-to-get-started-with-conceal)
  - [Conclusion](#conclusion)
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

1. Download the latest release available at [GitHub Releases](https://github.com/infamousjoeg/conceal/releases).
2. Move the `conceal` executable file to a directory in your `PATH`. (I use `~/bin`.)
3. In Terminal, run the following command to make sure it's in your `PATH`: \
   `$ conceal`

## Usage

### Add a secret

`$ conceal set dockerhub/token`
`$ echo "my-secret-value" | conceal set dockerhub/token`

To add a secret to Keychain, call `conceal` and use the `set` command to pass the account name to add. You will be immediately prompted to provide a secret value in a secure manner or you can provide it via STDIN.

### Update a secret

`$ conceal update dockerhub/token`
`$ echo "my-new-secret-value" | conceal update dockerhub/token`

To update a secret in Keychain, call `conceal` and use the `update` command to pass the account name to update. You will be immediately prompted to provide a secret value in a secure manner or you can provide it via STDIN.

### Get a secret value

`$ conceal get dockerhub/token`

To retrieve a secret from Keychain, call `conceal` and use the `get` command to pass the account name to retrieve from. The secret value will be added to your clipboard for 15 seconds.

### List Summon secrets

`$ conceal list`

To list all secrets associated with Summon in Keychain, call `conceal` and use the `list` command to list all accounts present.

To filter the list further, pipe to `grep` like this `$ conceal list | grep dockerhub/`.

### Remove a secret

`$ conceal unset dockerhub/token`

To remove a secret that was added for Summon, call `conceal` and use the `unset` command to pass the account name to remove.

### Install Conceal as Summon Provider

`$ conceal summon install`

To install Conceal as a Summon provider, call `conceal` with the `summon install` command. This will install `conceal` as an available provider for Summon under the name `conceal_summon`. For more information about Summon's providers, check out the documentation at [cyberark.github.io/summon](https://cyberark.github.io/summon).

### Show a secret

**Note: This command is not recommended for use in scripts as it will print the secret to the terminal. It is only available for the Summon provider integration.**

`$ conceal show dockerhub/token`

To display a secret from Keychain to STDOUT, call `conceal` and use the `show` command to pass the account name to display. This is useful for debugging and testing purposes. It is used by Summon to retrieve the secret value from the `conceal_summon` provider.

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

#### func  AddSecret

```go
func AddSecret(secretID string, secret []byte)
```
AddSecret is a non-return function that adds the secret and secret value to
keychain.

#### func  DeleteSecret

```go
func DeleteSecret(secretID string)
```
DeleteSecret is a non-return function that removes the secret from keychain.

#### func  ListSecrets

```go
func ListSecrets() []string
```
ListSecrets is a string array function that returns all secrets in keychain with
the label `summon`.

#### func  SecretExists

```go
func SecretExists(secretID string) bool
```
SecretExists is a boolean function to verify a secret is present in keychain.

#### func  UpdateSecret

```go
func UpdateSecret(secretID string, secret []byte)
```
UpdateSecret is a non-return function that updates the secret value in keychain.

#### func  GetSecret

```go
func GetSecret(secretID string, delivery string)
```
GetSecret is a non-return function that retrieves the secret value from keychain and delivers it in the declared method. If `delivery` is set to `clipboard`, the secret value is copied to the clipboard. If a signal interrupt is detected, the content is immediately cleared. If `delivery` is set to `stdout`, the secret value is printed to the terminal.

## clipboard Package

```go
import "github.com/infamousjoeg/conceal/pkg/conceal/clipboard"
```

### Usage

#### func  Secret

```go
func Secret(secret string)
```
Secret is a non-return function that adds content to the host clipboard that
persists for 15 seconds. If a signal interrupt is detected, the content is
immediately cleared.

#### func  SetupCloseHandler

```go
func SetupCloseHandler()
```
SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
program if it receives an interrupt from the OS. We then handle this by calling
our clean up procedure and exiting the program.

## Concept

### Why Choose Conceal for Your Secret Management Needs?

In modern software development, securely managing secrets (such as API keys, passwords, and other sensitive data) is crucial. Conceal, developed by Joe Garcia, is a powerful utility designed to simplify and secure the management of these secrets. Here’s why you should consider using Conceal:

#### **Leverage Existing Tools**
**"Why not use what Steve and Bill gave us?"**
- Conceal allows developers to use built-in tools and environments (like macOS Keychain) to manage secrets without needing to commit any code or set up a dedicated secrets manager initially. This means you can start development immediately without additional setup overhead.

#### **Seamless Integration with Summon**
- Conceal works seamlessly with Summon, a tool that injects secrets as environment variables into your applications. This allows for easy transitioning between different environments without changing the code. As you move from development to staging to production, the secrets provider can change without any code modification, enhancing flexibility and security.

#### **Establish Secure Coding Practices Early**
**"You're establishing secure coding habits by starting development using environment variables out of the gate."**
- By using Conceal and Summon together, you adopt best practices from the start. Managing secrets via environment variables is a secure method that avoids hardcoding sensitive information in your application code, thus preventing technical debt and security vulnerabilities.

#### **Avoid Technical Debt**
**"...instead of creating technical debt that then becomes a problem later on down the line when a secrets manager needs to be baked into it."**
- Starting with good practices means you won't need to refactor your code later to integrate a secrets manager. Conceal helps avoid this costly and time-consuming process by providing a secure solution from the beginning.

#### **Cost-Effective Solution**
**"Free or overpay, which do you choose?"**
- Conceal leverages free, existing tools, avoiding the need for expensive enterprise secrets management solutions. This makes it a cost-effective choice, especially for startups and small teams.

### Key Features of Conceal

1. **Local Development-Friendly**: Ideal for local development environments where access to a full secrets management system might not be available.
2. **Ease of Use**: Simple commands to set and retrieve secrets, integrated smoothly with the development workflow.
3. **Security**: Ensures that secrets are not hardcoded, reducing the risk of accidental exposure.

### How to Get Started with Conceal

1. **Install Conceal**: Follow the instructions in the [Conceal GitHub repository](https://github.com/infamousjoeg/conceal) to install the utility.
2. **Set Secrets**: Use the `conceal set` command to securely store your secrets.
3. **Retrieve Secrets**: Integrate with Summon to retrieve secrets as environment variables in your application.

### Conclusion

Conceal is a powerful and useful utility for any developer looking to securely manage secrets without incurring additional setup costs or creating technical debt. By integrating with existing tools and promoting secure practices from the start, Conceal ensures your development process remains efficient, secure, and cost-effective. Choose Conceal to simplify your secret management and focus on building great software.

For more information and to get started, visit the [Conceal GitHub page](https://github.com/infamousjoeg/conceal).

## Maintainer

[@infamousjoeg](https://github.com/infamousjoeg)

[![Buy me a coffee][buymeacoffee-shield]][buymeacoffee]

[buymeacoffee]: https://www.buymeacoffee.com/infamousjoeg
[buymeacoffee-shield]: https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png

## Contributions

Pull Requests are currently being accepted.  Please read and follow the guidelines laid out in [CONTRIBUTING.md]().

## License

[Apache 2.0](LICENSE)
