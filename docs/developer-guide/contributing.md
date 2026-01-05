# Contributing Guide

## Overview

Thank you for your interest in contributing to AINative Code! This guide will walk you through the contribution process, from finding an issue to getting your changes merged.

## Getting Started

### Before You Start

1. **Read the Documentation**:
   - [README.md](../../README.md) - Project overview
   - [Architecture Guide](architecture.md) - System architecture
   - [Code Style Guide](code-style.md) - Coding standards

2. **Set Up Your Environment**:
   - Follow [Development Setup](setup.md)
   - Install all required tools
   - Verify your setup with `make test`

3. **Find an Issue**:
   - Browse [GitHub Issues](https://github.com/AINative-studio/ainative-code/issues)
   - Look for `good-first-issue` or `help-wanted` labels
   - Check that the issue isn't already assigned

## Git Workflow

### 1. Fork and Clone

```bash
# Fork the repository on GitHub first

# Clone your fork
git clone https://github.com/YOUR_USERNAME/ainative-code.git
cd ainative-code

# Add upstream remote
git remote add upstream https://github.com/AINative-studio/ainative-code.git

# Verify remotes
git remote -v
```

### 2. Stay in Sync

```bash
# Fetch latest changes from upstream
git fetch upstream

# Update your main branch
git checkout main
git merge upstream/main
git push origin main
```

### 3. Create a Feature Branch

```bash
# Create and checkout a new branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

**Branch Naming Conventions**:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions or changes
- `chore/` - Build process or tooling changes

**Examples**:
- `feature/add-openai-provider`
- `fix/config-validation-error`
- `docs/update-installation-guide`
- `refactor/simplify-auth-flow`

### 4. Make Your Changes

```bash
# Make changes to the code
# ... edit files ...

# Check what changed
git status
git diff

# Add changes to staging
git add .

# Or add specific files
git add internal/provider/openai.go
```

### 5. Commit Your Changes

```bash
# Commit with a descriptive message
git commit -m "feat: add OpenAI provider support"

# Or commit with detailed message
git commit
# Opens editor for multi-line commit message
```

See [Commit Message Guidelines](#commit-message-guidelines) below.

### 6. Push to Your Fork

```bash
# Push your branch to your fork
git push origin feature/your-feature-name

# If you need to update a pushed branch
git push --force-with-lease origin feature/your-feature-name
```

### 7. Create a Pull Request

1. Go to your fork on GitHub
2. Click "Compare & pull request"
3. Fill out the PR template
4. Submit the pull request

See [Pull Request Process](#pull-request-process) below.

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature for the user
- `fix`: Bug fix for the user
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring (no functional changes)
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `build`: Changes to build system or dependencies
- `ci`: CI/CD configuration changes
- `chore`: Other changes (tooling, etc.)
- `revert`: Reverts a previous commit

### Scope (Optional)

Common scopes in our project:
- `auth` - Authentication and authorization
- `provider` - LLM provider integrations
- `tui` - Terminal UI
- `config` - Configuration management
- `logger` - Logging system
- `database` - Database operations
- `cli` - Command-line interface
- `tools` - Tool system
- `mcp` - MCP server
- `client` - Platform clients (zerodb, strapi, etc.)

### Examples

**Feature Addition**:
```
feat(provider): add OpenAI provider support

Implement OpenAI provider with streaming support and error handling.
Includes comprehensive tests and documentation.

Closes #123
```

**Bug Fix**:
```
fix(auth): correct token expiration check

The JWT expiration was being checked incorrectly, causing valid
tokens to be rejected. Fixed the time comparison logic.

Fixes #456
```

**Documentation**:
```
docs(readme): update installation instructions

Add Homebrew installation method and improve Docker instructions
with troubleshooting steps.
```

**Refactoring**:
```
refactor(config): simplify resolver implementation

Extracted command resolution into separate function and added
better error messages. No functional changes.
```

**Breaking Change**:
```
feat(provider)!: change provider interface signature

BREAKING CHANGE: The Chat method now requires context.Context as
the first parameter. This provides better timeout and cancellation
support.

Migration guide:
- Before: provider.Chat(messages)
- After: provider.Chat(ctx, messages)
```

### Commit Message Tips

1. **Use imperative mood**: "add feature" not "added feature"
2. **Keep subject line under 72 characters**
3. **Capitalize subject line**
4. **No period at the end of subject**
5. **Separate subject from body with blank line**
6. **Wrap body at 72 characters**
7. **Explain what and why, not how**

## Pull Request Process

### Before Opening a PR

1. **Ensure Tests Pass**:
   ```bash
   make test
   ```

2. **Check Code Coverage**:
   ```bash
   make test-coverage
   # Ensure coverage meets 80% threshold
   ```

3. **Run Linter**:
   ```bash
   make lint
   ```

4. **Format Code**:
   ```bash
   make fmt
   ```

5. **Run All CI Checks**:
   ```bash
   make ci
   ```

6. **Update Documentation**:
   - Update relevant docs in `/docs`
   - Add code examples if needed
   - Update README.md if necessary

7. **Add Tests**:
   - Write unit tests for new code
   - Add integration tests if needed
   - Ensure tests are comprehensive

8. **Update CHANGELOG**:
   - Add entry under "Unreleased" section
   - Follow existing format

### PR Title

Use the same format as commit messages:

```
feat(provider): add OpenAI provider support
fix(auth): correct token expiration check
docs(readme): update installation instructions
```

### PR Description Template

```markdown
## Summary
Brief description of the changes

## Motivation
Why is this change needed? What problem does it solve?

## Changes
- List of specific changes made
- One change per bullet point
- Include any breaking changes

## Testing
Describe how you tested these changes:
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Screenshots (if applicable)
Add screenshots for UI changes

## Related Issues
Fixes #123
Relates to #456

## Checklist
- [ ] Tests added and passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Code follows style guide
- [ ] All CI checks passing
- [ ] Reviewed own code
```

### PR Review Process

1. **Automated Checks**:
   - Tests must pass
   - Linter must pass
   - Coverage must meet threshold
   - Build must succeed

2. **Code Review**:
   - At least one maintainer approval required
   - Address all review comments
   - Push new commits to the same branch

3. **Merging**:
   - Maintainer will merge once approved
   - PR will be squashed or merged based on preference
   - Branch will be deleted after merge

### Addressing Review Comments

```bash
# Make changes based on feedback
# ... edit files ...

# Commit changes
git add .
git commit -m "address review comments"

# Push to update PR
git push origin feature/your-feature-name
```

### Keeping PR Up to Date

```bash
# Fetch latest changes
git fetch upstream

# Rebase your branch on latest main
git checkout feature/your-feature-name
git rebase upstream/main

# Resolve any conflicts
# ... fix conflicts ...
git add .
git rebase --continue

# Force push (with lease for safety)
git push --force-with-lease origin feature/your-feature-name
```

## Code Review Guidelines

### For Authors

1. **Keep PRs Small**: Aim for < 400 lines of changes
2. **Single Responsibility**: One feature or fix per PR
3. **Self-Review**: Review your own code first
4. **Add Context**: Explain why, not just what
5. **Respond Promptly**: Address comments within 48 hours

### For Reviewers

1. **Be Respectful**: Constructive feedback only
2. **Explain Why**: Don't just say "change this"
3. **Acknowledge Good Work**: Praise good solutions
4. **Ask Questions**: "What if...?" instead of "This is wrong"
5. **Review Timely**: Respond within 2 business days

### Review Checklist

- [ ] Code follows style guide
- [ ] Tests are comprehensive and passing
- [ ] Documentation is updated
- [ ] No obvious bugs or security issues
- [ ] Performance considerations addressed
- [ ] Error handling is appropriate
- [ ] Code is maintainable and readable

## Issue Triage

### Creating Issues

When creating an issue, use the appropriate template:

**Bug Report**:
```markdown
## Description
Clear description of the bug

## Steps to Reproduce
1. Step one
2. Step two
3. ...

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: macOS 14.0
- Go version: 1.21.0
- AINative Code version: 1.0.0

## Additional Context
Any other relevant information
```

**Feature Request**:
```markdown
## Feature Description
What feature you want added

## Use Case
Why is this feature needed?

## Proposed Solution
How should this be implemented?

## Alternatives Considered
What other approaches did you consider?

## Additional Context
Any other relevant information
```

### Issue Labels

- `bug` - Something isn't working
- `enhancement` - New feature or request
- `documentation` - Improvements to documentation
- `good-first-issue` - Good for newcomers
- `help-wanted` - Extra attention needed
- `question` - Further information requested
- `wontfix` - Won't be worked on
- `duplicate` - Duplicate issue
- `priority:high` - High priority
- `priority:medium` - Medium priority
- `priority:low` - Low priority

## Documentation Standards

### Code Documentation

```go
// AuthenticateUser validates user credentials and returns a JWT token.
// It returns an error if authentication fails or if the credentials are invalid.
//
// The token is valid for 24 hours and includes the user ID and email in claims.
//
// Example:
//
//	token, err := AuthenticateUser("user@example.com", "password")
//	if err != nil {
//	    return fmt.Errorf("authentication failed: %w", err)
//	}
//
func AuthenticateUser(email, password string) (string, error) {
    // Implementation
}
```

### Package Documentation

```go
// Package provider defines the interface for LLM providers and includes
// implementations for various AI providers like Anthropic, OpenAI, and others.
//
// The main interface is Provider, which defines methods for chat completion
// and streaming responses. Each provider implementation handles authentication,
// API communication, and error handling specific to that provider.
//
// Example usage:
//
//	provider := anthropic.NewProvider(config)
//	response, err := provider.Chat(ctx, messages)
//
package provider
```

### User Documentation

When adding features, update:
- User guide in `/docs/user-guide`
- API reference in `/docs/api-reference`
- Examples in `/docs/examples`
- Main README.md if user-facing

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes (v1.0.0 → v2.0.0)
- **MINOR**: New features, backward compatible (v1.0.0 → v1.1.0)
- **PATCH**: Bug fixes, backward compatible (v1.0.0 → v1.0.1)

### Release Steps (Maintainers Only)

1. **Update Version**:
   - Update CHANGELOG.md
   - Update version in documentation

2. **Create Tag**:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```

3. **GitHub Actions**:
   - Automatically builds binaries
   - Creates GitHub release
   - Publishes Docker images
   - Updates Homebrew formula

4. **Post-Release**:
   - Announce in Discussions
   - Update documentation site
   - Close milestone on GitHub

## Community Guidelines

### Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Help others when you can
- Stay on topic in discussions
- Report inappropriate behavior

### Getting Help

1. **Search First**: Check docs and existing issues
2. **Ask in Discussions**: For questions and help
3. **Create Issue**: For bugs or feature requests
4. **Join Community**: Participate in discussions

### Recognition

Contributors are recognized in:
- Release notes
- CONTRIBUTORS.md (coming soon)
- Special mentions for significant contributions

## Tips for Success

### First Contributions

1. Start with `good-first-issue` labels
2. Read all related documentation
3. Ask questions if anything is unclear
4. Keep initial PRs small and focused
5. Be patient with the review process

### Writing Quality Code

1. Follow the [Code Style Guide](code-style.md)
2. Write comprehensive tests
3. Document your code
4. Keep functions small and focused
5. Handle errors gracefully

### Effective Communication

1. Be clear and concise
2. Provide context and examples
3. Be responsive to feedback
4. Ask for help when needed
5. Thank reviewers for their time

## Resources

- [Git Best Practices](https://git-scm.com/book/en/v2)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Code Review Best Practices](https://google.github.io/eng-practices/review/)

## Questions?

If you have questions about contributing:
1. Check the [FAQ](../user-guide/faq.md)
2. Search [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)
3. Create a new Discussion
4. Email: support@ainative.studio

---

**Thank you for contributing to AINative Code!**

Copyright © 2024-2025 AINative Studio. All rights reserved.
