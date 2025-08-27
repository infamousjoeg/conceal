# Usage Examples

This document provides comprehensive examples of using Conceal across different platforms and scenarios.

## Basic Usage

### Storing Secrets

```bash
# Interactive secret entry (secure, recommended)
conceal set myapp/database/password

# Via pipeline (useful for scripts)
echo "mySecretPassword123" | conceal set myapp/database/password

# Multiple secrets for an application
conceal set myapp/api/key
conceal set myapp/database/url  
conceal set myapp/jwt/secret
```

### Retrieving Secrets

```bash
# Copy to clipboard (15-second auto-clear)
conceal get myapp/database/password

# Display to stdout (for Summon integration)
conceal show myapp/database/password

# List all stored secrets
conceal list

# Filter secrets by prefix
conceal list | grep myapp/
```

### Managing Secrets

```bash
# Update an existing secret
conceal update myapp/database/password

# Remove a secret
conceal unset myapp/database/password

# Check if a secret exists (via list)
conceal list | grep -q "myapp/api/key" && echo "exists" || echo "not found"
```

## Platform-Specific Examples

### macOS

```bash
# Secrets are stored in macOS Keychain
# View in Keychain Access.app under "Passwords" with service "summon"

# Set a secret (will appear in Keychain Access)
conceal set github/personal-token

# The secret will be visible in:
# Keychain Access > Passwords > summon (github/personal-token)
```

### Windows

```cmd
# Secrets are stored in Windows Credential Manager
# View in Control Panel > Credential Manager > Generic Credentials

# Set a secret (will appear in Credential Manager)
conceal set github/personal-token

# The secret will be visible in:
# Control Panel > Credential Manager > Generic Credentials > summon/github/personal-token
```

## Development Workflows

### Local Development Setup

```bash
#!/bin/bash
# setup-dev-secrets.sh - Initialize development secrets

echo "Setting up development secrets..."

# Database credentials
echo "Enter database password:"
conceal set myapp/dev/db/password

# API keys
echo "Enter GitHub API token:"
conceal set myapp/dev/github/token

echo "Enter Stripe test key:"
conceal set myapp/dev/stripe/key

# JWT signing key
openssl rand -base64 32 | conceal set myapp/dev/jwt/secret

echo "Development secrets configured!"
echo "Use 'conceal list' to verify"
```

### CI/CD Pipeline Secrets

```bash
# Store deployment keys
conceal set myapp/prod/deploy/key
conceal set myapp/staging/deploy/key

# Store service credentials  
conceal set myapp/prod/database/url
conceal set myapp/prod/redis/url
conceal set myapp/prod/s3/access-key
conceal set myapp/prod/s3/secret-key
```

### Team Secret Sharing

```bash
# Standardized secret naming for team projects
# Format: {project}/{environment}/{service}/{credential}

# Examples:
conceal set webapp/dev/postgres/password
conceal set webapp/dev/redis/password
conceal set webapp/staging/postgres/password
conceal set webapp/prod/postgres/password

conceal set api/dev/jwt/secret
conceal set api/staging/jwt/secret  
conceal set api/prod/jwt/secret
```

## Summon Integration

### Basic Summon Usage

1. **Install Conceal as Summon provider**:
   ```bash
   conceal summon install
   ```

2. **Create secrets.yml**:
   ```yaml
   # secrets.yml
   DATABASE_URL: !var myapp/prod/database/url
   API_KEY: !var myapp/prod/api/key
   JWT_SECRET: !var myapp/prod/jwt/secret
   ```

3. **Run application with secrets**:
   ```bash
   summon -p conceal_summon myapp
   ```

### Advanced Summon Configuration

```yaml
# secrets.yml - Advanced configuration
production:
  database:
    host: !var myapp/prod/db/host
    port: !var myapp/prod/db/port  
    username: !var myapp/prod/db/username
    password: !var myapp/prod/db/password
    name: !var myapp/prod/db/name
  
  external_apis:
    stripe_key: !var myapp/prod/stripe/secret
    github_token: !var myapp/prod/github/token
    sendgrid_key: !var myapp/prod/sendgrid/key
  
  security:
    jwt_secret: !var myapp/prod/jwt/secret
    encryption_key: !var myapp/prod/encryption/key

development:
  database:
    host: localhost
    port: 5432
    username: devuser
    password: !var myapp/dev/db/password
    name: myapp_dev
```

## Scripting Examples

### Backup Script

```bash
#!/bin/bash
# backup-secrets.sh - Export secret names (not values) for backup

echo "Backing up secret inventory..."
conceal list > secrets-inventory-$(date +%Y%m%d).txt
echo "Secret inventory saved to secrets-inventory-$(date +%Y%m%d).txt"
echo "Note: This contains secret names only, not values"
```

### Migration Script

```bash
#!/bin/bash  
# migrate-secrets.sh - Migrate secrets between environments

SOURCE_ENV="dev"
TARGET_ENV="staging"

echo "Migrating secrets from $SOURCE_ENV to $TARGET_ENV..."

# Get all dev secrets
conceal list | grep "myapp/$SOURCE_ENV/" | while read -r secret_name; do
    # Extract the service/credential part
    service_part=$(echo "$secret_name" | sed "s|myapp/$SOURCE_ENV/||")
    target_name="myapp/$TARGET_ENV/$service_part"
    
    echo "Migrating $secret_name -> $target_name"
    
    # Get the secret value and store in target
    conceal show "$secret_name" | conceal set "$target_name"
done

echo "Migration complete!"
```

### Audit Script

```bash
#!/bin/bash
# audit-secrets.sh - Audit secret usage and age

echo "Secret Audit Report - $(date)"
echo "=================================="

# List all secrets with timestamps (requires additional tooling for creation dates)
echo "Current secrets:"
conceal list | sort

echo -e "\nSecret count by environment:"
conceal list | cut -d'/' -f2 | sort | uniq -c

echo -e "\nSecret count by service:"  
conceal list | cut -d'/' -f3 | sort | uniq -c

echo -e "\nRecommendations:"
echo "- Rotate secrets older than 90 days"
echo "- Remove unused development secrets"
echo "- Ensure production secrets are properly backed up"
```

## Cross-Platform Automation

### PowerShell (Windows)

```powershell
# setup-windows-secrets.ps1
param(
    [string]$Environment = "dev"
)

Write-Host "Setting up $Environment secrets on Windows..."

# Function to set secret securely
function Set-ConcealSecret {
    param([string]$Name)
    
    $SecureString = Read-Host "Enter secret for $Name" -AsSecureString
    $PlainText = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
        [Runtime.InteropServices.Marshal]::SecureStringToBSTR($SecureString)
    )
    
    $PlainText | conceal set $Name
    [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($PlainText)
}

# Set secrets
Set-ConcealSecret "myapp/$Environment/database/password"
Set-ConcealSecret "myapp/$Environment/api/key"

Write-Host "Windows secrets configured!"
```

### Bash (macOS/Linux)

```bash
#!/bin/bash
# setup-unix-secrets.sh

ENVIRONMENT="${1:-dev}"

echo "Setting up $ENVIRONMENT secrets on Unix..."

# Function to set secret securely  
set_conceal_secret() {
    local name="$1"
    echo "Enter secret for $name:"
    read -s secret_value
    echo "$secret_value" | conceal set "$name"
    unset secret_value
}

# Set secrets
set_conceal_secret "myapp/$ENVIRONMENT/database/password"
set_conceal_secret "myapp/$ENVIRONMENT/api/key"

echo "Unix secrets configured!"
```

## Best Practices

### Secret Naming Convention

```bash
# Recommended format: {project}/{environment}/{service}/{credential}
# Examples:
conceal set ecommerce/prod/postgres/password
conceal set ecommerce/prod/redis/password  
conceal set ecommerce/staging/postgres/password
conceal set blog/dev/mysql/password
conceal set api/prod/jwt/signing-key
```

### Environment Separation

```bash
# Keep environments completely separate
conceal set myapp/dev/database/url     # Development
conceal set myapp/staging/database/url # Staging  
conceal set myapp/prod/database/url    # Production

# Never copy production secrets to lower environments
```

### Secret Rotation

```bash
#!/bin/bash
# rotate-secret.sh - Rotate a secret safely

SECRET_NAME="$1"
if [ -z "$SECRET_NAME" ]; then
    echo "Usage: $0 <secret_name>"
    exit 1
fi

echo "Rotating secret: $SECRET_NAME"

# Backup old secret (optional)
OLD_SECRET=$(conceal show "$SECRET_NAME")
echo "Old secret backed up"

# Set new secret
echo "Enter new secret value:"
conceal update "$SECRET_NAME"

echo "Secret rotated successfully!"
echo "Remember to update any applications using this secret"
```

## Troubleshooting Examples

### Debug Secret Access

```bash
# Check if secret exists
if conceal list | grep -q "myapp/prod/database/password"; then
    echo "Secret exists"
else
    echo "Secret not found"
fi

# Test secret retrieval
echo "Testing secret retrieval..."
conceal get myapp/test/dummy 2>&1 | grep -q "not found" && echo "Access working" || echo "Access issue"
```

### Platform-Specific Debugging

```bash
# macOS: Check Keychain Access
echo "Checking macOS Keychain..."
security find-generic-password -s "summon" -a "myapp/prod/database/password" 2>/dev/null && echo "Found in Keychain" || echo "Not in Keychain"

# Windows: Check Credential Manager (PowerShell)
# Get-StoredCredential -Target "summon/myapp/prod/database/password"
```

These examples demonstrate the flexibility and power of Conceal across different platforms and use cases. Adapt them to fit your specific needs and security requirements.