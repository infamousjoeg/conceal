# Security Policy

## Supported Versions

We take security seriously and provide security updates for the following versions:

| Version | Supported          | End of Life |
| ------- | ------------------ | ----------- |
| 4.1.x   | ✅ Full support    | N/A         |
| 4.0.x   | ✅ Security fixes  | 2025-12-31  |
| < 4.0   | ❌ Not supported   | 2024-01-01  |

## Security Features

### Platform-Specific Security

**macOS Keychain:**
- Leverages macOS Security Framework
- Encrypted storage using hardware encryption when available
- Respects keychain access policies and user permissions
- Integrates with Touch ID/Face ID authentication when configured

**Windows Credential Manager:**
- Uses Windows Credential Manager APIs (advapi32.dll)
- Encrypted storage using Windows Data Protection API (DPAPI)
- Respects Windows user access controls and group policies
- Integrates with Windows Hello when available

### Application Security

- **No Secret Logging**: Secrets are never written to logs or console output (except via explicit `show` command)
- **Memory Safety**: Secrets are cleared from memory when possible
- **Clipboard Security**: 15-second auto-clear with interrupt handling
- **Input Validation**: All user inputs are validated and sanitized
- **Build Security**: Reproducible builds with dependency verification

## Automated Security

### Continuous Security Monitoring

Our CI/CD pipeline includes:

- **Gosec**: Static security analysis for Go code
- **Nancy**: Dependency vulnerability scanning  
- **govulncheck**: Official Go vulnerability database checking
- **Dependabot**: Automated dependency updates
- **CodeQL**: Advanced semantic code analysis (planned)

### Security Workflows

- **Weekly Scans**: Automated security scans every Monday
- **PR Security**: All pull requests undergo security analysis
- **Dependency Updates**: Automatic updates for security patches
- **Release Security**: Security verification before each release

## Reporting a Vulnerability

### Responsible Disclosure

We appreciate security researchers who report vulnerabilities responsibly. Please follow our responsible disclosure process:

1. **Contact**: Email `joe dot garcia at cyberark dot com` directly
2. **Include**:
   - Detailed description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact assessment
   - Your contact information for follow-up

### What to Expect

| Timeline | Action |
|----------|---------|
| 24 hours | Acknowledgment of your report |
| 48 hours | Initial assessment and classification |
| 7 days   | Detailed investigation and response plan |
| 30 days  | Fix implementation and testing |
| Release  | Public disclosure and credit (if desired) |

### Security Response

**Critical Vulnerabilities:**
- Immediate investigation
- Emergency patch within 48-72 hours
- Public security advisory

**High/Medium Vulnerabilities:**
- Investigation within 7 days
- Patch in next scheduled release
- Security note in release notes

**Low Vulnerabilities:**
- Investigation within 30 days
- Fix in future release
- Standard development process

## Security Best Practices

### For Users

1. **Keep Updated**: Always use the latest version
2. **Secure Environment**: Use Conceal only on trusted systems
3. **Access Control**: Limit who has access to your credential stores
4. **Regular Audits**: Periodically review stored secrets with `conceal list`
5. **Backup Strategy**: Ensure your OS credential store is backed up securely

### For Developers

1. **Code Review**: All code changes require security-focused review
2. **Dependency Management**: Keep dependencies updated and vetted
3. **Testing**: Include security test cases in your contributions
4. **Documentation**: Document any security-relevant changes
5. **Build Tags**: Use appropriate build tags to limit platform exposure

## Threat Model

### Assets
- User secrets stored in OS credential stores
- Application source code and binaries
- User systems and credential stores

### Threats
- **Malicious Code Injection**: Mitigated by code review and static analysis
- **Dependency Vulnerabilities**: Mitigated by automated scanning and updates
- **Memory Extraction**: Mitigated by secure memory practices
- **Privilege Escalation**: Mitigated by using OS-native security boundaries
- **Supply Chain Attacks**: Mitigated by reproducible builds and verification

### Assumptions
- User's operating system and credential store are secure
- User has legitimate access to their credential store
- Development and build environments are secure
- Go toolchain and standard library are secure

## Security Contact

For security-related questions or concerns that don't constitute vulnerabilities:

- **Email**: joe dot garcia at cyberark dot com  
- **Subject**: [Conceal Security] Your Question
- **Response Time**: 7-14 business days

For urgent security matters, please include "URGENT" in the subject line.
