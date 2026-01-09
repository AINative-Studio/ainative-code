# Issue #103 Fix Report: Setup Wizard Arrow Key Navigation

## Summary
Fixed arrow key navigation in the setup wizard which was not working despite the text input fix in issue #93.

## Root Cause Analysis

### The Problem
The arrow key navigation was broken due to a subtle bug in how the `PromptModel` managed the choices state:

1. **Value Receiver Issue**: The `renderChoices()` method on line 298 had a value receiver (`func (m PromptModel)`) instead of a pointer receiver
2. **Stale State**: The method attempted to update `m.choices = choices` on line 299, but this only modified a **copy** of the model, not the original
3. **Empty Choices Field**: The original model's `m.choices` field remained empty throughout the application lifecycle
4. **Failed Condition**: In the `Update()` method at line 86, the condition `if len(m.choices) > 0` always evaluated to false
5. **No Navigation**: Arrow key presses were ignored because the cursor movement logic was conditional on having choices

### Why Text Input Worked
The fix for issue #93 added proper text input handling in the `default` case of the key handler (lines 104-109), which passed unhandled keys to the textinput component. This worked because it didn't depend on the `m.choices` field.

## Solution Implemented

### Changes Made

#### 1. Removed Unused `choices` Field
**File**: `/Users/aideveloper/AINative-Code/internal/setup/prompts.go`
**Lines**: 33-40

```go
// Before:
type PromptModel struct {
    currentStep Step
    Selections  map[string]interface{}
    textInput   textinput.Model
    cursor      int
    choices     []string  // ← Removed this problematic field
    err         error
}

// After:
type PromptModel struct {
    currentStep Step
    Selections  map[string]interface{}
    textInput   textinput.Model
    cursor      int
    err         error
}
```

#### 2. Updated Model Initialization
**File**: `/Users/aideveloper/AINative-Code/internal/setup/prompts.go`
**Lines**: 49-54

Removed the `choices: []string{}` initialization from `NewPromptModel()`.

#### 3. Fixed Arrow Key Handling
**File**: `/Users/aideveloper/AINative-Code/internal/setup/prompts.go`
**Lines**: 80-88

```go
// Before:
case "up", "k":
    if m.cursor > 0 {
        m.cursor--
    }

case "down", "j":
    if len(m.choices) > 0 && m.cursor < len(m.choices)-1 {
        m.cursor++
    }

// After:
case "up", "k":
    if !m.isTextInputStep() && m.cursor > 0 {
        m.cursor--
    }

case "down", "j":
    if !m.isTextInputStep() && m.cursor < m.getChoiceCount()-1 {
        m.cursor++
    }
```

**Key improvements**:
- Added `!m.isTextInputStep()` guard to prevent arrow keys from moving cursor during text input
- Replaced `len(m.choices)` with `m.getChoiceCount()` which determines choice count based on current step

#### 4. Simplified `renderChoices()` Method
**File**: `/Users/aideveloper/AINative-Code/internal/setup/prompts.go`
**Lines**: 296-318

Removed the line `m.choices = choices` since we no longer use that field. The method now only renders the UI without trying to modify state.

#### 5. Added `getChoiceCount()` Helper Method
**File**: `/Users/aideveloper/AINative-Code/internal/setup/prompts.go`
**Lines**: 515-530

```go
func (m PromptModel) getChoiceCount() int {
    switch m.currentStep {
    case StepProvider:
        return 4 // Anthropic, OpenAI, Google, Ollama
    case StepAnthropicModel:
        return 4 // 4 Claude models
    case StepOpenAIModel:
        return 3 // 3 GPT models
    case StepGoogleModel:
        return 2 // 2 Gemini models
    case StepColorScheme:
        return 3 // Auto, Light, Dark
    default:
        return 0
    }
}
```

This method provides the correct number of choices for each step without relying on stored state.

#### 6. Fixed Compilation Errors in session.go
**File**: `/Users/aideveloper/AINative-Code/internal/cmd/session.go`
**Lines**: 386, 424

Fixed type conversion issues where `MessageRole` (custom type) needed to be converted to `string`:
- Line 386: `roleCounts[string(msg.Role)]++`
- Line 424: `strings.ToUpper(string(msg.Role))`

## Testing

### New Test File Created
**File**: `/Users/aideveloper/AINative-Code/internal/setup/prompts_test.go`

Created comprehensive test suite with 4 test functions and 23 test cases:

1. **TestArrowKeyNavigation** (9 test cases)
   - Tests up/down arrow navigation across all choice-based steps
   - Tests vim-style navigation (j/k keys)
   - Tests boundary conditions (first/last options)

2. **TestArrowKeysDoNotWorkOnTextInputSteps** (6 test cases)
   - Ensures arrow keys are disabled during text input steps
   - Tests all text input steps: API key entries, URLs, model names

3. **TestGetChoiceCount** (7 test cases)
   - Validates correct choice counts for each step type
   - Tests choice steps, text input steps, and yes/no steps

4. **TestCursorResetOnStepTransition** (1 test case)
   - Ensures cursor resets to 0 when moving to next step

### Test Results
All 32 setup package tests pass:
```
PASS: TestArrowKeyNavigation (0.00s)
PASS: TestArrowKeysDoNotWorkOnTextInputSteps (0.00s)
PASS: TestGetChoiceCount (0.00s)
PASS: TestCursorResetOnStepTransition (0.00s)
... (28 existing tests also pass)
ok   github.com/AINative-studio/ainative-code/internal/setup  0.989s
```

## How to Test the Fix

### Manual Testing

1. **Build the application**:
   ```bash
   go build -o ainative-code ./cmd/ainative-code
   ```

2. **Remove existing configuration** (to trigger setup wizard):
   ```bash
   rm -rf ~/.ainative-code/config.yaml ~/.ainative-code/.setup_complete
   ```

3. **Run the setup wizard**:
   ```bash
   ./ainative-code setup
   ```

4. **Test arrow key navigation**:
   - On the "LLM Provider Selection" screen:
     - Press ↓ (down arrow) - cursor should move to "OpenAI (GPT)"
     - Press ↓ again - cursor should move to "Google (Gemini)"
     - Press ↑ (up arrow) - cursor should move back to "OpenAI (GPT)"
     - Press 'j' - cursor should move down (vim-style)
     - Press 'k' - cursor should move up (vim-style)

   - Select "Anthropic (Claude)" and press Enter

   - On the "Enter your Anthropic API key" screen:
     - Arrow keys should NOT move cursor
     - Text input should work normally
     - Type a test key and press Enter

   - On the "Which Claude model" screen:
     - Arrow keys should work again
     - Navigate through the 4 model options

   - Continue testing through other steps (color scheme, etc.)

5. **Test boundary conditions**:
   - Press ↑ when on first option - should stay on first option
   - Press ↓ when on last option - should stay on last option

### Automated Testing

Run the test suite:
```bash
# Run all setup tests
go test -v ./internal/setup

# Run only arrow key tests
go test -v ./internal/setup -run "TestArrowKey"

# Run with coverage
go test -cover ./internal/setup
```

## Impact

### Fixed Issues
- Arrow key navigation now works on all choice-based steps
- Vim-style navigation (j/k) works correctly
- Arrow keys properly disabled during text input
- Maintains all functionality from issue #93 fix

### Affected Components
- Provider selection (4 choices)
- Anthropic model selection (4 choices)
- OpenAI model selection (3 choices)
- Google model selection (2 choices)
- Color scheme selection (3 choices)

### Backward Compatibility
- No breaking changes
- All existing tests continue to pass
- Configuration format unchanged

## Files Modified

1. `/Users/aideveloper/AINative-Code/internal/setup/prompts.go`
   - Lines 33-40: Removed `choices` field from `PromptModel`
   - Lines 49-54: Updated `NewPromptModel()` initialization
   - Lines 80-88: Fixed arrow key handling logic
   - Lines 296-318: Simplified `renderChoices()` method
   - Lines 515-530: Added `getChoiceCount()` helper method

2. `/Users/aideveloper/AINative-Code/internal/cmd/session.go`
   - Line 386: Added type conversion for `msg.Role`
   - Line 424: Added type conversion for `msg.Role`

3. `/Users/aideveloper/AINative-Code/internal/setup/prompts_test.go` (NEW)
   - Comprehensive test suite for arrow key navigation
   - 4 test functions, 23 test cases

## Conclusion

The fix addresses the root cause by eliminating the problematic state management pattern and replacing it with a deterministic approach based on the current step. Arrow key navigation now works reliably across all choice-based steps while preserving the text input functionality fixed in issue #93.
