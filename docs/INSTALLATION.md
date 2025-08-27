# Installation Guide

This guide provides detailed installation instructions for Conceal across all supported platforms.

## Supported Platforms

| Platform | Version | Architecture | Status |
|----------|---------|-------------|--------|
| macOS | 10.12+ | Intel (x86_64) | ✅ |
| macOS | 11.0+ | Apple Silicon (arm64) | ✅ |
| Windows | 10+ | x86_64 | ✅ |
| Windows | Server 2016+ | x86_64 | ✅ |

## macOS Installation

### Option 1: Homebrew (Recommended)

```bash
# Add the Conceal tap
brew tap infamousjoeg/tap

# Install Conceal
brew install conceal

# Verify installation
conceal version
```

### Option 2: Manual Installation

1. **Download the binary**:
   - Intel Macs: Download `conceal_Darwin_x86_64.tar.gz`
   - Apple Silicon Macs: Download `conceal_Darwin_arm64.tar.gz`
   - From: https://github.com/infamousjoeg/conceal/releases/latest

2. **Extract and install**:
   ```bash
   # Extract (replace with your downloaded file)
   tar -xzf conceal_Darwin_x86_64.tar.gz
   
   # Move to PATH directory
   sudo mv conceal /usr/local/bin/
   
   # Make executable (if needed)
   chmod +x /usr/local/bin/conceal
   ```

3. **Verify installation**:
   ```bash
   conceal version
   ```

### Option 3: Build from Source

```bash
# Install Go (if not already installed)
brew install go

# Clone and build
git clone https://github.com/infamousjoeg/conceal.git
cd conceal
go build -o conceal .

# Move to PATH
sudo mv conceal /usr/local/bin/
```

## Windows Installation

### Option 1: Manual Installation (Current)

1. **Download the binary**:
   - Download `conceal_Windows_x86_64.zip` from [GitHub Releases](https://github.com/infamousjoeg/conceal/releases/latest)

2. **Extract and install**:
   - Extract `conceal.exe` from the ZIP file
   - Choose an installation directory:
     - **System-wide**: `C:\Program Files\Conceal\`
     - **User-specific**: `%USERPROFILE%\AppData\Local\Programs\Conceal\`

3. **Add to PATH**:
   
   **Via PowerShell (Recommended)**:
   ```powershell
   # For system-wide installation
   $env:PATH += ";C:\Program Files\Conceal"
   [Environment]::SetEnvironmentVariable("Path", $env:PATH, [EnvironmentVariableTarget]::Machine)
   
   # For user-specific installation
   $env:PATH += ";$env:USERPROFILE\AppData\Local\Programs\Conceal"
   [Environment]::SetEnvironmentVariable("Path", $env:PATH, [EnvironmentVariableTarget]::User)
   ```
   
   **Via GUI**:
   - Open "Environment Variables" from System Properties
   - Add the installation directory to your PATH
   - Restart your terminal

4. **Verify installation**:
   ```cmd
   conceal version
   ```

### Option 2: Chocolatey (Future)

```powershell
# Coming soon
# choco install conceal
```

### Option 3: Build from Source

1. **Install Go**:
   - Download from https://golang.org/dl/
   - Or via Chocolatey: `choco install golang`

2. **Build**:
   ```cmd
   git clone https://github.com/infamousjoeg/conceal.git
   cd conceal
   go build -o conceal.exe .
   ```

3. **Install** following steps from Option 1

## Verification

After installation, verify Conceal is working correctly:

```bash
# Check version
conceal version

# View help
conceal help

# Test basic functionality (this will prompt for platform support)
conceal list
```

## Troubleshooting

### macOS Issues

**"conceal cannot be opened because the developer cannot be verified"**
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /usr/local/bin/conceal

# Or allow in System Preferences > Security & Privacy
```

**Keychain access denied**
- Conceal may prompt for keychain access on first use
- Enter your macOS password when prompted
- Check "Always Allow" to avoid future prompts

### Windows Issues

**"conceal is not recognized as an internal or external command"**
- Verify the executable is in your PATH
- Restart your terminal after adding to PATH
- Check that you downloaded the Windows version (.exe)

**Credential Manager access denied**
- Run as Administrator if needed
- Check Windows User Account Control (UAC) settings
- Verify you have permission to access Credential Manager

**Antivirus false positive**
- Some antivirus software may flag the binary
- Add conceal.exe to your antivirus exclusions
- Download only from official GitHub releases

### General Issues

**Network connectivity errors**
- Check your internet connection for releases download
- Use corporate proxy settings if applicable
- Try downloading from a different network

**Permission errors**
- Ensure you have write access to the installation directory
- Use `sudo` on macOS/Linux or "Run as Administrator" on Windows
- Check filesystem permissions

## Uninstallation

### macOS
```bash
# If installed via Homebrew
brew uninstall conceal
brew untap infamousjoeg/tap

# If installed manually
sudo rm /usr/local/bin/conceal
```

### Windows
```cmd
# Remove from PATH (via System Properties > Environment Variables)
# Delete the installation directory
del "C:\Program Files\Conceal\conceal.exe"
```

## Next Steps

After successful installation:

1. **First Use**: Run `conceal help` to see available commands
2. **Store a Secret**: Try `conceal set test/secret`
3. **Retrieve a Secret**: Try `conceal get test/secret`
4. **Integration**: See our [Summon integration guide](https://cyberark.github.io/summon)

## Getting Help

- **Documentation**: https://github.com/infamousjoeg/conceal
- **Issues**: https://github.com/infamousjoeg/conceal/issues
- **Discussions**: https://github.com/infamousjoeg/conceal/discussions