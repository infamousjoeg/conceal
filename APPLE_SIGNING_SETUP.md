# Apple Developer Code Signing & Notarization Setup

This document explains how to configure Apple Developer code signing and notarization for the Conceal project releases using CyberArk Conjur for secure secrets management.

## Prerequisites

1. **Apple Developer Account**: You need a paid Apple Developer account ($99/year)
2. **Developer ID Certificate**: Required for signing applications distributed outside the App Store
3. **App Store Connect API Key**: Required for notarization
4. **CyberArk Conjur Access**: Access to the CyberArk Conjur instance at `https://pineapple.secretsmgr.cyberark.cloud/api`

## Step 1: Generate Developer ID Certificates

### Developer ID Application Certificate (for binaries)
1. Log into [Apple Developer Portal](https://developer.apple.com/account/)
2. Go to **Certificates, Identifiers & Profiles** → **Certificates**
3. Click **+** to create a new certificate
4. Select **Developer ID Application** under "Software"
5. Follow the instructions to create a Certificate Signing Request (CSR)
6. Upload the CSR and download the certificate (.cer file)
7. Install the certificate in your Keychain Access
8. Export as `.p12` file with a strong password

### Developer ID Installer Certificate (for .pkg files)
1. In the same **Certificates** section
2. Click **+** to create another new certificate  
3. Select **Developer ID Installer** under "Software"
4. Use the same CSR or create a new one
5. Upload the CSR and download the installer certificate (.cer file)
6. Install the certificate in your Keychain Access
7. Export as `.p12` file with a strong password

## Step 2: Create App Store Connect API Key

1. Log into [App Store Connect](https://appstoreconnect.apple.com/)
2. Go to **Users and Access** → **Keys**
3. Click **+** to generate a new API Key
4. Enter a name like "Conceal Notarization"
5. Select **Developer** role
6. Click **Generate**
7. Download the `.p8` file (you can only download it once!)
8. Note the **Key ID** and **Issuer ID**

## Step 3: Prepare Certificate and Key Files

### Convert Certificate to Base64
```bash
# Convert your .p12 certificate to base64
base64 -i /path/to/your/certificate.p12 | pbcopy
```

### Convert API Key to Base64
```bash
# Convert your .p8 API key to base64  
base64 -i /path/to/your/AuthKey_XXXXXXXXXX.p8 | pbcopy
```

## Step 4: Store Secrets in CyberArk Conjur

You need to store the following secrets in CyberArk Conjur at the specified paths:

### Conjur Secret Paths

All secrets should be stored under: `data/infamousdevops/ci/github/infamousjoeg/conceal/`

| Variable Name | Conjur Secret Path | Description | Example Value |
|---------------|-------------------|-------------|---------------|
| `MACOS_SIGN_P12` | `data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_SIGN_P12` | Base64 encoded Developer ID Application .p12 | `MIIKuw...` (very long base64 string) |
| `MACOS_SIGN_PASSWORD` | `data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_SIGN_PASSWORD` | Password for the Application .p12 certificate | `your-app-certificate-password` |
| `MACOS_SIGN_IDENTITY` | `data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_SIGN_IDENTITY` | Developer ID Application identity | `Developer ID Application: Your Name (TEAM123456)` |
| `MACOS_INSTALLER_P12` | `data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_INSTALLER_P12` | Base64 encoded Developer ID Installer .p12 | `MIIKuw...` (very long base64 string) |
| `MACOS_INSTALLER_PASSWORD` | `data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_INSTALLER_PASSWORD` | Password for the Installer .p12 certificate | `your-installer-certificate-password` |
| `MACOS_INSTALLER_IDENTITY` | `data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_INSTALLER_IDENTITY` | Developer ID Installer identity | `Developer ID Installer: Your Name (TEAM123456)` |
| `MACOS_NOTARY_ISSUER_ID` | `data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_NOTARY_ISSUER_ID` | App Store Connect Issuer ID | `12345678-1234-1234-1234-123456789012` |
| `MACOS_NOTARY_KEY_ID` | `data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_NOTARY_KEY_ID` | App Store Connect API Key ID | `ABC123DEF4` |
| `MACOS_NOTARY_KEY` | `data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_NOTARY_KEY` | Base64 encoded .p8 API key | `LS0tLS1C...` (base64 string) |
| `KEYCHAIN_PASSWORD` | `data/infamousdevops/ci/github/infamousjoeg/conceal/KEYCHAIN_PASSWORD` | Temporary keychain password | `temp-signing-password` ✅ **Auto-generated** |

### Storing Secrets in Conjur

Use the Conjur CLI or web interface to store these secrets:

```bash
# Example using Conjur CLI
conjur variable value add data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_SIGN_P12 "$(base64 -i /path/to/certificate.p12)"
conjur variable value add data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_SIGN_PASSWORD "your-certificate-password"
# ... continue for all other secrets
```

## How to Find Your Values

### Finding Your Developer ID Identities
```bash
# List all signing identities in your keychain
security find-identity -v -p codesigning

# Look for something like:
# 1) 1234567890ABCDEF1234567890ABCDEF12345678 "Developer ID Application: Your Name (TEAM123456)"
# 2) 0987654321FEDCBA0987654321FEDCBA09876543 "Developer ID Installer: Your Name (TEAM123456)"
```

### Distinguishing Between Certificate Types
- **Developer ID Application**: Used for signing binaries (executables)
- **Developer ID Installer**: Used for signing .pkg installer packages

### Finding Your Issuer ID
1. Go to App Store Connect → Users and Access → Keys
2. The **Issuer ID** is displayed at the top of the page

### Finding Your Key ID
1. In App Store Connect → Users and Access → Keys  
2. Click on your API key
3. The **Key ID** is shown in the key details

## Testing the Setup

After storing all secrets in Conjur, you can test by:

1. **Verify Conjur Access**: Ensure the GitHub repository has proper access to fetch secrets from Conjur
2. **Create a test tag**: `git tag v1.0.0-test && git push origin v1.0.0-test`
3. **Monitor the workflow**: Check the Actions tab to see if:
   - Secrets are successfully fetched from Conjur
   - Code signing completes without errors
   - Notarization succeeds
4. **Clean up**: Delete the test tag and release when done

### Conjur Authentication

The workflow uses JWT authentication with the following configuration:
- **Conjur URL**: `https://pineapple.secretsmgr.cyberark.cloud/api`
- **Service ID**: `inf-github`
- **Account**: `cyberark`
- **JWT Token**: Uses `GITHUB_TOKEN` for authentication

## Troubleshooting

### Common Issues

1. **"Failed to authenticate with Conjur"**
   - Verify the GitHub repository is properly configured in Conjur
   - Check that the Service ID `inf-github` has access to the secret paths
   - Ensure the JWT authentication is properly set up

2. **"Secret not found in Conjur"**
   - Verify all secret paths are correct: `data/infamousdevops/ci/github/infamousjoeg/conceal/VARIABLE_NAME`
   - Check that all secrets are properly stored in Conjur
   - Ensure the variable names match exactly (case-sensitive)

3. **"No signing identity found"**
   - Verify your `MACOS_SIGN_IDENTITY` matches exactly what's in your keychain
   - Check that the certificate is properly installed
   - Ensure the base64-encoded certificate in Conjur is correct

4. **"Invalid API key"**
   - Verify the API key is valid and has Developer role
   - Check that the Key ID and Issuer ID are correct
   - Ensure the base64-encoded API key in Conjur is correct

5. **"Notarization timeout"**
   - Increase timeout in `.goreleaser.yml` if needed
   - Check Apple's system status for service issues

6. **"Conjur connection timeout"**
   - Verify network connectivity to `https://pineapple.secretsmgr.cyberark.cloud/api`
   - Check if there are any firewall restrictions in GitHub Actions

### Checking Your Certificate
```bash
# Verify your certificate is in keychain
security find-identity -v -p codesigning | grep "Developer ID"

# Check certificate details
security find-certificate -c "Developer ID Application" -p | openssl x509 -text -noout
```

### Debugging Notarization
```bash
# Check notarization history (requires xcrun)
xcrun notarytool history --issuer-id YOUR_ISSUER_ID --key-id YOUR_KEY_ID --key /path/to/key.p8

# Get notarization info for a specific submission
xcrun notarytool info SUBMISSION_ID --issuer-id YOUR_ISSUER_ID --key-id YOUR_KEY_ID --key /path/to/key.p8
```

## Security Best Practices

1. **Rotate Keys Regularly**: Update API keys periodically
2. **Limit Access**: Only give secrets access to necessary people
3. **Monitor Usage**: Check for unexpected signing activity
4. **Backup Certificates**: Keep secure backups of your certificates
5. **Use Strong Passwords**: Use complex passwords for certificate files

### Debugging Conjur Integration

```bash
# Test Conjur connectivity (if you have access)
conjur authn authenticate -a cyberark
conjur variable value data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_SIGN_IDENTITY

# Check if all required variables exist
conjur variable list | grep "data/infamousdevops/ci/github/infamousjoeg/conceal/"
```

## Manual Steps Required

You'll need to manually add values for these secrets using the Conjur CLI:

### **Apple Developer Application Certificate (for binaries)**:
```bash
# Export your Developer ID Application certificate as .p12, then:
conjur variable set -i data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_SIGN_P12 -v "$(base64 -i /path/to/your/application_certificate.p12)"
conjur variable set -i data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_SIGN_PASSWORD -v "your-application-certificate-password"
conjur variable set -i data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_SIGN_IDENTITY -v "Developer ID Application: Your Name (TEAM123456)"
```

### **Apple Developer Installer Certificate (for .pkg files)**:
```bash
# Export your Developer ID Installer certificate as .p12, then:
conjur variable set -i data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_INSTALLER_P12 -v "$(base64 -i /path/to/your/installer_certificate.p12)"
conjur variable set -i data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_INSTALLER_PASSWORD -v "your-installer-certificate-password"
conjur variable set -i data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_INSTALLER_IDENTITY -v "Developer ID Installer: Your Name (TEAM123456)"
```

### **App Store Connect API Details**:
```bash
# From App Store Connect → Users and Access → Keys
conjur variable set -i data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_NOTARY_ISSUER_ID -v "12345678-1234-1234-1234-123456789012"
conjur variable set -i data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_NOTARY_KEY_ID -v "ABC123DEF4"

# Convert your .p8 API key to base64:
conjur variable set -i data/infamousdevops/ci/github/infamousjoeg/conceal/MACOS_NOTARY_KEY -v "$(base64 -i /path/to/your/AuthKey_XXXXXXXXXX.p8)"
```

## Support

If you encounter issues:

1. **Conjur Issues**: 
   - Check Conjur logs and authentication setup
   - Verify the Service ID `inf-github` has proper permissions
   - Ensure all secret paths are correct and accessible

2. **Apple Developer Issues**:
   - Check the [Apple Developer Documentation](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
   - Verify certificates and API keys are valid

3. **GitHub Actions Issues**:
   - Review GitHub Actions logs for specific error messages  
   - Check if secrets are being fetched successfully from Conjur
   - Test locally with `goreleaser build --single-target` first

4. **Authentication Flow**:
   - The workflow uses GitHub's OIDC token to authenticate with Conjur
   - Ensure the JWT authentication is configured correctly in Conjur