# Contributing to Conceal

Thank you for your interest in contributing to Conceal! This document provides guidelines for contributing to our cross-platform secret management tool.

## Table of Contents
- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Cross-Platform Development](#cross-platform-development)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Code of Conduct](#code-of-conduct)

## Getting Started

When contributing to this repository, please first discuss the change you wish to make via issue, email, or any other method with the maintainers before making a change.

### Prerequisites

- **Go**: Version 1.21 or later
- **Git**: For version control
- **Platform Access**: 
  - macOS with Xcode Command Line Tools (for Keychain testing)
  - Windows 10+ (for Credential Manager testing)

## Development Environment

### Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/infamousjoeg/conceal.git
   cd conceal
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Build for your platform**:
   ```bash
   go build .
   ```

4. **Run tests**:
   ```bash
   go test -v ./...
   ```

## Cross-Platform Development

Conceal supports multiple platforms with platform-specific implementations:

### Supported Platforms
- **macOS**: Full support via Keychain Access
- **Windows**: Full support via Credential Manager
- **Linux**: Not currently supported

### Build Tags

We use Go build tags to separate platform-specific code:

```go
//go:build windows
// +build windows

//go:build darwin  
// +build darwin

//go:build !darwin && !windows
// +build !darwin,!windows
```

### Testing Cross-Platform Code

#### Local Testing
```bash
# Test current platform
go test ./...

# Cross-compile (build only, won't run)
GOOS=windows GOARCH=amd64 go build .
GOOS=darwin GOARCH=amd64 go build .
```

#### CI/CD Testing
Our GitHub Actions automatically test on:
- Ubuntu (latest)
- Windows (latest)  
- macOS (latest)

### Adding New Platform Support

1. **Create platform-specific files**:
   ```
   pkg/conceal/keychain/keychain_newplatform.go
   ```

2. **Implement the interface**:
   ```go
   func SecretExists(secretID string) bool
   func ListSecrets() []QueryResult
   func AddSecret(secretID string, secret []byte) error
   func DeleteSecret(secretID string) error
   func GetSecret(secretID string, delivery string) error
   func UpdateSecret(secretID string, secret []byte) error
   ```

3. **Add build tags**:
   ```go
   //go:build newplatform
   // +build newplatform
   ```

4. **Update documentation** and platform support matrix

## Testing

### Running Tests

```bash
# All tests
go test -v ./...

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...
```

### Platform-Specific Testing

**macOS Testing:**
- Tests use actual Keychain Access APIs
- Requires user permission for keychain access
- Cannot fully automate in CI (requires user interaction)

**Windows Testing:**
- Tests use actual Credential Manager APIs  
- Requires Windows environment
- Cannot fully automate in CI (requires user interaction)

**Cross-Platform Testing:**
- Interface compatibility tests run on all platforms
- Build verification tests ensure compilation succeeds

### Adding Tests

1. **Unit tests**: Test individual functions
2. **Integration tests**: Test credential store interactions
3. **Platform compatibility tests**: Verify interfaces work across platforms

## Pull Request Process

### Before Submitting

1. **Run all tests**:
   ```bash
   go test ./...
   go test -race ./...
   ```

2. **Run linting**:
   ```bash
   golangci-lint run
   ```

3. **Format code**:
   ```bash
   go fmt ./...
   ```

4. **Check go.mod**:
   ```bash
   go mod tidy
   ```

### PR Requirements

1. **Code Quality**: 
   - All tests pass
   - Linting passes
   - Code is formatted with `go fmt`
   - `go mod tidy` has been run

2. **Documentation**:
   - Update README.md if adding new features
   - Add/update function documentation
   - Update platform support matrix if needed

3. **Version**: 
   - Bump version in `pkg/conceal/version.go` following [SemVer](http://semver.org/)
   - Major: Breaking changes
   - Minor: New features (like new platform support)
   - Patch: Bug fixes

4. **Cross-Platform**: 
   - Ensure changes work on all supported platforms
   - Add appropriate build tags
   - Test compilation on all platforms

### Review Process

1. **Automated Checks**: All CI/CD checks must pass
2. **Code Review**: Maintainer review required
3. **Testing**: Platform-specific testing as applicable
4. **Merge**: Squash and merge preferred

## Code Style

### Go Guidelines
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `go fmt` for formatting
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Platform-Specific Code
- Use build tags consistently
- Maintain consistent interfaces across platforms
- Handle platform differences gracefully
- Provide meaningful error messages for unsupported platforms

### Security
- Never log secrets or sensitive data
- Use secure random generation when needed
- Follow platform security best practices
- Validate all inputs

## Code of Conduct

### Our Pledge

In the interest of fostering an open and welcoming environment, we as
contributors and maintainers pledge to making participation in our project and
our community a harassment-free experience for everyone, regardless of age, body
size, disability, ethnicity, gender identity and expression, level of experience,
nationality, personal appearance, race, religion, or sexual identity and
orientation.

### Our Standards

Examples of behavior that contributes to creating a positive environment
include:

* Using welcoming and inclusive language
* Being respectful of differing viewpoints and experiences
* Gracefully accepting constructive criticism
* Focusing on what is best for the community
* Showing empathy towards other community members

Examples of unacceptable behavior by participants include:

* The use of sexualized language or imagery and unwelcome sexual attention or
advances
* Trolling, insulting/derogatory comments, and personal or political attacks
* Public or private harassment
* Publishing others' private information, such as a physical or electronic
  address, without explicit permission
* Other conduct which could reasonably be considered inappropriate in a
  professional setting

### Our Responsibilities

Project maintainers are responsible for clarifying the standards of acceptable
behavior and are expected to take appropriate and fair corrective action in
response to any instances of unacceptable behavior.

Project maintainers have the right and responsibility to remove, edit, or
reject comments, commits, code, wiki edits, issues, and other contributions
that are not aligned to this Code of Conduct, or to ban temporarily or
permanently any contributor for other behaviors that they deem inappropriate,
threatening, offensive, or harmful.

### Scope

This Code of Conduct applies both within project spaces and in public spaces
when an individual is representing the project or its community. Examples of
representing a project or community include using an official project e-mail
address, posting via an official social media account, or acting as an appointed
representative at an online or offline event. Representation of a project may be
further defined and clarified by project maintainers.

### Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be
reported by contacting the project team at joe dot garcia at cyberark dot com. All
complaints will be reviewed and investigated and will result in a response that
is deemed necessary and appropriate to the circumstances. The project team is
obligated to maintain confidentiality with regard to the reporter of an incident.
Further details of specific enforcement policies may be posted separately.

Project maintainers who do not follow or enforce the Code of Conduct in good
faith may face temporary or permanent repercussions as determined by other
members of the project's leadership.

### Attribution

This Code of Conduct is adapted from the [Contributor Covenant][homepage], version 1.4,
available at [http://contributor-covenant.org/version/1/4][version]

[homepage]: http://contributor-covenant.org
[version]: http://contributor-covenant.org/version/1/4/
