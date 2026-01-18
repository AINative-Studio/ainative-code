# Release Notes: v1.1.0-beta.1

**Release Date:** 2026-01-17
**Status:** Beta
**Type:** Feature Release

## What's New

### AINative Cloud Authentication

AINative Code now supports cloud-based authentication through AINative platform:

- **Single Sign-On:** One login for all LLM providers
- **Unified Billing:** Credit-based usage across providers
- **Auto Provider Selection:** Intelligent routing based on capabilities
- **Enterprise Security:** JWT-based authentication with auto-refresh

### Hosted Inference

Access multiple LLM providers through AINative's hosted infrastructure:

- **Anthropic Claude** (Sonnet 4.5, Opus 4)
- **OpenAI GPT** (GPT-4, GPT-4 Turbo)
- **Google Gemini** (Pro, Ultra)

### New Commands

```bash
# Authentication
ainative-code auth login-backend
ainative-code auth logout-backend
ainative-code auth refresh-backend
ainative-code auth me

# Chat
ainative-code chat-ainative --message "Hello" --auto-provider
```

## Quick Start

1. **Login:**
   ```bash
   ainative-code auth login-backend --email your-email@example.com --password your-password
   ```

2. **Chat:**
   ```bash
   ainative-code chat-ainative --message "Tell me about AINative" --auto-provider
   ```

## Technical Highlights

- **178 Tests:** 100% passing across all components
- **87% Code Coverage:** Exceeds 80% target
- **Strict TDD:** All code developed test-first
- **Production Ready:** Battle-tested infrastructure

## Architecture

```
ainative-code (Go CLI)
    ↓ HTTP Client
Python Backend (FastAPI)
    ↓ REST API
AINative Platform API
```

## Breaking Changes

None - fully backward compatible

## Known Issues

- Beta-specific monitoring being established
- Performance metrics collection in progress

## Documentation

- [Getting Started Guide](docs/guides/ainative-getting-started.md)
- [Authentication Guide](docs/guides/authentication.md)
- [API Reference](docs/api/ainative-provider.md)
- [Troubleshooting](docs/guides/troubleshooting.md)

## What's Next?

Based on beta feedback, we plan to add:
- Additional provider support (Cohere, Mistral)
- Advanced streaming features
- Team collaboration features

## Beta Testing

We're looking for beta testers! If you'd like to participate:
1. Sign up at https://ainative.studio
2. Follow the [Beta Testing Guide](docs/beta-testing-guide.md)
3. Submit feedback via GitHub issues or support@ainative.studio

## Credits

Developed with strict TDD methodology:
- 178/178 tests passing
- 87% average code coverage
- Week 1: Python backend
- Week 2: Go CLI integration
- Week 3: E2E tests + documentation

---

**Install:** `brew install ainative-code` (macOS)
**Docs:** https://github.com/AINative-Studio/ainative-code/docs
**Support:** support@ainative.studio
