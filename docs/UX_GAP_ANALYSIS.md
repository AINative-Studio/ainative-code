# UX Gap Analysis: AINative Code vs Crush

**Date**: 2026-01-09
**Purpose**: Identify UX/feature gaps between AINative Code and Crush CLI
**Status**: Analysis Complete

---

## Executive Summary

AINative Code is based on Charm's Crush CLI but is missing several critical UX features that make Crush exceptional. This analysis identifies gaps and provides actionable recommendations for bringing AINative Code to feature parity with Crush.

**Key Finding**: While AINative Code has the foundation, it lacks the polished terminal UX, extensibility, and user-centric features that make Crush a production-ready CLI tool.

---

## Architecture Comparison

### Crush (Reference Implementation)
- **Type**: Go compiled CLI binary
- **UI Framework**: Bubbletea v2 (Terminal UI)
- **Session Management**: SQLite with Ent ORM
- **Extensibility**: MCP (stdio, HTTP, SSE)
- **Code Context**: LSP integration
- **Distribution**: Binary (Homebrew, npm, apt, yum, winget, scoop)
- **Configuration**: JSON with schema validation
- **Logging**: Project-relative `.crush/logs/crush.log`

### AINative Code (Current)
- **Type**: Go CLI application
- **UI Framework**: Cobra + basic terminal I/O
- **Session Management**: SQLite (custom implementation)
- **Extensibility**: MCP (partial implementation)
- **Code Context**: None
- **Distribution**: Manual build only
- **Configuration**: YAML
- **Logging**: Structured JSON to stdout

---

## Critical UX Gaps

### 1. Terminal User Interface (TUI)

#### Crush Has:
- ✅ Full-screen interactive TUI with Bubbletea
- ✅ Real-time streaming responses
- ✅ Syntax highlighting
- ✅ Scrollable history
- ✅ Model switching mid-session
- ✅ Progress indicators
- ✅ Error display in-context
- ✅ Session persistence across restarts

#### AINative Code Has:
- ❌ Basic line-by-line chat (`fmt.Scanln`)
- ❌ No TUI framework
- ❌ Limited visual feedback
- ❌ No scrolling or history navigation
- ❌ Sessions not persistent across restarts
- ❌ No progress indicators
- ❌ No syntax highlighting

**Impact**: **CRITICAL** - User experience is significantly worse

**Recommendation**:
1. Integrate Bubbletea v2 framework
2. Create interactive TUI for chat sessions
3. Add real-time streaming with visual indicators
4. Implement scrollable history view

---

### 2. LSP Integration (Code Context)

#### Crush Has:
- ✅ LSP server integration for code context
- ✅ Configurable per language (Go, TypeScript, Python, Nix, etc.)
- ✅ Environment variable support for LSP config
- ✅ Automatic code analysis before AI responses
- ✅ Diagnostic information from LSPs

#### AINative Code Has:
- ❌ No LSP integration at all
- ❌ No code context awareness
- ❌ AI responses lack codebase understanding

**Impact**: **HIGH** - AI responses lack context about actual code

**Recommendation**:
1. Add LSP client library (e.g., `go-lsp`)
2. Create LSP manager for multiple language servers
3. Add LSP configuration to setup wizard
4. Integrate LSP diagnostics into AI context

**Files to Reference**:
- `/Users/aideveloper/crush/internal/lsp/` - LSP implementation

---

### 3. MCP Extensibility

#### Crush Has:
- ✅ Full MCP support (stdio, HTTP, SSE)
- ✅ MCP server discovery
- ✅ Tool discovery from MCP servers
- ✅ Environment variable expansion in config
- ✅ MCP timeout configuration
- ✅ Disabled server support

#### AINative Code Has:
- ⚠️ Partial MCP implementation
- ❌ No stdio transport
- ❌ No SSE transport
- ❌ Limited tool discovery
- ❌ No environment variable expansion
- ❌ Basic timeout only

**Impact**: **HIGH** - Extensibility is limited

**Recommendation**:
1. Complete MCP stdio transport (internal/cmd/mcp.go:107-112)
2. Add SSE transport support
3. Implement environment variable expansion `$(echo $VAR)`
4. Add MCP server health checks
5. Implement tool discovery from all transports

---

### 4. Permission System

#### Crush Has:
- ✅ Tool permission prompts before execution
- ✅ `allowed_tools` configuration
- ✅ `--yolo` flag to skip all prompts
- ✅ Per-tool permission memory
- ✅ Visual permission dialogs in TUI

#### AINative Code Has:
- ❌ No permission system
- ❌ Tools execute without confirmation
- ❌ No security prompts
- ❌ No allow-list configuration

**Impact**: **CRITICAL** - Security risk

**Recommendation**:
1. Add permission system before tool execution
2. Implement `allowed_tools` config option
3. Add `--yolo` flag for automation
4. Create permission prompt UI
5. Store permission decisions per session

---

### 5. Configuration Management

#### Crush Has:
- ✅ JSON schema validation (`https://charm.land/crush.json`)
- ✅ Priority: `.crush.json` > `crush.json` > `~/.config/crush/crush.json`
- ✅ Hot-reload configuration
- ✅ `crush schema` command to output schema
- ✅ Project-local vs global config

#### AINative Code Has:
- ⚠️ YAML configuration (less validated)
- ⚠️ Single config location (`~/.ainative-code.yaml`)
- ❌ No schema validation
- ❌ No hot-reload
- ❌ No project-local config support

**Impact**: **MEDIUM** - Configuration is less flexible

**Recommendation**:
1. Support project-local `.ainative-code.yaml`
2. Add JSON schema for configuration validation
3. Implement config priority (local > global)
4. Add `config validate` command
5. Support hot-reload during chat sessions

---

### 6. Logging & Debugging

#### Crush Has:
- ✅ Project-relative logs: `./.crush/logs/crush.log`
- ✅ `crush logs` command with `--tail` and `--follow`
- ✅ `--debug` and `--debug-lsp` flags
- ✅ Structured logging with context
- ✅ Separate logs per project

#### AINative Code Has:
- ⚠️ Structured logging (good)
- ❌ Logs to stdout only (not file)
- ❌ No `logs` command
- ❌ No project-relative log files
- ❌ No log tailing/following

**Impact**: **MEDIUM** - Harder to debug issues

**Recommendation**:
1. Add project-relative log directory (`.ainative-code/logs/`)
2. Implement `logs` command with `--tail` and `--follow`
3. Keep structured logging but also write to files
4. Add `--debug` flag for verbose output
5. Rotate logs automatically

---

### 7. Provider Management

#### Crush Has:
- ✅ Auto-update providers from Catwalk (community database)
- ✅ `update-providers` command
- ✅ Disable auto-update option
- ✅ Custom provider base URLs
- ✅ Embedded fallback providers
- ✅ Reset to embedded providers

#### AINative Code Has:
- ⚠️ Hardcoded providers in config
- ❌ No provider auto-update
- ❌ No provider database integration
- ❌ Manual provider configuration only

**Impact**: **MEDIUM** - Users must manually track model updates

**Recommendation**:
1. Consider integrating Catwalk or similar provider database
2. Add `update-providers` command
3. Implement auto-update with opt-out
4. Support embedded provider fallback

---

### 8. Attribution & Git Integration

#### Crush Has:
- ✅ Configurable attribution in commits
- ✅ `Co-Authored-By: Crush <crush@charm.land>`
- ✅ `💘 Generated with Crush` footer
- ✅ Attribution in PR descriptions
- ✅ `attribution` config options

#### AINative Code Has:
- ⚠️ Basic git commit support
- ❌ No attribution in commits
- ❌ No PR description generation
- ❌ No attribution configuration

**Impact**: **LOW** - Nice to have, not critical

**Recommendation**:
1. Add configurable attribution to commits
2. Add `Co-Authored-By` option
3. Support PR description generation
4. Make attribution opt-in/opt-out

---

### 9. Ignore Files & Context Management

#### Crush Has:
- ✅ Respects `.gitignore` by default
- ✅ Additional `.crushignore` file support
- ✅ Same syntax as `.gitignore`
- ✅ Context path configuration
- ✅ Directory-level ignore files

#### AINative Code Has:
- ❌ No file ignore system
- ❌ No `.ainative-codeignore` support
- ❌ No context path configuration

**Impact**: **MEDIUM** - AI context may include unwanted files

**Recommendation**:
1. Respect `.gitignore` by default
2. Add `.ainative-codeignore` support
3. Implement context path filtering
4. Add `context_paths` configuration option

---

### 10. Session Management

#### Crush Has:
- ✅ SQLite database with Ent ORM
- ✅ Named sessions
- ✅ Session switching mid-chat
- ✅ Session history with timestamps
- ✅ Session export/import
- ✅ Session search

#### AINative Code Has:
- ⚠️ SQLite database (custom)
- ⚠️ Basic session list/view
- ❌ No mid-session switching
- ❌ Limited session search (FTS5 but basic)
- ❌ No session import/export

**Impact**: **MEDIUM** - Session management is less polished

**Recommendation**:
1. Add session switching without exiting chat
2. Enhance session search with better UX
3. Add session export/import commands
4. Improve session listing UI

---

### 11. Distribution & Packaging

#### Crush Has:
- ✅ Homebrew (macOS)
- ✅ npm (cross-platform)
- ✅ apt/yum (Linux)
- ✅ winget/scoop (Windows)
- ✅ Arch AUR
- ✅ Nix (NUR)
- ✅ GoReleaser CI/CD
- ✅ Signed binaries
- ✅ Multi-platform releases

#### AINative Code Has:
- ❌ No package managers
- ❌ Manual build only
- ⚠️ GitHub Actions (but no releases)
- ❌ No signed binaries
- ❌ No distribution strategy

**Impact**: **HIGH** - Hard for users to install

**Recommendation**:
1. Set up GoReleaser for automated releases
2. Publish to Homebrew tap
3. Create npm wrapper package
4. Add apt/yum repositories
5. Support winget/scoop on Windows
6. Sign macOS binaries

---

### 12. Onboarding & First-Run Experience

#### Crush Has:
- ✅ Interactive API key prompt on first run
- ✅ Guided provider selection
- ✅ Auto-detect environment variables
- ✅ Helpful error messages with examples
- ✅ Quick start guide in README

#### AINative Code Has:
- ⚠️ Setup wizard (good)
- ⚠️ Guided configuration
- ❌ Runs setup every time (issue #105 - fixed)
- ❌ Less polished prompts
- ❌ No auto-detection of existing keys

**Impact**: **MEDIUM** - First-run could be smoother

**Recommendation**:
1. Auto-detect API keys from environment
2. Skip wizard if keys already configured
3. Improve setup wizard UI with Bubbletea
4. Add quick start examples
5. Better error messages with fix suggestions

---

## Feature Comparison Matrix

| Feature | Crush | AINative Code | Priority | Effort |
|---------|-------|---------------|----------|--------|
| **TUI Framework** | ✅ Bubbletea | ❌ None | **CRITICAL** | **HIGH** |
| **LSP Integration** | ✅ Full | ❌ None | **HIGH** | **HIGH** |
| **MCP (stdio)** | ✅ Full | ❌ Partial | **HIGH** | **MEDIUM** |
| **MCP (SSE)** | ✅ Full | ❌ None | **MEDIUM** | **MEDIUM** |
| **Permission System** | ✅ Full | ❌ None | **CRITICAL** | **MEDIUM** |
| **Project-local Config** | ✅ Yes | ❌ No | **MEDIUM** | **LOW** |
| **File Logging** | ✅ Yes | ❌ No | **MEDIUM** | **LOW** |
| **Logs Command** | ✅ Yes | ❌ No | **MEDIUM** | **LOW** |
| **Provider Auto-update** | ✅ Yes | ❌ No | **MEDIUM** | **MEDIUM** |
| **Ignore Files** | ✅ Yes | ❌ No | **MEDIUM** | **LOW** |
| **Session Switching** | ✅ Yes | ❌ No | **LOW** | **LOW** |
| **Attribution Config** | ✅ Yes | ❌ No | **LOW** | **LOW** |
| **Distribution** | ✅ Full | ❌ None | **HIGH** | **MEDIUM** |
| **Schema Validation** | ✅ Yes | ❌ No | **LOW** | **LOW** |
| **Session Export** | ✅ Yes | ❌ No | **LOW** | **LOW** |

---

## Prioritized Roadmap

### Phase 1: Critical UX Improvements (2-3 weeks)
**Goal**: Match basic Crush UX quality

1. **Integrate Bubbletea TUI** (1 week)
   - Add bubbletea dependency
   - Create interactive chat UI
   - Implement streaming display
   - Add scrollable history

2. **Add Permission System** (3-5 days)
   - Tool execution confirmation prompts
   - `allowed_tools` configuration
   - `--yolo` flag for automation
   - Permission decision persistence

3. **Improve Logging** (2-3 days)
   - Project-relative log files
   - `logs` command with `--tail`/`--follow`
   - Keep structured logging

### Phase 2: Code Context & Extensibility (3-4 weeks)
**Goal**: Add intelligent code awareness

1. **LSP Integration** (2 weeks)
   - Add LSP client library
   - Implement LSP manager
   - Add LSP configuration
   - Integrate diagnostics into context

2. **Complete MCP Implementation** (1 week)
   - stdio transport
   - SSE transport
   - Environment variable expansion
   - Health checks

3. **Ignore Files** (2-3 days)
   - Respect `.gitignore`
   - Add `.ainative-codeignore`
   - Context path filtering

### Phase 3: Distribution & Polish (2-3 weeks)
**Goal**: Professional distribution

1. **GoReleaser Setup** (1 week)
   - Configure GoReleaser
   - GitHub Actions for releases
   - Multi-platform builds
   - Code signing

2. **Package Managers** (1 week)
   - Homebrew tap
   - npm wrapper
   - apt/yum repos (if applicable)

3. **Configuration Enhancements** (3-5 days)
   - Project-local config support
   - JSON schema validation
   - Config priority system
   - Hot-reload support

### Phase 4: Advanced Features (2-3 weeks)
**Goal**: Feature parity with Crush

1. **Provider Management** (1 week)
   - Provider database integration
   - Auto-update system
   - `update-providers` command

2. **Session Enhancements** (1 week)
   - Mid-session switching
   - Session export/import
   - Enhanced search UI

3. **Polish & Testing** (1 week)
   - Comprehensive testing
   - Documentation updates
   - Bug fixes

**Total Estimated Time: 9-13 weeks**

---

## Immediate Action Items

### This Week
1. ✅ Fix all 10 open issues (DONE)
2. ⬜ Integrate Bubbletea for TUI
3. ⬜ Add basic permission prompts
4. ⬜ Set up project-relative logging

### Next Week
1. ⬜ LSP integration research
2. ⬜ Complete MCP stdio transport
3. ⬜ Set up GoReleaser

### This Month
1. ⬜ Full TUI with streaming
2. ⬜ LSP integration complete
3. ⬜ First proper release (v0.2.0)

---

## Conclusion

**Current State**: AINative Code has solid foundations but lacks the polished UX and extensibility of Crush.

**Gap Summary**:
- **CRITICAL Gaps**: TUI, Permission System
- **HIGH Priority**: LSP Integration, MCP completion, Distribution
- **MEDIUM Priority**: Logging, Config management, Ignore files
- **LOW Priority**: Attribution, Session export, Provider auto-update

**Recommended Focus**: Prioritize Phase 1 (TUI + Permissions) immediately to match basic Crush UX, then tackle LSP and MCP for feature parity.

**Success Metrics**:
- Users can install via package manager (Homebrew)
- Interactive TUI matches Crush quality
- Tools execute with permission prompts
- Code context from LSP improves AI responses
- Project logs are accessible via `logs` command

---

## References

**Crush Files**:
- `/Users/aideveloper/crush/README.md` - Features overview
- `/Users/aideveloper/crush/QUICK_FACTS.md` - Architecture comparison
- `/Users/aideveloper/crush/HYBRID_APPROACH_SUMMARY.md` - Integration strategies
- `/Users/aideveloper/crush/internal/llm/agent/agent.go` - Core agent (1133 lines)
- `/Users/aideveloper/crush/internal/cmd/root.go` - CLI structure

**AINative Code Files**:
- `/Users/aideveloper/AINative-Code/internal/cmd/chat.go` - Current chat implementation
- `/Users/aideveloper/AINative-Code/internal/cmd/mcp.go` - MCP implementation
- `/Users/aideveloper/AINative-Code/internal/setup/wizard.go` - Setup wizard

---

**Document Version**: 1.0
**Last Updated**: 2026-01-09
**Status**: Ready for Review
