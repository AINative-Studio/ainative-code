# Bug #104 Code Changes - Visual Comparison

## Change 1: internal/setup/wizard.go - WizardConfig struct

### BEFORE
```go
// WizardConfig holds configuration for the setup wizard
type WizardConfig struct {
    ConfigPath      string
    SkipValidation  bool
    InteractiveMode bool
}
```

### AFTER
```go
// WizardConfig holds configuration for the setup wizard
type WizardConfig struct {
    ConfigPath      string
    SkipValidation  bool
    InteractiveMode bool
    Force           bool  // ← NEW: Added Force flag support
}
```

---

## Change 2: internal/setup/wizard.go - Run() method

### BEFORE
```go
// Run executes the setup wizard flow
func (w *Wizard) Run() (*WizardResult, error) {
    // Check if already initialized
    if w.checkAlreadyInitialized() {
        return w.result, nil
    }
```

### AFTER
```go
// Run executes the setup wizard flow
func (w *Wizard) Run() (*WizardResult, error) {
    // Check if already initialized (skip if force flag is set)
    if !w.config.Force && w.checkAlreadyInitialized() {
        return w.result, nil
    }
```

**Key Change:** Added `!w.config.Force &&` to bypass check when Force=true

---

## Change 3: internal/cmd/setup.go - wizardConfig initialization

### BEFORE
```go
// Configure wizard
wizardConfig := setup.WizardConfig{
    ConfigPath:      setupConfigPath,
    SkipValidation:  setupSkipValidation,
    InteractiveMode: !setupNonInteractive,
}
```

### AFTER
```go
// Configure wizard
wizardConfig := setup.WizardConfig{
    ConfigPath:      setupConfigPath,
    SkipValidation:  setupSkipValidation,
    InteractiveMode: !setupNonInteractive,
    Force:           setupForce,  // ← NEW: Pass force flag to wizard
}
```

---

## Change 4: internal/setup/wizard_test.go - New test functions

### ADDED TEST 1: TestForceFlag_BypassesInitializedCheck

```go
func TestForceFlag_BypassesInitializedCheck(t *testing.T) {
    // Test verifies that Force=true bypasses initialization check
    // even when config already exists
    
    // Part A: Test WITHOUT Force flag - should skip setup
    wizardNoForce := NewWizard(ctx, WizardConfig{
        Force: false,
    })
    result, err := wizardNoForce.Run()
    assert.True(t, result.SkippedSetup)  // ← Should skip
    
    // Part B: Test WITH Force flag - should run wizard
    wizardWithForce := NewWizard(ctx, WizardConfig{
        Force: true,
    })
    result, err = wizardWithForce.Run()
    assert.False(t, result.SkippedSetup)  // ← Should NOT skip
    assert.True(t, result.MarkerCreated)   // ← Should complete
}
```

### ADDED TEST 2: TestForceFlag_WorksOnFreshInstall

```go
func TestForceFlag_WorksOnFreshInstall(t *testing.T) {
    // Test verifies Force flag works correctly on fresh install
    // (no existing config files)
    
    wizard := NewWizard(ctx, WizardConfig{
        Force: true,
    })
    
    result, err := wizard.Run()
    assert.False(t, result.SkippedSetup)
    assert.True(t, result.MarkerCreated)
    assert.FileExists(t, result.ConfigPath)
}
```

---

## Summary of Changes

| File | Lines Changed | Type |
|------|---------------|------|
| internal/setup/wizard.go | +1 field, +1 condition | Feature |
| internal/cmd/setup.go | +1 field assignment | Integration |
| internal/setup/wizard_test.go | +93 lines | Tests |
| **Total** | **~95 lines** | **Bug Fix** |

---

## Logic Flow Comparison

### BEFORE (Broken)
```
┌─────────────────────┐
│ setup --force       │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ cmd/setup.go        │
│ Check if first run  │
│ (bypassed by force) │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ wizard.Run()        │
│ checkInitialized()  │ ◄── BUG: Always checks
│ → return early      │     even with --force
└─────────────────────┘

Result: Wizard never runs ❌
```

### AFTER (Fixed)
```
┌─────────────────────┐
│ setup --force       │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ cmd/setup.go        │
│ Pass Force=true     │ ◄── NEW: Pass flag
│ to wizard config    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ wizard.Run()        │
│ if !Force &&        │ ◄── FIX: Check Force
│    checkInit()      │     before returning
│ → continue wizard   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Interactive Wizard  │
│ Runs Successfully   │
└─────────────────────┘

Result: Wizard runs correctly ✅
```

---

## Testing Matrix

| Scenario | Force Flag | Already Init | Expected | Result |
|----------|-----------|--------------|----------|--------|
| Fresh install | false | No | Run wizard | ✅ Pass |
| Fresh install | true | No | Run wizard | ✅ Pass |
| Already configured | false | Yes | Skip wizard | ✅ Pass |
| Already configured | true | Yes | **Run wizard** | ✅ **Pass (Fixed!)** |

---

## Verification Commands

```bash
# Build the binary
go build -o bin/ainative-code ./cmd/ainative-code

# Test 1: Normal behavior (should skip)
ainative-code setup
# Expected: "AINative Code is already configured!"

# Test 2: Force flag (should run)
ainative-code setup --force --non-interactive --skip-validation
# Expected: "Welcome to AINative Code!" + "Setup Complete!"

# Test 3: Run unit tests
go test ./internal/setup/... -v -run TestForceFlag
# Expected: PASS (2 tests)
```
