# Testing Guide for Issue #94 Fix

## Quick Reference for Testing the Install Script Fallback Logic

### Automated Testing

Run the automated test suite:
```bash
cd /Users/aideveloper/AINative-Code
chmod +x test-install-fallback.sh
./test-install-fallback.sh
```

Expected: All 7 tests should pass.

---

## Manual Testing Scenarios

### Test 1: Successful Sudo Installation
**Purpose**: Verify normal installation with sudo works

```bash
# Clean up any previous installations
rm -f /usr/local/bin/ainative-code ~/.local/bin/ainative-code ~/bin/ainative-code

# Run installer (you'll need the actual binary or mock one)
./install.sh

# When prompted, enter your sudo password
# Expected: Installs to /usr/local/bin
# Verify:
which ainative-code
# Should output: /usr/local/bin/ainative-code
```

---

### Test 2: Cancelled Sudo Prompt (User Says No)
**Purpose**: Verify fallback when user cancels sudo

```bash
# Clean up previous installations
rm -f /usr/local/bin/ainative-code ~/.local/bin/ainative-code ~/bin/ainative-code

# Run installer
./install.sh

# When prompted for sudo password, press Ctrl+C or just wait for timeout
# Expected: Should fall back to ~/.local/bin
# Expected output should include:
# [WARNING] Sudo installation failed or was cancelled
# [INFO] Falling back to user directory installation...
# [SUCCESS] Successfully installed to ~/.local/bin/ainative-code

# Verify fallback installation
ls -la ~/.local/bin/ainative-code

# Add to PATH as instructed and verify
export PATH="$HOME/.local/bin:$PATH"
which ainative-code
# Should output: ~/.local/bin/ainative-code
```

---

### Test 3: Simulate No Sudo Access
**Purpose**: Verify behavior when user has no sudo privileges

```bash
# In a Docker container or VM where you don't have sudo:
docker run -it --rm -v $(pwd):/workspace ubuntu:latest bash
cd /workspace

# Try to use sudo (should fail)
sudo echo test
# Should output: sudo: command not found OR permission denied

# Run installer
./install.sh

# Expected: Should automatically fall back to ~/.local/bin without prompting
# Should NOT show sudo prompt at all
```

---

### Test 4: Passwordless Sudo
**Purpose**: Verify silent installation with passwordless sudo

```bash
# Set up passwordless sudo (in VM/container for safety):
echo "$USER ALL=(ALL) NOPASSWD: /bin/cp, /bin/chmod" | sudo tee /etc/sudoers.d/ainative-install

# Run installer
./install.sh

# Expected: Should install to /usr/local/bin without password prompt
# Should complete quickly and silently
```

---

### Test 5: Custom Install Directory
**Purpose**: Verify INSTALL_DIR environment variable works

```bash
# Create custom directory
mkdir -p ~/my-custom-bin

# Set environment variable
export INSTALL_DIR="$HOME/my-custom-bin"

# Run installer
./install.sh

# Expected: Should install to ~/my-custom-bin
# Verify:
ls -la ~/my-custom-bin/ainative-code
```

---

### Test 6: Verify PATH Setup Instructions
**Purpose**: Ensure users get clear PATH setup instructions

```bash
# Clean installations
rm -f ~/.local/bin/ainative-code

# Remove ~/.local/bin from PATH if it's there
export PATH=$(echo $PATH | tr ':' '\n' | grep -v ".local/bin" | tr '\n' ':')

# Run installer (cancel sudo when prompted)
./install.sh

# Expected output should include:
# [WARNING] Installation directory /home/user/.local/bin is not in your PATH
# [INFO] To use ainative-code, you need to add it to your PATH
#
# Run these commands to add it automatically:
#   echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
#   source ~/.bashrc

# Verify the instructions work
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
which ainative-code
```

---

### Test 7: Directory Creation
**Purpose**: Verify script creates ~/.local/bin if it doesn't exist

```bash
# Remove directory
rm -rf ~/.local/bin

# Run installer (cancel sudo)
./install.sh

# Expected:
# [INFO] Created directory /home/user/.local/bin
# [SUCCESS] Successfully installed to /home/user/.local/bin/ainative-code

# Verify directory was created
ls -la ~/.local/bin
```

---

### Test 8: Both Fallback Directories Fail (Edge Case)
**Purpose**: Verify error handling when all installation attempts fail

```bash
# Make both fallback directories unwritable (requires sudo)
sudo mkdir -p ~/.local/bin ~/bin
sudo chown root:root ~/.local/bin ~/bin
sudo chmod 000 ~/.local/bin ~/bin

# Also make /usr/local/bin unwritable
# (or cancel sudo prompt)

# Run installer
./install.sh

# Expected: Should show manual installation instructions
# [ERROR] Could not install to any location
# [ERROR] Tried: /usr/local/bin, ~/.local/bin, ~/bin
#
# You can manually install by running:
#   mkdir -p ~/.local/bin
#   cp /tmp/xxx/ainative-code ~/.local/bin/ainative-code
#   chmod +x ~/.local/bin/ainative-code
#   export PATH="$HOME/.local/bin:$PATH"

# Clean up
sudo rm -rf ~/.local/bin ~/bin
```

---

### Test 9: Verify Error Messages Are Visible
**Purpose**: Ensure no silent failures

```bash
# Run installer with network disconnected (to test download failure)
# Or modify install.sh to use invalid URL temporarily

# Expected: Clear error messages like:
# [ERROR] Failed to download from https://...
# [ERROR] Failed to download AINative Code archive

# NOT: Silent failure with no output
```

---

### Test 10: Shell Detection
**Purpose**: Verify correct shell RC file is detected

```bash
# Test with bash
bash -c './install.sh'
# Should detect .bashrc or .bash_profile

# Test with zsh
zsh -c './install.sh'
# Should detect .zshrc

# Verify in output:
# Should show correct file path in PATH instructions
```

---

## Expected Output Examples

### Successful Installation with Sudo
```
╔════════════════════════════════════════════╗
║      AINative Code Installation Script    ║
╚════════════════════════════════════════════╝

[INFO] Fetching latest version...
[INFO] Latest version: v1.0.0
[INFO] Detected platform: Darwin_arm64
[INFO] Downloading from https://github.com/...
[INFO] Verifying checksum...
[SUCCESS] Checksum verified
[INFO] Extracting archive...
[INFO] Installing to /usr/local/bin/ainative-code...
[WARNING] Installing to /usr/local/bin requires root privileges
[INFO] Please enter your password for sudo access...
Password:
[SUCCESS] Installation complete!
[INFO] Verifying installation...
[SUCCESS] AINative Code v1.0.0 installed successfully!

╔════════════════════════════════════════════╗
║           Installation Complete!           ║
╚════════════════════════════════════════════╝
```

### Fallback Installation (Sudo Cancelled)
```
[INFO] Installing to /usr/local/bin/ainative-code...
[WARNING] Installing to /usr/local/bin requires root privileges
[INFO] Please enter your password for sudo access...
^C
[WARNING] Sudo installation failed or was cancelled

[INFO] Falling back to user directory installation...
[INFO] Attempting fallback installation to /Users/user/.local/bin...
[INFO] Created directory /Users/user/.local/bin
[SUCCESS] Successfully installed to /Users/user/.local/bin/ainative-code
[SUCCESS] Installation complete!

[WARNING] Installation directory /Users/user/.local/bin is not in your PATH
[INFO] To use ainative-code, you need to add it to your PATH

Run these commands to add it automatically:

  echo 'export PATH="$HOME/.local/bin:$PATH"' >> /Users/user/.zshrc
  source /Users/user/.zshrc

After updating your PATH, verify the installation with:
  ainative-code version
```

---

## Validation Checklist

After running tests, verify:

- [ ] All automated tests pass
- [ ] No syntax errors in script
- [ ] Sudo installation works when password provided
- [ ] Fallback works when sudo cancelled
- [ ] Fallback works when no sudo access
- [ ] ~/.local/bin is created if missing
- [ ] PATH instructions are shown when needed
- [ ] No silent failures in any scenario
- [ ] Error messages are clear and helpful
- [ ] Manual installation instructions shown when needed
- [ ] Shell detection works (bash, zsh)
- [ ] Custom INSTALL_DIR respected

---

## Troubleshooting

### Test Fails: "ainative-code binary not found"
**Cause**: The test is trying to actually install, but the binary doesn't exist in releases yet

**Solution**: Create a mock binary for testing:
```bash
mkdir -p /tmp/ainative-code-test
cat > /tmp/ainative-code-test/ainative-code << 'EOF'
#!/bin/bash
echo "AINative Code v1.0.0 (test)"
exit 0
EOF
chmod +x /tmp/ainative-code-test/ainative-code

# Now modify the test to use this mock binary
```

### Test Fails: "Permission denied"
**Cause**: Test directories have incorrect permissions

**Solution**: Reset permissions:
```bash
chmod 755 ~/.local/bin ~/bin
chown $USER:$USER ~/.local/bin ~/bin
```

### Test Fails: "PATH not updated"
**Cause**: PATH changes don't persist in the current shell

**Solution**: Source your shell RC file or start a new shell:
```bash
source ~/.bashrc  # or ~/.zshrc
# or
exec $SHELL
```

---

## Clean Up After Testing

```bash
# Remove test installations
sudo rm -f /usr/local/bin/ainative-code
rm -f ~/.local/bin/ainative-code
rm -f ~/bin/ainative-code

# Remove test directories (if you want)
# rm -rf ~/.local/bin ~/bin  # Be careful!

# Remove PATH modifications from shell RC (if added during testing)
# Edit ~/.bashrc or ~/.zshrc and remove the export PATH line
```

---

## Testing in CI/CD

For automated testing in CI environments:

```yaml
# Example GitHub Actions workflow
- name: Test install script
  run: |
    chmod +x test-install-fallback.sh
    ./test-install-fallback.sh

- name: Test fallback installation (no sudo)
  run: |
    # CI environments typically don't have sudo
    ./install.sh
    # Verify installation
    test -f ~/.local/bin/ainative-code
    test -x ~/.local/bin/ainative-code
```

---

## Success Criteria

The fix is successful if:

1. ✅ Script never fails silently
2. ✅ Clear error messages for all failure scenarios
3. ✅ Fallback to user directory works automatically
4. ✅ PATH setup instructions are clear and accurate
5. ✅ All automated tests pass
6. ✅ Manual testing confirms expected behavior
7. ✅ No breaking changes to existing functionality

---

**Last Updated**: January 8, 2026
**Status**: Ready for Testing
