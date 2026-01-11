# Changelog

All notable changes to AINative Code will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - v1.0.0

## [0.1.9] - 2026-01-10

### Fixed

#### Critical/P0 Fixes
- **#117 - Setup wizard model validation** - Fixed setup wizard offering outdated Claude 3.5 models that chat command rejects
  - Updated to Claude 4.5 models (claude-sonnet-4-5-20250929, claude-haiku-4-5-20251001, claude-opus-4-1)
  - Set default to latest recommended model (claude-sonnet-4-5-20250929)
  - Added 9 comprehensive tests for model synchronization
  - Files: `internal/setup/prompts.go`, `internal/setup/wizard.go`

#### P1/High Priority Fixes
- **#125 - ZeroDB config path mismatch** - Fixed configuration path inconsistency between setup wizard and ZeroDB commands
  - Standardized to `services.zerodb.*` configuration paths
  - Updated environment variable mapping to `AINATIVE_CODE_SERVICES_ZERODB_*`
  - Added 8 integration tests for end-to-end workflow validation
  - File: `internal/cmd/zerodb_table.go`

- **#120 - MCP server persistence** - Fixed MCP servers not being persisted to disk
  - Implemented persistent storage with `~/.mcp.json` configuration file
  - Added atomic writes with rollback mechanism for data integrity
  - Created ConfigManager for centralized MCP configuration
  - Added 65 comprehensive tests (51 unit + 14 integration)
  - New files: `internal/mcp/config.go`, updated `internal/mcp/registry.go`

#### Medium Priority Fixes
- **#126 - Meta Llama provider validation** - Added Meta Llama to supported provider validation
  - Implemented `ValidateMetaLlamaKey()` method with proper API key validation
  - Added support for both "meta_llama" and "meta" provider aliases
  - All 5 providers now have consistent validation support
  - File: `internal/setup/validation.go`

- **#122 - Config validation provider check** - Fixed config validate command checking wrong provider field
  - Updated validation to check `llm.default_provider` (new structure) instead of root `provider` field
  - Added support for all 8 providers (anthropic, openai, google, bedrock, azure, ollama, meta_llama, meta)
  - Maintained backward compatibility with legacy root provider field
  - Added 13 comprehensive test cases
  - File: `internal/cmd/config.go`

- **#121 - Flag naming inconsistency** - Standardized file output flags across all commands
  - Unified all commands to use `-f, --file` flag (previously mixed -f and -o)
  - Added backward compatibility with deprecation warnings for old `-o` flag
  - Updated commands: `rlhf export`, `design extract`, `design generate`, `session export`
  - Added file input flag to `design validate` command
  - Created 28 integration tests for flag standardization
  - Files: Multiple command files in `internal/cmd/`

#### Low Priority Fixes
- **#119 - Chat empty message validation** - Added local validation for empty chat messages
  - Validates messages before API calls to save costs and improve UX
  - Catches empty strings, whitespace-only, tabs, newlines, and mixed whitespace
  - 500x faster error response (<1ms vs ~500ms)
  - Added 10 comprehensive tests
  - File: `internal/cmd/chat.go`

- **#123 - Session list negative limit validation** - Added validation for session list limit parameter
  - Rejects negative and zero limit values with clear error message
  - Prevents potential performance issues with unlimited queries
  - Added comprehensive unit and E2E tests
  - File: `internal/cmd/session.go`

- **#110 - Config file existence validation** - Added validation for explicitly provided config files
  - Validates file existence and type when `--config` flag is used
  - Provides clear error messages for missing files or directories
  - Maintains graceful fallback for default config paths
  - Added 13 comprehensive tests
  - File: `internal/cmd/root.go`

### Testing
- Added 200+ new tests across all bug fixes
- 100% test pass rate for all new functionality
- Comprehensive integration and E2E test coverage
- All fixes are backward compatible with zero breaking changes

### Documentation
- Created 25+ comprehensive fix reports and guides
- Detailed technical documentation for each bug fix
- Quick reference guides for developers
- Executive summaries for stakeholders

### Added

#### Authentication & Security
- **Hybrid Authentication System** - JWT/OAuth 2.0 for AINative services, API keys for LLM providers
- **JWT Token Structures** (TASK-040) - Complete JWT implementation with claims validation
- **Local JWT Validation** (TASK-041) - Public key caching for offline validation
- **API Token Validation** (TASK-042) - Token validation with fallback mechanisms
- **OAuth 2.0 PKCE Flow** (TASK-044) - Secure authorization code flow with proof key
- **Token Auto-Refresh Manager** (TASK-045) - Automatic token refresh before expiration
- **OS Keychain Integration** (TASK-046) - Secure credential storage using system keychain
- **Authentication CLI Commands** (TASK-047) - Full suite of auth commands (login, logout, refresh, status)
- **Local Authentication Fallback** (TASK-043) - Fallback authentication when network is unavailable
- **Security Configuration** - Encryption, TLS, CORS, and secret rotation support

#### LLM Providers
- **Anthropic Claude Integration** - Complete API integration with streaming support
- **OpenAI GPT Support** - GPT-4 and GPT-3.5 with organization settings
- **Google Gemini Integration** - Gemini Pro with Vertex AI support
- **AWS Bedrock Support** - Claude and other models on AWS infrastructure
- **Azure OpenAI Integration** - Azure-hosted OpenAI models with deployment management
- **Ollama Local LLM Support** - Local model execution for open-source models
- **Fallback System** - Automatic provider switching with configurable retry logic
- **Prompt Caching** (TASK-075) - Cache prompts for reduced costs and latency
- **Provider Configuration** - Comprehensive configuration system for all 6 providers

#### Session Management
- **Session Storage** - SQLite-based local session management
- **Session Persistence** - Resume conversations across CLI sessions
- **Session History** - Complete conversation history with timestamps
- **Multi-Session Support** - Manage multiple concurrent sessions

#### Tools & Integrations

##### Model Context Protocol (MCP)
- **MCP Server Protocol** (TASK-070) - Complete MCP protocol implementation
- **MCP Server Management CLI** (TASK-071) - Commands for server lifecycle management
- **MCP Resource Access** - Browse and read resources from MCP servers
- **MCP Tool Invocation** - Execute tools provided by MCP servers
- **MCP Prompt Templates** - Use server-provided prompt templates

##### ZeroDB Platform Integration
- **AINative API Client** (TASK-050) - Core HTTP client with JWT authentication
- **ZeroDB NoSQL CLI** (TASK-052) - Complete CRUD operations for NoSQL tables
- **ZeroDB Vector Operations** (TASK-051) - Vector upsert, search, and management
- **ZeroDB Agent Memory CLI** (TASK-053) - Semantic memory storage and retrieval
- **ZeroDB Quantum Features** (TASK-054) - Vector entanglement, compression, and quantum search
- **ZeroDB File Operations** (TASK-055) - File upload, download, and management
- **ZeroDB Event Stream** (TASK-056) - Event creation and listing
- **ZeroDB Project Management** (TASK-057) - Project info, stats, and logs
- **PostgreSQL Provisioning** (TASK-058) - Dedicated PostgreSQL instance management
- **RLHF Feedback Collection** (TASK-059) - Reinforcement learning feedback submission

##### Design & Content Management
- **Design Token Operations** (TASK-060) - Extract, sync, and validate design tokens
- **Design Code Generation** (TASK-061) - Generate UI code from design tokens
- **Strapi Content Operations** (TASK-062) - CMS content management and blog publishing
- **Strapi Blog Publishing** (TASK-063) - Blog post creation, updates, and management

##### Google Analytics Integration
- **GA4 Data Retrieval** - Query Google Analytics 4 data with smart protections
- **GA Schema Search** - Search for dimensions and metrics
- **GA Quick Reports** - Generate analytics overview reports
- **GA Category Listing** - Browse available dimensions and metrics by category

#### TUI/CLI Features
- **Beautiful Bubble Tea TUI** - Sophisticated terminal interface with rich formatting
- **Chat Mode** - Interactive chat with streaming AI responses
- **Code Generation** - Generate code from natural language prompts
- **Syntax Highlighting** - Rich code formatting in terminal
- **Extended Thinking Visualization** (TASK-074) - Real-time visualization of extended thinking process
- **Progress Indicators** - Spinners and progress bars for long-running operations
- **Table Formatting** - Tabular display for structured data
- **JSON Output Support** - Machine-readable JSON output for all commands
- **Color-Coded Output** - Context-aware color coding for better readability

#### Configuration
- **Multi-Source Configuration** (TASK-005) - File, environment variables, and command flags
- **YAML Configuration Schema** - Comprehensive schema for all settings
- **Environment Variable Support** - `AINATIVE_CODE_*` prefix for all settings
- **Example Configuration** - Production-ready example with extensive documentation
- **Configuration Validation** - Comprehensive validation with clear error messages
- **Secret Management** - Environment-based secret loading
- **Hot Reload Support** - Watch and reload configuration changes

#### Performance
- **Caching System** - Multi-backend support (memory, Redis, Memcached)
- **Rate Limiting** - Request throttling and burst control
- **Concurrency Control** - Worker pool and queue management
- **Circuit Breaker** - Failure prevention and recovery mechanisms
- **Connection Pooling** - Efficient connection reuse for all services
- **Request Batching** - Batch similar requests for efficiency

#### Logging & Monitoring
- **Structured Logging System** (TASK-009) - Production-ready logging with zerolog
- **Multiple Log Levels** - DEBUG, INFO, WARN, ERROR, FATAL with configurable minimum
- **Output Formats** - JSON and text (console) formats with color support
- **Log Rotation** - Automatic rotation by size, age, and backup count
- **Context-Aware Logging** - Automatic request ID, session ID, and user ID extraction
- **Performance Logging** - Sub-microsecond overhead for disabled levels
- **Sensitive Data Filtering** - Prevent logging of secrets and PII

#### Development & DevOps
- **Comprehensive CI/CD Pipeline** (TASK-003) - GitHub Actions for testing, building, and releasing
- **Multi-Platform Builds** - macOS (Intel/ARM), Linux (amd64/arm64), Windows (amd64)
- **Docker Support** - Multi-platform Docker images with security hardening
- **Automated Testing** - Unit, integration, and benchmark tests
- **Code Quality Checks** - golangci-lint with 40+ rules
- **Security Scanning** - gosec and govulncheck integration
- **Coverage Reporting** - Codecov integration with 80% threshold
- **Dependency Management** - Automated weekly dependency updates
- **Release Automation** - Automated GitHub releases with changelogs and assets

#### Documentation
- **Comprehensive User Guide** - Complete documentation for all features
- **API Reference** - Detailed API documentation
- **Architecture Guide** - System design and technical architecture
- **Development Guide** - Contributing and development setup
- **Configuration Guide** (TASK-005) - 800+ lines of configuration documentation
- **Logging Guide** (TASK-009) - Complete logging documentation with examples
- **CI/CD Guide** (TASK-003) - Comprehensive CI/CD documentation
- **Database Guide** - SQLite schema and migration documentation
- **Examples** - Real-world code examples and use cases
- **ZeroDB Documentation** - Complete guides for all ZeroDB features
- **Quantum Features Guide** - Comprehensive quantum operations documentation

### Changed

#### Branding
- **Renamed to AINative Code** (TASK-002) - Complete rebrand from previous naming
- **Brand Identity** - Tagline: "AI-Native Development, Natively"
- **Brand Colors** - Professional color palette (Indigo, Purple, Green, Red)
- **Copyright Notice** - "© 2024 AINative Studio. All rights reserved."

#### Architecture
- **Modular Package Structure** - Clean separation of concerns across packages
- **Error Handling Framework** (TASK-011) - Comprehensive error types and handling
- **Configuration System** - Centralized configuration with validation
- **Client Architecture** - Shared HTTP client with consistent error handling

### Fixed
- **Thread Safety** - Mutex protection for global logger and shared state
- **Memory Leaks** - Proper resource cleanup in HTTP clients
- **Context Propagation** - Consistent context passing through call chains
- **Error Messages** - Clear, actionable error messages throughout

### Security
- **Keychain Integration** - Secure OS-level credential storage
- **Secret Rotation** - Support for automatic secret rotation
- **Path Validation** - Prevent path traversal attacks
- **Command Allowlists** - Security restrictions for terminal tool
- **TLS Configuration** - Secure TLS with certificate validation
- **Input Validation** - Comprehensive validation of all user inputs

### Performance
- **Zero-Allocation Logging** - Sub-microsecond overhead for disabled log levels
- **Connection Pooling** - Reuse HTTP connections across requests
- **Caching** - Multi-tier caching for API responses
- **Compression** - Quantum compression for vector storage optimization
- **Batch Operations** - Reduce API calls through batching

### Dependencies
- **Go 1.25.5** - Latest stable Go version with generics and performance improvements
- **Bubble Tea v1.3.10** - Terminal UI framework
- **Cobra v1.10.2** - CLI framework
- **Viper v1.21.0** - Configuration management
- **zerolog v1.34.0** - High-performance structured logging
- **lumberjack v2.2.1** - Log rotation
- **Anthropic SDK v1.19.0** - Claude API client
- **JWT v5.3.0** - JWT token handling
- **SQLite v1.14.32** - Local database

## [0.1.0] - 2024-01-07

### Added
- Initial project setup
- Core package structure
- Basic error handling framework (TASK-011)
- Configuration schema (TASK-005)
- CI/CD pipeline (TASK-003)
- Branding system (TASK-002)
- Logging infrastructure (TASK-009)
- Multi-provider LLM support (Anthropic, OpenAI, Gemini, Bedrock, Azure, Ollama)
- Comprehensive authentication system with JWT/OAuth 2.0
- Session management and persistence
- Beautiful Bubble Tea TUI with syntax highlighting
- ZeroDB platform integration (vectors, NoSQL, files, events)
- LSP client integration for code intelligence
- Full-text conversation search and export
- Production-ready structured logging with rotation
- Complete documentation suite (user guide, API reference, architecture)
- Multi-platform CI/CD pipeline with automated releases
- Security audit and CVE vulnerability fixes

---

## Release Notes

For detailed release notes, see [docs/releases/v1.0-release-notes.md](docs/releases/v1.0-release-notes.md)

## Migration Guide

For upgrade instructions, see [docs/releases/migration-guide.md](docs/releases/migration-guide.md)

## Links

- **GitHub Repository**: https://github.com/AINative-studio/ainative-code
- **Documentation**: https://docs.ainative.studio/code
- **Issues**: https://github.com/AINative-studio/ainative-code/issues
- **Releases**: https://github.com/AINative-studio/ainative-code/releases

[1.0.0]: https://github.com/AINative-studio/ainative-code/releases/tag/v1.0.0
[0.1.9]: https://github.com/AINative-studio/ainative-code/releases/tag/v0.1.9
[0.1.8]: https://github.com/AINative-studio/ainative-code/releases/tag/v0.1.8
[0.1.0]: https://github.com/AINative-studio/ainative-code/releases/tag/v0.1.0
