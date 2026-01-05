# Compatibility Matrix - AINative Code v1.0

This document provides comprehensive compatibility information for AINative Code v1.0, including platform support, dependency versions, and tested configurations.

**Last Updated**: January 4, 2026
**Version**: v1.0.0

---

## Table of Contents

1. [Platform Compatibility](#platform-compatibility)
2. [Go Version Requirements](#go-version-requirements)
3. [LLM Provider Compatibility](#llm-provider-compatibility)
4. [Dependency Versions](#dependency-versions)
5. [Operating System Support](#operating-system-support)
6. [Terminal Emulator Compatibility](#terminal-emulator-compatibility)
7. [Architecture Support](#architecture-support)
8. [Cloud Platform Integration](#cloud-platform-integration)
9. [Browser Support](#browser-support)
10. [Tested Configurations](#tested-configurations)

---

## Platform Compatibility

### Summary Table

| Platform | Status | Minimum Version | Recommended Version | Architectures |
|----------|--------|----------------|---------------------|---------------|
| macOS | ✅ Fully Supported | 10.15 Catalina | 14.0 Sonoma | amd64, arm64 |
| Linux | ✅ Fully Supported | Ubuntu 20.04, Debian 10 | Ubuntu 22.04, Debian 12 | amd64, arm64 |
| Windows | ✅ Fully Supported | Windows 10 | Windows 11 | amd64 |
| FreeBSD | ⚠️ Community Support | FreeBSD 13 | FreeBSD 14 | amd64 |
| Docker | ✅ Fully Supported | N/A | Latest | linux/amd64, linux/arm64 |

**Legend**:
- ✅ Fully Supported: Official support with regular testing
- ⚠️ Community Support: Built and tested by community
- ❌ Not Supported: Not compatible

---

## Go Version Requirements

### Required Go Version

| AINative Code Version | Minimum Go | Recommended Go | Maximum Tested |
|----------------------|------------|----------------|----------------|
| v1.0.0 | 1.21.0 | 1.25.5 | 1.25.5 |

### Go Version Testing Matrix

| Go Version | Linux | macOS | Windows | Status |
|------------|-------|-------|---------|--------|
| 1.21.x | ✅ | ✅ | ✅ | Supported |
| 1.22.x | ✅ | ✅ | ✅ | Supported |
| 1.23.x | ✅ | ✅ | ✅ | Supported |
| 1.24.x | ✅ | ✅ | ✅ | Supported |
| 1.25.x | ✅ | ✅ | ✅ | Recommended |
| 1.20.x and below | ❌ | ❌ | ❌ | Not Supported |

### Required Go Features

AINative Code uses the following Go features:
- **Generics** (Go 1.18+): Type-safe collections and utilities
- **Workspace Mode** (Go 1.18+): Development workflow
- **Build Constraints** (Go 1.17+): Platform-specific builds
- **Embed Directive** (Go 1.16+): Embedded resources

---

## LLM Provider Compatibility

### Supported Providers

| Provider | Status | Minimum API Version | Features | Notes |
|----------|--------|-------------------|----------|-------|
| **Anthropic Claude** | ✅ Full Support | 2023-06-01 | Streaming, Caching, Extended Thinking | All features supported |
| **OpenAI GPT** | ✅ Full Support | v1 (2023-07) | Streaming, Function Calling | GPT-3.5, GPT-4 |
| **Google Gemini** | ✅ Full Support | v1beta | Streaming, Multimodal | Gemini Pro, Vertex AI |
| **AWS Bedrock** | ✅ Full Support | 2023-09-30 | Streaming, Multiple Models | Claude on Bedrock |
| **Azure OpenAI** | ✅ Full Support | 2023-12-01-preview | Streaming, Deployments | Azure-hosted GPT |
| **Ollama** | ✅ Full Support | 0.1.0+ | Local Models | Open-source models |

### Provider Feature Matrix

| Feature | Claude | OpenAI | Gemini | Bedrock | Azure | Ollama |
|---------|--------|--------|--------|---------|-------|--------|
| Streaming | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Function Calling | ✅ | ✅ | ✅ | ✅ | ✅ | ⚠️ |
| Vision/Multimodal | ✅ | ✅ | ✅ | ✅ | ✅ | ⚠️ |
| Prompt Caching | ✅ | ❌ | ❌ | ⚠️ | ❌ | ❌ |
| Extended Thinking | ✅ | ❌ | ❌ | ⚠️ | ❌ | ❌ |
| JSON Mode | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Token Counting | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

**Legend**:
- ✅ Fully Supported
- ⚠️ Partial Support (depends on model)
- ❌ Not Available

### Model Compatibility

#### Anthropic Claude
| Model | Status | Context Window | Features |
|-------|--------|---------------|----------|
| claude-3-5-sonnet-20241022 | ✅ Recommended | 200K | Caching, Thinking |
| claude-3-5-sonnet-20240620 | ✅ Supported | 200K | Caching |
| claude-3-opus-20240229 | ✅ Supported | 200K | - |
| claude-3-sonnet-20240229 | ✅ Supported | 200K | - |
| claude-3-haiku-20240307 | ✅ Supported | 200K | - |

#### OpenAI
| Model | Status | Context Window | Features |
|-------|--------|---------------|----------|
| gpt-4-turbo-preview | ✅ Recommended | 128K | Vision, JSON |
| gpt-4 | ✅ Supported | 8K | Function Calling |
| gpt-4-32k | ✅ Supported | 32K | Function Calling |
| gpt-3.5-turbo | ✅ Supported | 16K | Function Calling |
| gpt-3.5-turbo-16k | ✅ Supported | 16K | Function Calling |

#### Google Gemini
| Model | Status | Context Window | Features |
|-------|--------|---------------|----------|
| gemini-1.5-pro-latest | ✅ Recommended | 1M | Multimodal |
| gemini-1.5-flash-latest | ✅ Supported | 1M | Fast, Multimodal |
| gemini-pro | ✅ Supported | 32K | - |

---

## Dependency Versions

### Core Dependencies

| Dependency | Version | Purpose | License |
|------------|---------|---------|---------|
| Go | 1.25.5 | Runtime | BSD-3-Clause |
| Bubble Tea | v1.3.10 | TUI Framework | MIT |
| Cobra | v1.10.2 | CLI Framework | Apache-2.0 |
| Viper | v1.21.0 | Configuration | MIT |
| zerolog | v1.34.0 | Logging | MIT |
| lumberjack | v2.2.1 | Log Rotation | MIT |
| Anthropic SDK | v1.19.0 | Claude API | Apache-2.0 |
| JWT | v5.3.0 | Authentication | MIT |
| SQLite | v1.14.32 | Local Storage | MIT |
| UUID | v1.6.0 | ID Generation | BSD-3-Clause |

### Complete Dependency Tree

See [go.mod](../../go.mod) for complete dependency list with exact versions.

### Dependency Update Policy

- **Major Dependencies**: Updated quarterly after thorough testing
- **Security Patches**: Applied immediately upon release
- **Minor Updates**: Applied monthly
- **Breaking Changes**: Announced 30 days in advance

---

## Operating System Support

### macOS

| Version | Code Name | Status | Architectures | Notes |
|---------|-----------|--------|---------------|-------|
| 14.x | Sonoma | ✅ Fully Tested | Intel, Apple Silicon | Recommended |
| 13.x | Ventura | ✅ Fully Tested | Intel, Apple Silicon | Supported |
| 12.x | Monterey | ✅ Tested | Intel, Apple Silicon | Supported |
| 11.x | Big Sur | ✅ Tested | Intel, Apple Silicon | Supported |
| 10.15 | Catalina | ⚠️ Community Tested | Intel | Minimum version |
| 10.14 and below | - | ❌ Not Supported | Intel | Too old |

**macOS-Specific Features**:
- ✅ Keychain integration
- ✅ Notarization (planned v1.1)
- ✅ Universal binaries (Intel + Apple Silicon)
- ✅ Homebrew distribution

### Linux

#### Ubuntu
| Version | Code Name | Status | Architectures | Notes |
|---------|-----------|--------|---------------|-------|
| 24.04 LTS | Noble | ✅ Fully Tested | amd64, arm64 | Latest LTS |
| 22.04 LTS | Jammy | ✅ Fully Tested | amd64, arm64 | Recommended |
| 20.04 LTS | Focal | ✅ Tested | amd64, arm64 | Minimum LTS |
| 18.04 LTS | Bionic | ⚠️ Community Tested | amd64 | EOL April 2023 |

#### Debian
| Version | Code Name | Status | Architectures | Notes |
|---------|-----------|--------|---------------|-------|
| 12 | Bookworm | ✅ Fully Tested | amd64, arm64 | Latest stable |
| 11 | Bullseye | ✅ Tested | amd64, arm64 | Supported |
| 10 | Buster | ⚠️ Community Tested | amd64, arm64 | Minimum version |

#### Red Hat Enterprise Linux (RHEL) / CentOS / Fedora
| Distribution | Versions | Status | Notes |
|--------------|----------|--------|-------|
| RHEL | 8.x, 9.x | ✅ Tested | Enterprise Linux |
| CentOS Stream | 8, 9 | ✅ Tested | Community version |
| Fedora | 38, 39, 40 | ✅ Tested | Latest releases |
| Rocky Linux | 8.x, 9.x | ✅ Tested | RHEL compatible |
| AlmaLinux | 8.x, 9.x | ✅ Tested | RHEL compatible |

#### Other Linux Distributions
| Distribution | Status | Notes |
|--------------|--------|-------|
| Arch Linux | ✅ Community Support | Rolling release |
| Manjaro | ✅ Community Support | Arch-based |
| openSUSE | ⚠️ Community Support | Leap and Tumbleweed |
| Gentoo | ⚠️ Community Support | Build from source |
| Alpine Linux | ✅ Docker Images | Minimal distro |

**Linux-Specific Features**:
- ✅ Secret Service integration (GNOME Keyring, KWallet)
- ✅ SystemD integration
- ✅ APT/YUM package repositories (planned)

### Windows

| Version | Status | Architectures | Notes |
|---------|--------|---------------|-------|
| Windows 11 | ✅ Fully Tested | amd64 | Recommended |
| Windows 10 22H2 | ✅ Fully Tested | amd64 | Supported |
| Windows 10 21H2 | ✅ Tested | amd64 | Supported |
| Windows 10 20H2 | ⚠️ Community Tested | amd64 | Minimum version |
| Windows Server 2022 | ✅ Tested | amd64 | Server OS |
| Windows Server 2019 | ✅ Tested | amd64 | Server OS |
| Windows 8.1 and below | ❌ Not Supported | - | Too old |

**Windows-Specific Features**:
- ✅ Windows Credential Manager integration
- ✅ Windows Terminal support
- ✅ PowerShell integration
- ✅ Authenticode signing (planned v1.1)

---

## Terminal Emulator Compatibility

### Fully Compatible

| Terminal | macOS | Linux | Windows | Features | Notes |
|----------|-------|-------|---------|----------|-------|
| iTerm2 | ✅ | - | - | Full color, Unicode | Recommended for macOS |
| Warp | ✅ | ✅ | - | Full color, Unicode | Modern terminal |
| Alacritty | ✅ | ✅ | ✅ | Full color, Unicode | Cross-platform |
| Kitty | ✅ | ✅ | - | Full color, Unicode | GPU-accelerated |
| Windows Terminal | - | - | ✅ | Full color, Unicode | Recommended for Windows |
| Hyper | ✅ | ✅ | ✅ | Full color, Unicode | Electron-based |

### Compatible with Limitations

| Terminal | macOS | Linux | Windows | Limitations |
|----------|-------|-------|---------|-------------|
| Terminal.app | ✅ | - | - | Limited color support on older macOS |
| GNOME Terminal | - | ✅ | - | Unicode issues on some locales |
| Konsole | - | ✅ | - | Color scheme compatibility |
| xterm | - | ✅ | - | Limited Unicode, basic colors only |
| PowerShell | - | - | ✅ | UTF-8 encoding issues |
| cmd.exe | - | - | ⚠️ | Minimal Unicode support |

### Recommended Settings

For best experience:

```bash
# Set UTF-8 encoding
export LANG=en_US.UTF-8
export LC_ALL=en_US.UTF-8

# Enable 256 colors
export TERM=xterm-256color

# Or for true color support
export COLORTERM=truecolor
```

---

## Architecture Support

### Supported Architectures

| Architecture | Linux | macOS | Windows | Docker | Status |
|--------------|-------|-------|---------|--------|--------|
| amd64 (x86-64) | ✅ | ✅ | ✅ | ✅ | Fully Supported |
| arm64 (aarch64) | ✅ | ✅ (M1/M2/M3) | ⚠️ Preview | ✅ | Fully Supported |
| armv7 | ⚠️ Community | - | - | ⚠️ | Limited Support |
| 386 (i386) | ❌ | ❌ | ❌ | ❌ | Not Supported |

### Performance by Architecture

| Architecture | Relative Performance | Notes |
|--------------|---------------------|-------|
| Apple Silicon (M1/M2/M3) | 100% (baseline) | Fastest |
| amd64 (modern Intel/AMD) | 95% | Excellent |
| arm64 (Linux) | 90% | Very Good |
| armv7 | 50% | Limited performance |

---

## Cloud Platform Integration

### AINative Platform

| Service | Endpoint | Status | Version | Notes |
|---------|----------|--------|---------|-------|
| Auth Service | auth.ainative.studio | ✅ | v1 | JWT/OAuth 2.0 |
| ZeroDB | api.zerodb.ainative.studio | ✅ | v1 | NoSQL, Vector, Quantum |
| Design | design.ainative.studio | ✅ | v1 | Design Tokens |
| Strapi CMS | strapi.ainative.studio | ✅ | v4 | Content Management |
| RLHF | rlhf.ainative.studio | ✅ | v1 | Feedback Collection |

### Cloud Provider Support

| Provider | Status | Features | Notes |
|----------|--------|----------|-------|
| **AWS** | ✅ Full Support | Bedrock, S3, Secrets Manager | IAM authentication |
| **Google Cloud** | ✅ Full Support | Vertex AI, Cloud Storage | Service account auth |
| **Azure** | ✅ Full Support | Azure OpenAI, Blob Storage | Azure AD integration |
| **DigitalOcean** | ⚠️ Community | Spaces, PostgreSQL | Basic support |
| **Cloudflare** | ⚠️ Community | Workers AI, R2 | Experimental |

---

## Browser Support

For Web UI (planned v1.1):

| Browser | Desktop | Mobile | Status | Minimum Version |
|---------|---------|--------|--------|----------------|
| Chrome | ✅ | ✅ | Fully Supported | 100+ |
| Firefox | ✅ | ✅ | Fully Supported | 100+ |
| Safari | ✅ | ✅ | Fully Supported | 15.4+ |
| Edge | ✅ | ✅ | Fully Supported | 100+ |
| Opera | ✅ | - | Community Support | 85+ |
| Brave | ✅ | ✅ | Community Support | 1.40+ |
| IE 11 | ❌ | - | Not Supported | - |

---

## Tested Configurations

### Primary Test Matrix

These configurations are tested on every release:

| OS | Architecture | Go Version | Terminal | CI Status |
|----|--------------|------------|----------|-----------|
| Ubuntu 22.04 | amd64 | 1.25.5 | GitHub Actions | ✅ |
| Ubuntu 22.04 | amd64 | 1.21.0 | GitHub Actions | ✅ |
| macOS 14 (Sonoma) | arm64 (M3) | 1.25.5 | iTerm2 | ✅ |
| macOS 14 (Sonoma) | amd64 | 1.25.5 | iTerm2 | ✅ |
| Windows 11 | amd64 | 1.25.5 | Windows Terminal | ✅ |
| Docker (Alpine) | amd64 | 1.25.5 | N/A | ✅ |
| Docker (Alpine) | arm64 | 1.25.5 | N/A | ✅ |

### Extended Test Matrix

Additional configurations tested before major releases:

| OS | Architecture | Go Version | Status |
|----|--------------|------------|--------|
| Ubuntu 24.04 | amd64 | 1.25.5 | ✅ Tested |
| Ubuntu 20.04 | amd64 | 1.21.0 | ✅ Tested |
| Debian 12 | arm64 | 1.25.5 | ✅ Tested |
| Fedora 40 | amd64 | 1.25.5 | ✅ Tested |
| RHEL 9 | amd64 | 1.25.5 | ✅ Tested |
| macOS 13 (Ventura) | arm64 | 1.25.5 | ✅ Tested |
| macOS 12 (Monterey) | amd64 | 1.24.0 | ✅ Tested |
| Windows 10 22H2 | amd64 | 1.25.5 | ✅ Tested |
| Windows Server 2022 | amd64 | 1.25.5 | ✅ Tested |

### Community Test Results

Configurations reported working by community members:

- Arch Linux (amd64, rolling release)
- Manjaro (amd64, arm64)
- Raspberry Pi OS (arm64)
- NixOS (amd64)
- FreeBSD 14 (amd64)

---

## Minimum System Requirements

### Hardware

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **CPU** | 1 GHz single-core | 2 GHz multi-core |
| **RAM** | 512 MB | 2 GB+ |
| **Disk Space** | 100 MB | 500 MB |
| **Network** | Internet connection for API calls | Broadband |

### Software

| Component | Requirement |
|-----------|-------------|
| **Operating System** | See OS compatibility above |
| **Go Runtime** | 1.21.0+ (if building from source) |
| **Terminal** | UTF-8 and ANSI color support recommended |
| **Git** | Optional, for version control integration |

---

## Upgrade Path

### Supported Upgrade Paths

| From Version | To Version | Supported | Notes |
|--------------|------------|-----------|-------|
| Beta/Dev | v1.0.0 | ✅ | See migration guide |
| v1.0.x | v1.1.x | ✅ | Minor version upgrade |
| v1.x | v2.0 | ✅ | Major version upgrade (announced) |

### Backward Compatibility

- **Configuration**: v1.x configuration files compatible across v1.x releases
- **Session Data**: Session database migrated automatically
- **API**: Breaking API changes only in major versions

---

## Testing & Certification

### Automated Testing

- **Unit Tests**: Run on every commit
- **Integration Tests**: Run daily
- **E2E Tests**: Run weekly
- **Performance Tests**: Run before releases

### Manual Testing

- **Platforms**: macOS, Linux, Windows tested manually
- **Terminals**: Top 5 terminals tested per platform
- **Providers**: All 6 LLM providers tested before release

### Certification

Currently not certified, but working towards:
- SOC 2 Type II (planned 2026)
- ISO 27001 (planned 2026)
- GDPR compliance (in progress)

---

## Known Incompatibilities

### Not Supported

1. **Operating Systems**:
   - Windows 8.1 and below
   - macOS 10.14 and below
   - Linux kernels < 3.10

2. **Architectures**:
   - 32-bit systems (386, armv6)
   - RISC-V (not yet tested)
   - MIPS architectures

3. **Terminals**:
   - Pure ASCII terminals (no ANSI)
   - Terminals without UTF-8 support

4. **Go Versions**:
   - Go 1.20 and below (missing required features)

---

## Reporting Compatibility Issues

If you encounter compatibility issues:

1. **Check This Document**: Verify your configuration is supported
2. **Search Issues**: Check if already reported
3. **Create Issue**: Report with details:
   - OS and version
   - Architecture
   - Go version
   - Terminal emulator
   - Error messages
   - Steps to reproduce

**Issue Template**: [Compatibility Issue Template](https://github.com/AINative-studio/ainative-code/issues/new?template=compatibility.md)

---

## Future Compatibility

Planned support for upcoming releases:

- **Go 1.26**: Testing planned for Q2 2026
- **Windows ARM64**: Preview in v1.1
- **Alpine Linux 3.20**: Testing planned
- **Raspberry Pi 5**: Testing in progress

---

**Last Updated**: January 4, 2026
**Next Review**: April 1, 2026

For the latest compatibility information, visit: [docs.ainative.studio/code/compatibility](https://docs.ainative.studio/code/compatibility)

---

**Copyright © 2024 AINative Studio. All rights reserved.**
