# Authentication Troubleshooting Guide

## Overview

This guide provides solutions to common authentication issues you may encounter with AINative Code. Issues are organized by category with symptoms, causes, and step-by-step solutions.

## Table of Contents

- [Login Issues](#login-issues)
- [Token Expiration](#token-expiration)
- [Network Problems](#network-problems)
- [Keychain Access Issues](#keychain-access-issues)
- [Browser Issues](#browser-issues)
- [Callback Server Issues](#callback-server-issues)
- [Token Validation Errors](#token-validation-errors)
- [Platform-Specific Issues](#platform-specific-issues)

---

## Login Issues

### Browser Doesn't Open Automatically

**Symptoms**:
```
Initiating authentication flow...
Opening browser for authentication...
Error: failed to open browser
```

**Causes**:
- No default browser configured
- Browser executable not in PATH
- Running in headless environment (SSH, Docker)
- Browser permission issues

**Solutions**:

**1. Manual Browser Opening**:
```bash
# CLI will print the authorization URL
ainative-code auth login

# Copy the URL printed in terminal
# Example: https://auth.ainative.studio/oauth/authorize?...

# Paste into browser manually
```

**2. Set Default Browser** (macOS):
```bash
# Check current default browser
defaults read com.apple.LaunchServices/com.apple.launchservices.secure LSHandlers

# Set Chrome as default
open -a "Google Chrome" --args --make-default-browser
```

**3. Set Default Browser** (Linux):
```bash
# Set Firefox as default
xdg-settings set default-web-browser firefox.desktop

# Or Chrome
xdg-settings set default-web-browser google-chrome.desktop

# Verify
xdg-settings get default-web-browser
```

**4. SSH/Remote Environment**:
```bash
# Use port forwarding to access callback server
ssh -L 8080:localhost:8080 user@remote-host

# On remote host
ainative-code auth login

# On local machine, open the authorization URL in browser
```

### Authorization Denied

**Symptoms**:
```
Error: user denied authorization
Please run 'ainative-code auth login' again to authenticate
```

**Causes**:
- User clicked "Deny" or "Cancel" in browser
- User closed browser window before completing authorization

**Solutions**:

**1. Retry Login**:
```bash
ainative-code auth login
# Click "Authorize" this time
```

**2. Check Required Scopes**:
```bash
# Ensure you're comfortable with requested scopes
ainative-code auth login --scopes read,write,offline_access
```

### Invalid Credentials

**Symptoms**:
```
Login failed: Invalid email or password
```

**Causes**:
- Incorrect email or password
- Account doesn't exist
- Account suspended

**Solutions**:

**1. Reset Password**:
- Visit https://ainative.studio/forgot-password
- Follow password reset instructions

**2. Create Account**:
- Visit https://ainative.studio/signup
- Create new account
- Verify email address

**3. Contact Support**:
- Email: support@ainative.studio
- Include: Account email, error message

---

## Token Expiration

### Access Token Expired

**Symptoms**:
```
Error: token has expired
```

**Causes**:
- Access token lifetime exceeded (> 1 hour)
- System clock incorrect
- Token manually deleted

**Solutions**:

**1. Automatic Refresh** (Recommended):
```bash
# CLI automatically refreshes tokens
# Just retry the failed command
ainative-code zerodb query "SELECT * FROM users"
```

**2. Manual Refresh**:
```bash
ainative-code auth token refresh
```

**3. Re-authenticate if Refresh Fails**:
```bash
ainative-code auth login
```

**4. Check System Clock**:
```bash
# macOS
date

# Sync time
sudo sntp -sS time.apple.com

# Linux
date
sudo ntpdate pool.ntp.org

# Windows (PowerShell as Admin)
w32tm /resync
```

### Refresh Token Expired

**Symptoms**:
```
Error: refresh token has expired
Run 'ainative-code auth login' to authenticate
```

**Causes**:
- No activity for 30 days
- Refresh token revoked
- Session terminated on server

**Solutions**:

**1. Re-authenticate**:
```bash
ainative-code auth login
```

**2. Enable Auto-Refresh** (Prevents This):
```yaml
# ~/.config/ainative-code/config.yaml
platform:
  keychain:
    auto_refresh: true
    refresh_threshold: 5m
```

### Token Expiring Soon Warning

**Symptoms**:
```
Authenticated User:
  Email: user@example.com
  Expires In: 2m 15s
  ⚠️  Token expiring soon!
```

**Cause**:
- Token will expire within 5 minutes

**Solutions**:

**1. Refresh Proactively**:
```bash
ainative-code auth token refresh
```

**2. Increase Refresh Threshold**:
```yaml
# ~/.config/ainative-code/config.yaml
platform:
  keychain:
    refresh_threshold: 10m  # Refresh when < 10 min remaining
```

---

## Network Problems

### Connection Timeout

**Symptoms**:
```
Error: HTTP request timeout
Failed to connect to auth.ainative.studio
```

**Causes**:
- Network connectivity issues
- Firewall blocking requests
- Proxy configuration required
- Server downtime

**Solutions**:

**1. Check Internet Connection**:
```bash
# Test basic connectivity
ping auth.ainative.studio

# Test HTTPS access
curl -I https://auth.ainative.studio
```

**2. Configure Proxy**:
```bash
# Set proxy environment variables
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
export NO_PROXY=localhost,127.0.0.1

# Or in config
```

```yaml
# ~/.config/ainative-code/config.yaml
http:
  proxy: "http://proxy.example.com:8080"
  timeout: 60s
```

**3. Check Firewall**:
```bash
# macOS - Allow in Security & Privacy
# Linux - Configure iptables
sudo iptables -A OUTPUT -p tcp --dport 443 -j ACCEPT

# Windows - Add firewall rule
netsh advfirewall firewall add rule name="AINative Code" dir=out action=allow protocol=TCP localport=443
```

**4. Increase Timeout**:
```bash
ainative-code auth login --timeout 60s
```

### SSL/TLS Errors

**Symptoms**:
```
Error: x509: certificate signed by unknown authority
```

**Causes**:
- Corporate SSL inspection
- Outdated root certificates
- Custom CA certificates

**Solutions**:

**1. Update Root Certificates**:
```bash
# macOS
sudo /System/Library/CoreServices/Install\ in\ Progress.app/Contents/Resources/SecurityUpdate

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install ca-certificates

# CentOS/RHEL
sudo yum update ca-certificates
```

**2. Add Corporate CA Certificate**:
```bash
# macOS
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain /path/to/ca.crt

# Linux
sudo cp /path/to/ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# Windows
certutil -addstore -f "ROOT" C:\path\to\ca.crt
```

**3. Disable SSL Verification** (Not Recommended for Production):
```yaml
# ~/.config/ainative-code/config.yaml
http:
  insecure_skip_verify: true  # USE WITH CAUTION
```

### Rate Limiting

**Symptoms**:
```
Error: HTTP 429 Too Many Requests
Retry after: 60 seconds
```

**Cause**:
- Too many authentication attempts
- API rate limit exceeded

**Solutions**:

**1. Wait and Retry**:
```bash
# Wait for rate limit to reset (usually 1 minute)
sleep 60
ainative-code auth login
```

**2. Check for Multiple Processes**:
```bash
# Kill duplicate processes
ps aux | grep ainative-code
kill <PID>
```

---

## Keychain Access Issues

### macOS: Keychain Access Denied

**Symptoms**:
```
Error: keychain access denied
Failed to store tokens in keychain
```

**Causes**:
- Keychain locked
- Permission issues
- First-time access prompt dismissed

**Solutions**:

**1. Unlock Keychain**:
```bash
# Unlock login keychain
security unlock-keychain ~/Library/Keychains/login.keychain-db
# Enter password when prompted
```

**2. Grant Access in Keychain App**:
1. Open "Keychain Access" app
2. Search for "ainative-code"
3. Right-click entry → "Get Info"
4. Go to "Access Control" tab
5. Click "+" and add `/usr/local/bin/ainative-code`
6. Save changes

**3. Reset Keychain Permissions**:
```bash
# Delete existing entry
security delete-generic-password -s "ainative-code" ~/Library/Keychains/login.keychain-db

# Re-authenticate
ainative-code auth login
# Grant access when prompted
```

**4. Create New Keychain**:
```bash
# Create dedicated keychain
security create-keychain ainative.keychain
security unlock-keychain ainative.keychain
security set-keychain-settings ainative.keychain

# Configure CLI to use it
export AINATIVE_KEYCHAIN_PATH="~/Library/Keychains/ainative.keychain"
```

### Linux: Secret Service Not Available

**Symptoms**:
```
Error: failed to connect to secret service
Keychain access unavailable
```

**Causes**:
- gnome-keyring not running
- dbus session not available
- Running in headless environment

**Solutions**:

**1. Start gnome-keyring**:
```bash
# Check if running
ps aux | grep gnome-keyring

# Start if not running
eval $(gnome-keyring-daemon --start)
export $(gnome-keyring-daemon --start)
```

**2. Install gnome-keyring** (if missing):
```bash
# Ubuntu/Debian
sudo apt-get install gnome-keyring

# Fedora
sudo dnf install gnome-keyring

# Arch
sudo pacman -S gnome-keyring
```

**3. Use Alternative Storage** (Fallback):
```yaml
# ~/.config/ainative-code/config.yaml
platform:
  keychain:
    backend: "file"  # Use encrypted file instead
    file_path: "~/.config/ainative-code/tokens.enc"
```

### Windows: Credential Manager Issues

**Symptoms**:
```
Error: failed to access Windows Credential Manager
```

**Causes**:
- Credential Manager service stopped
- Permission issues
- Credential vault corrupted

**Solutions**:

**1. Check Credential Manager Service**:
```powershell
# Open Services (services.msc)
# Find "Credential Manager" service
# Ensure status is "Running"

# Or via PowerShell
Get-Service -Name VaultSvc | Start-Service
```

**2. Clear Corrupted Credentials**:
```powershell
# Open Credential Manager
# Control Panel → User Accounts → Credential Manager
# Delete "ainative-code" entries
# Re-authenticate
```

**3. Run as Administrator**:
```powershell
# Right-click Command Prompt
# Select "Run as administrator"
ainative-code auth login
```

---

## Browser Issues

### Callback Not Received

**Symptoms**:
```
Opening browser for authentication...
Waiting for callback...
Error: callback timeout (5 minutes elapsed)
```

**Causes**:
- Browser blocked popup
- Browser closed before completion
- Callback URL not opened
- Firewall blocking localhost:8080

**Solutions**:

**1. Check Browser Console**:
- Open browser Developer Tools (F12)
- Check Console tab for errors
- Look for redirect errors

**2. Manually Copy Callback URL**:
```bash
# After authorizing in browser, copy the callback URL
# http://localhost:8080/callback?code=...&state=...

# Paste it in terminal if CLI provides prompt
```

**3. Check Popup Blocker**:
- Disable popup blocker for auth.ainative.studio
- Add to allowed sites

**4. Use Different Browser**:
```bash
# Set default browser to one that works
export BROWSER=firefox
ainative-code auth login
```

### Multiple Browser Windows

**Symptoms**:
- Multiple browser tabs/windows open
- Confusion about which to use

**Cause**:
- Multiple login attempts
- Browser session restore

**Solutions**:

**1. Close Extra Windows**:
- Close all authentication browser windows
- Restart authentication with single window

**2. Clear Browser Cache**:
```bash
# Chrome
# Settings → Privacy → Clear browsing data

# Firefox
# Preferences → Privacy → Clear Data
```

---

## Callback Server Issues

### Port Already in Use

**Symptoms**:
```
Error: failed to start callback server
listen tcp :8080: bind: address already in use
```

**Cause**:
- Another process using port 8080
- Previous CLI instance still running

**Solutions**:

**1. Find and Kill Process**:
```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or killall
killall -9 ainative-code
```

**2. Use Alternative Port**:
```yaml
# ~/.config/ainative-code/config.yaml
platform:
  oauth:
    redirect_uri: "http://localhost:8081/callback"
    callback_port: 8081
```

```bash
# Then update server configuration to allow this redirect URI
```

**3. Wait for Port Release**:
```bash
# Ports are released after TIME_WAIT (usually 60 seconds)
sleep 60
ainative-code auth login
```

### Callback Server Won't Start

**Symptoms**:
```
Error: callback server start failed
Permission denied
```

**Causes**:
- Port requires root privileges (< 1024)
- Firewall blocking localhost
- SELinux/AppArmor restrictions

**Solutions**:

**1. Use Higher Port**:
```yaml
# Use port > 1024 (doesn't require root)
platform:
  oauth:
    callback_port: 8080  # Default, no special privileges needed
```

**2. Configure Firewall**:
```bash
# macOS - Allow in Firewall preferences
# Linux iptables
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT

# UFW
sudo ufw allow 8080/tcp
```

**3. SELinux Context** (Linux):
```bash
# Add SELinux rule
sudo semanage port -a -t http_port_t -p tcp 8080
```

---

## Token Validation Errors

### Invalid Signature

**Symptoms**:
```
Error: invalid token signature
Token verification failed
```

**Causes**:
- Token tampered with
- Wrong public key used for verification
- Token from different environment (dev vs prod)

**Solutions**:

**1. Re-authenticate**:
```bash
# Get fresh tokens
ainative-code auth logout
ainative-code auth login
```

**2. Update Public Key**:
```bash
# Download latest public key
curl -o ~/.ainative/jwt-public.pem \
  https://auth.ainative.studio/.well-known/jwks.json

# Or configure path
```

```yaml
# ~/.config/ainative-code/config.yaml
platform:
  jwt:
    public_key_path: "~/.ainative/jwt-public.pem"
```

**3. Check Environment**:
```bash
# Ensure using correct environment
export AINATIVE_ENV=production
ainative-code auth login
```

### Invalid Claims

**Symptoms**:
```
Error: invalid token claims
Missing required claim: email
```

**Causes**:
- Incomplete token from server
- Token from old version
- Custom authentication server issue

**Solutions**:

**1. Check Token Contents**:
```bash
# Decode JWT (doesn't verify)
echo "eyJhbG..." | base64 -d | jq .

# Look for missing claims
```

**2. Re-authenticate**:
```bash
ainative-code auth logout
ainative-code auth login
```

**3. Verify Scopes**:
```bash
# Ensure all required scopes requested
ainative-code auth login --scopes read,write,offline_access,profile,email
```

### Wrong Issuer or Audience

**Symptoms**:
```
Error: invalid token issuer
Expected 'ainative-auth', got 'other-issuer'
```

**Cause**:
- Token from wrong authentication server
- Configuration mismatch

**Solutions**:

**1. Check Configuration**:
```yaml
# ~/.config/ainative-code/config.yaml
platform:
  jwt:
    issuer: "ainative-auth"      # Must match token
    audience: "ainative-code"    # Must match token
```

**2. Verify Endpoints**:
```bash
# Ensure using correct auth server
ainative-code auth login \
  --auth-url https://auth.ainative.studio/oauth/authorize \
  --token-url https://auth.ainative.studio/oauth/token
```

---

## Platform-Specific Issues

### macOS: Code Signing Issues

**Symptoms**:
```
"ainative-code" cannot be opened because the developer cannot be verified
```

**Solutions**:

**1. Allow App**:
```bash
# System Preferences → Security & Privacy
# Click "Open Anyway" button

# Or via command
sudo spctl --master-disable
```

**2. Remove Quarantine**:
```bash
sudo xattr -r -d com.apple.quarantine /usr/local/bin/ainative-code
```

### Linux: AppArmor Restrictions

**Symptoms**:
```
Error: permission denied
AppArmor restriction
```

**Solutions**:

**1. Check AppArmor Profile**:
```bash
sudo aa-status | grep ainative
```

**2. Create AppArmor Profile**:
```bash
# /etc/apparmor.d/usr.local.bin.ainative-code
/usr/local/bin/ainative-code {
  #include <abstractions/base>
  #include <abstractions/nameservice>

  /usr/local/bin/ainative-code mr,
  /home/*/.config/ainative-code/** rw,
  /home/*/.ainative/** rw,

  network inet stream,
  network inet6 stream,
}

# Load profile
sudo apparmor_parser -r /etc/apparmor.d/usr.local.bin.ainative-code
```

### Windows: Antivirus Blocking

**Symptoms**:
- Antivirus blocks executable
- Network requests blocked

**Solutions**:

**1. Add Exception**:
- Open antivirus software
- Add `ainative-code.exe` to exceptions
- Add auth.ainative.studio to trusted sites

**2. Windows Defender**:
```powershell
# Add folder exclusion
Add-MpPreference -ExclusionPath "C:\Program Files\ainative-code"

# Add process exclusion
Add-MpPreference -ExclusionProcess "ainative-code.exe"
```

---

## Debug Mode

### Enable Detailed Logging

For any issue, enable debug logging for more information:

```bash
# Set log level
export AINATIVE_LOG_LEVEL=debug

# Or use flag
ainative-code --log-level debug auth login

# Save logs to file
ainative-code --log-level debug auth login 2>&1 | tee auth-debug.log
```

### View Structured Logs

```bash
# JSON output for parsing
export AINATIVE_LOG_FORMAT=json
ainative-code auth login

# Pretty print JSON logs
ainative-code --log-format json auth login 2>&1 | jq .
```

---

## Getting Help

If you still experience issues:

**1. Check Documentation**:
- [Authentication Overview](README.md)
- [OAuth Flow](oauth-flow.md)
- [User Guide](user-guide.md)

**2. Search Issues**:
- https://github.com/AINative-studio/ainative-code/issues

**3. Report Bug**:
```bash
# Include:
# - OS and version
# - CLI version: ainative-code version
# - Debug logs
# - Steps to reproduce

# Create issue:
# https://github.com/AINative-studio/ainative-code/issues/new
```

**4. Contact Support**:
- Email: support@ainative.studio
- Include debug logs and error messages
- Describe steps to reproduce

---

## Common Error Messages Reference

| Error | Meaning | Solution |
|-------|---------|----------|
| `ErrAuthorizationDenied` | User denied authorization | Retry login and approve |
| `ErrAuthorizationTimeout` | User didn't complete in time | Retry with faster response |
| `ErrInvalidState` | CSRF check failed | Retry login, check for MITM |
| `ErrTokenExpired` | Token past expiration | Refresh or re-authenticate |
| `ErrInvalidSignature` | Token signature invalid | Re-authenticate |
| `ErrKeychainAccess` | Can't access keychain | Grant permissions |
| `ErrCallbackServerStart` | Port in use | Kill process or use different port |
| `ErrBrowserOpen` | Can't open browser | Open URL manually |
| `ErrCodeExchangeFailed` | Token exchange failed | Retry authentication |
| `ErrHTTPTimeout` | Request timed out | Check network, increase timeout |

---

## Prevention Tips

**1. Keep CLI Updated**:
```bash
# Check for updates
ainative-code version --check

# Update via Homebrew (macOS)
brew upgrade ainative-code

# Update via apt (Linux)
sudo apt update && sudo apt upgrade ainative-code
```

**2. Regular Token Refresh**:
```yaml
# Enable auto-refresh
platform:
  keychain:
    auto_refresh: true
    refresh_threshold: 5m
```

**3. Monitor Token Status**:
```bash
# Add to cron or scheduled task
ainative-code auth token status
```

**4. Backup Configuration**:
```bash
# Backup config
cp ~/.config/ainative-code/config.yaml \
   ~/.config/ainative-code/config.yaml.backup
```
