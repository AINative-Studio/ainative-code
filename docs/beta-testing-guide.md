# Beta Testing Guide

Thank you for participating in the AINative Cloud beta program!

## Beta Program Overview

**Duration:** 2 weeks (Jan 17 - Jan 31, 2026)
**Participants:** Internal (10) + External (20)
**Focus:** Authentication + Hosted Inference

## What We're Testing

1. **Authentication Flow**
   - Login/logout functionality
   - Token refresh behavior
   - Error handling

2. **Chat Completions**
   - Response quality
   - Response time
   - Streaming reliability

3. **Provider Selection**
   - Auto-selection logic
   - Fallback behavior
   - Credit management

4. **Documentation Quality**
   - Clarity of guides
   - Code example accuracy
   - Troubleshooting effectiveness

## Getting Started

### 1. Installation

```bash
# Install beta version
brew install ainative-code@beta

# Verify installation
ainative-code --version
# Should show: v1.1.0-beta.1
```

### 2. Setup

```bash
# Login with your beta account
ainative-code auth login-backend \
  --email your-beta-email@example.com \
  --password your-beta-password

# Check your credits
ainative-code auth me
```

### 3. Try Key Features

#### Basic Chat
```bash
ainative-code chat-ainative \
  --message "Hello! I'm testing the beta" \
  --auto-provider
```

#### Streaming
```bash
ainative-code chat-ainative \
  --message "Count from 1 to 10" \
  --stream
```

#### Provider Selection
```bash
# Try each provider
ainative-code chat-ainative -m "Test Anthropic" --provider anthropic
ainative-code chat-ainative -m "Test OpenAI" --provider openai
ainative-code chat-ainative -m "Test Google" --provider google
```

## Test Scenarios

Please test these scenarios and report any issues:

### Scenario 1: First-Time User
1. Install CLI
2. Create account
3. Login
4. Send first chat
5. Check credits

**Expected:** Smooth onboarding, clear feedback

### Scenario 2: Token Refresh
1. Login
2. Wait 15+ minutes
3. Send chat (should auto-refresh)

**Expected:** Seamless token refresh, no user action

### Scenario 3: Error Handling
1. Logout
2. Try to chat without auth
3. Login with wrong password
4. Exceed rate limit (send 10 rapid requests)

**Expected:** Clear error messages, helpful suggestions

### Scenario 4: Low Credits
1. Use credits until < 50
2. Send chat

**Expected:** Low credit warning displayed

### Scenario 5: Provider Fallback
1. Set preferred provider
2. Simulate provider failure (we'll guide you)
3. Send chat

**Expected:** Automatic fallback to secondary provider

## Reporting Issues

### Bug Report Template

```markdown
**Bug Description:**
[Clear description of the issue]

**Steps to Reproduce:**
1. Step one
2. Step two
3. ...

**Expected Behavior:**
[What should happen]

**Actual Behavior:**
[What actually happened]

**Environment:**
- OS: [macOS/Linux/Windows]
- CLI Version: [from ainative-code --version]
- Account Tier: [free/pro/enterprise]

**Logs:**
```bash
[Paste relevant logs here]
```
```

**Submit to:** https://github.com/AINative-Studio/ainative-code/issues

### Feature Request Template

```markdown
**Feature:**
[What feature would you like?]

**Use Case:**
[Why do you need this?]

**Proposed Solution:**
[How should it work?]

**Alternatives:**
[Other options considered]
```

## Feedback Collection

### Weekly Survey

We'll send a survey every Friday asking about:
- Overall experience (1-5 stars)
- Feature usability
- Documentation clarity
- Bugs encountered
- Feature requests

### User Interviews

We'd love to interview 5-10 beta testers! If interested:
- **Duration:** 30 minutes
- **Format:** Video call
- **Incentive:** $50 Amazon gift card
- **Sign up:** support@ainative.studio

## Beta Support

### Support Channels

1. **Email:** beta@ainative.studio
2. **GitHub Issues:** Tag with `beta` label
3. **Slack:** #beta-testing channel

### Response Times

- **P0 (Critical):** <2 hours
- **P1 (High):** <4 hours
- **P2 (Medium):** <24 hours
- **P3 (Low):** <72 hours

## Success Metrics

We're tracking:
- Error rate (target: <1%)
- Auth success rate (target: >95%)
- Chat completion success (target: >98%)
- P95 latency (target: <2s)
- User satisfaction (target: >80% positive)

## What Happens After Beta?

1. **Week 1:** Fix critical bugs
2. **Week 2:** Incorporate feedback, polish features
3. **Week 3:** General Availability (GA) release

## Thank You!

Your feedback is invaluable. Every bug report and feature request helps us build a better product.

Questions? Email beta@ainative.studio
