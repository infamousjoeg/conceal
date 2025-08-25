# Conceal <!-- omit in toc -->

Conceal is a cross-platform command-line utility that eases the interaction between developers and native OS credential stores. It securely manages secrets using **macOS Keychain** and **Windows Credential Manager**. Conceal is the open-source companion to [Summon](https://cyberark.github.io/summon), storing all secrets in Summon-compliant formatting for seamless integration.

[![Go Test](https://github.com/infamousjoeg/conceal/workflows/Go%20Test/badge.svg)](https://github.com/infamousjoeg/conceal/actions/workflows/go-test.yml) [![Code Quality](https://github.com/infamousjoeg/conceal/workflows/Code%20Quality/badge.svg)](https://github.com/infamousjoeg/conceal/actions/workflows/golangci-lint.yml) [![Security](https://github.com/infamousjoeg/conceal/workflows/Security%20Checks/badge.svg)](https://github.com/infamousjoeg/conceal/actions/workflows/security.yml) [![](https://img.shields.io/github/downloads/infamousjoeg/conceal/latest/total?color=blue&label=Download%20Latest%20Release&logo=github)](https://github.com/infamousjoeg/conceal/releases/latest)

## Table of Contents <!-- omit in toc -->
- [Supported Platforms](#supported-platforms)
- [Requirements](#requirements)
- [Installation](#installation)
  - [Homebrew (macOS)](#homebrew-macos)
  - [Windows](#windows)
  - [Manual Installation](#manual-installation)
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
- [Cross-Platform Development](#cross-platform-development)
- [Maintainer](#maintainer)
- [Contributions](#contributions)
- [License](#license)

## Supported Platforms

| Platform | Secret Store | Status | Notes |
|----------|--------------|---------|-------|
| 🍎 **macOS** | Keychain Access | ✅ Full Support | Native integration via Security Framework |
| 🪟 **Windows** | Credential Manager | ✅ Full Support | Native integration via Windows API |
| 🐧 **Linux** | - | ❌ Not Supported | Future consideration for keyring/libsecret |

## Requirements

### macOS
* macOS 10.12 (Sierra) or later
* Xcode Command Line Tools (for Keychain access)

### Windows  
* Windows 10 or later
* Windows Server 2016 or later

## Installation

### Homebrew (macOS)

```bash
brew tap infamousjoeg/tap
brew install conceal
```

### Windows

#### Via GitHub Releases
1. Download `conceal_Windows_x86_64.zip` from [GitHub Releases](https://github.com/infamousjoeg/conceal/releases/latest)
2. Extract `conceal.exe` to a folder in your PATH (e.g., `C:\Program Files\Conceal\`)
3. Open Command Prompt or PowerShell and verify: `conceal version`

#### Via PowerShell (Future)
```powershell
# Chocolatey package coming soon
# choco install conceal
```

### Manual Installation

#### All Platforms
1. Download the appropriate binary for your platform from [GitHub Releases](https://github.com/infamousjoeg/conceal/releases/latest):
   - **macOS**: `conceal_Darwin_x86_64.tar.gz` (Intel) or `conceal_Darwin_arm64.tar.gz` (Apple Silicon)
   - **Windows**: `conceal_Windows_x86_64.zip` 
2. Extract and move the executable to a directory in your `PATH`
3. Verify installation: `conceal version`

#### Build from Source
```bash
git clone https://github.com/infamousjoeg/conceal.git
cd conceal
go build -o conceal .
```

## Usage

### Add a secret

`$ conceal set dockerhub/token`
`$ echo "my-secret-value" | conceal set dockerhub/token`

To add a secret to your OS credential store, call `conceal` and use the `set` command to pass the account name to add. You will be immediately prompted to provide a secret value in a secure manner or you can provide it via STDIN.

**Platform-specific behavior:**
- **macOS**: Stores in Keychain Access with service "summon"
- **Windows**: Stores in Credential Manager under "Generic Credentials" with target "summon/{secret_name}"

### Update a secret

`$ conceal update dockerhub/token`
`$ echo "my-new-secret-value" | conceal update dockerhub/token`

To update an existing secret in your OS credential store, call `conceal` and use the `update` command to pass the account name to update. You will be immediately prompted to provide a secret value in a secure manner or you can provide it via STDIN.

### Get a secret value

`$ conceal get dockerhub/token`

To retrieve a secret from your OS credential store, call `conceal` and use the `get` command to pass the account name to retrieve from. The secret value will be added to your clipboard for 15 seconds and automatically cleared for security.

### List Summon secrets

`$ conceal list`

To list all secrets associated with Summon in your OS credential store, call `conceal` and use the `list` command to list all accounts present.

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

## Go Package Documentation

### keychain Package (Cross-Platform)

```go
import "github.com/infamousjoeg/conceal/pkg/conceal/keychain"
```

The keychain package provides a unified interface for interacting with native OS credential stores across platforms.

#### Platform Support
- **macOS**: Uses Security Framework and Keychain Access
- **Windows**: Uses Windows Credential Manager API
- **Other platforms**: Returns appropriate error messages

#### Functions

##### func AddSecret

```go
func AddSecret(secretID string, secret []byte) error
```
AddSecret stores a secret in the platform's native credential store. Returns an error if the secret already exists or if the operation fails.

**Platform behavior:**
- **macOS**: Stores in Keychain with service "summon"
- **Windows**: Stores in Credential Manager with target "summon/{secretID}"

##### func DeleteSecret

```go
func DeleteSecret(secretID string) error
```
DeleteSecret removes a secret from the platform's credential store. Returns an error if the secret doesn't exist or removal fails.

##### func GetSecret

```go
func GetSecret(secretID string, delivery string) error
```
GetSecret retrieves a secret and delivers it via the specified method:
- `"clipboard"`: Copies to clipboard for 15 seconds with auto-clear
- `"stdout"`: Prints to standard output (for Summon integration)

##### func ListSecrets

```go
func ListSecrets() []QueryResult
```
ListSecrets returns all Summon-compatible secrets from the credential store.

```go
type QueryResult struct {
    Account string  // The secret identifier
}
```

##### func SecretExists

```go
func SecretExists(secretID string) bool
```
SecretExists checks if a secret exists in the credential store without retrieving its value.

##### func UpdateSecret

```go
func UpdateSecret(secretID string, secret []byte) error
```
UpdateSecret modifies an existing secret's value. Returns an error if the secret doesn't exist.

### clipboard Package

```go
import "github.com/infamousjoeg/conceal/pkg/conceal/clipboard"
```

The clipboard package provides secure clipboard management with automatic clearing.

#### Functions

##### func Secret

```go
func Secret(secret string)
```
Secret copies the provided string to the system clipboard and automatically clears it after 15 seconds. If a signal interrupt (Ctrl+C) is detected, the clipboard is immediately cleared for security.

**Cross-platform support:**
- **macOS**: Uses `pbcopy` and `pbpaste` via atotto/clipboard
- **Windows**: Uses Windows Clipboard API via atotto/clipboard
- **Linux**: Uses xclip/xsel via atotto/clipboard

##### func SetupCloseHandler

```go
func SetupCloseHandler()
```
SetupCloseHandler creates a signal handler that listens for OS interrupts (SIGTERM, SIGINT) and immediately clears the clipboard when received. This ensures secrets are not left in the clipboard if the application is terminated unexpectedly.

### wincred Package (Windows-Specific)

```go
import "github.com/infamousjoeg/conceal/pkg/conceal/wincred"
```

The wincred package provides Windows-specific credential management functionality. This package is only available on Windows builds.

#### Functions

##### func SecretExists

```go
func SecretExists(secretID string) bool
```

##### func ListSecrets  

```go
func ListSecrets() ([]SecretInfo, error)
```

##### func AddSecret

```go
func AddSecret(secretID string, secret []byte) error
```

##### func DeleteSecret

```go  
func DeleteSecret(secretID string) error
```

##### func GetSecret

```go
func GetSecret(secretID string, delivery string) ([]byte, error)
```

##### func UpdateSecret

```go
func UpdateSecret(secretID string, secret []byte) error
```

**Note**: This package uses build tags and is only compiled on Windows systems. It provides the underlying implementation for Windows Credential Manager integration.

## Concept

### Why Choose Conceal for Your Secret Management Needs?

In modern software development, securely managing secrets (such as API keys, passwords, and other sensitive data) is crucial. Conceal, developed by Joe Garcia, is a powerful utility designed to simplify and secure the management of these secrets. Here’s why you should consider using Conceal:

#### **Leverage Existing Tools**
**"Why not use what the OS provides?"**
- Conceal allows developers to use built-in OS credential stores (macOS Keychain and Windows Credential Manager) to manage secrets without needing to set up dedicated secrets management infrastructure initially. This means you can start development immediately without additional setup overhead or cloud dependencies.

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

1. **Cross-Platform**: Native support for macOS Keychain and Windows Credential Manager
2. **Local Development-Friendly**: Ideal for local development environments without cloud dependencies
3. **Ease of Use**: Simple, consistent CLI commands across all supported platforms
4. **Security**: Leverages OS-native encryption and access controls
5. **Summon Integration**: Seamless compatibility with CyberArk Summon for environment injection
6. **Zero Configuration**: Works immediately after installation with no setup required

### How to Get Started with Conceal

1. **Install Conceal**: Follow the instructions in the [Conceal GitHub repository](https://github.com/infamousjoeg/conceal) to install the utility.
2. **Set Secrets**: Use the `conceal set` command to securely store your secrets.
3. **Retrieve Secrets**: Integrate with Summon to retrieve secrets as environment variables in your application.

### Conclusion

Conceal is a powerful and useful utility for any developer looking to securely manage secrets without incurring additional setup costs or creating technical debt. By integrating with existing tools and promoting secure practices from the start, Conceal ensures your development process remains efficient, secure, and cost-effective. Choose Conceal to simplify your secret management and focus on building great software.

For more information and to get started, visit the [Conceal GitHub page](https://github.com/infamousjoeg/conceal).

## Cross-Platform Development

Conceal is built with Go and uses platform-specific build tags to provide native integration with each operating system's credential store.

### Architecture

```
pkg/conceal/
├── keychain/
│   ├── keychain_darwin.go     # macOS implementation
│   ├── keychain_windows.go    # Windows implementation  
│   └── keychain_other.go      # Unsupported platforms
└── wincred/
    ├── wincred_windows.go     # Windows Credential Manager API
    └── wincred_other.go       # Non-Windows stub
```

### Building for Different Platforms

```bash
# Build for current platform
go build .

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -o conceal.exe .

# Cross-compile for macOS  
GOOS=darwin GOARCH=amd64 go build -o conceal-darwin .

# Cross-compile for macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o conceal-arm64 .
```

### Testing

Our CI/CD pipeline automatically tests on multiple platforms:

- **Unit Tests**: Run on Ubuntu, Windows, and macOS
- **Integration Tests**: Platform-specific credential store testing
- **Cross-Compilation**: Verify all target platforms build successfully
- **Security Scans**: Automated vulnerability detection

## Maintainer

[@infamousjoeg](https://github.com/infamousjoeg)

[![Buy me a coffee][buymeacoffee-shield]][buymeacoffee]

[buymeacoffee]: https://www.buymeacoffee.com/infamousjoeg
[buymeacoffee-shield]: https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png

## Contributions

Pull Requests are currently being accepted.  Please read and follow the guidelines laid out in [CONTRIBUTING.md]().

## License

[Apache 2.0](LICENSE)
