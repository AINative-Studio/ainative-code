# Git Workflow

This document describes the Git workflow and version control practices for the AINative Code project.

## Table of Contents

- [Branch Strategy](#branch-strategy)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)
- [Common Git Tasks](#common-git-tasks)
- [Git Hooks](#git-hooks)
- [Best Practices](#best-practices)

## Branch Strategy

### Main Branches

#### `main`
- Production-ready code
- Always stable and deployable
- Protected branch (requires PR reviews)
- Tagged for releases

#### `develop`
- Integration branch for features
- Latest delivered development changes
- Source for release branches

### Supporting Branches

#### Feature Branches
```bash
# Naming: feature/<issue-number>-<short-description>
feature/123-add-user-authentication
feature/456-improve-logging
```

**Purpose**: Develop new features or enhancements

**Lifecycle**:
- Branch from: `develop`
- Merge into: `develop`
- Delete after merge

**Create feature branch**:
```bash
git checkout develop
git pull origin develop
git checkout -b feature/123-add-user-authentication
```

#### Bugfix Branches
```bash
# Naming: bugfix/<issue-number>-<short-description>
bugfix/789-fix-memory-leak
bugfix/101-correct-config-validation
```

**Purpose**: Fix bugs in development

**Lifecycle**:
- Branch from: `develop`
- Merge into: `develop`
- Delete after merge

#### Hotfix Branches
```bash
# Naming: hotfix/<version>-<description>
hotfix/1.2.1-critical-security-fix
hotfix/1.3.1-database-connection-issue
```

**Purpose**: Emergency fixes for production

**Lifecycle**:
- Branch from: `main`
- Merge into: `main` AND `develop`
- Delete after merge

**Create hotfix**:
```bash
git checkout main
git pull origin main
git checkout -b hotfix/1.2.1-critical-security-fix
```

#### Release Branches
```bash
# Naming: release/<version>
release/1.2.0
release/2.0.0
```

**Purpose**: Prepare for production release

**Lifecycle**:
- Branch from: `develop`
- Merge into: `main` AND `develop`
- Delete after merge

**Create release branch**:
```bash
git checkout develop
git pull origin develop
git checkout -b release/1.2.0
```

## Commit Guidelines

### Commit Message Format

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Commit Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, semicolons, etc.)
- **refactor**: Code refactoring (neither fixes a bug nor adds a feature)
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **chore**: Maintenance tasks (dependencies, build, etc.)
- **ci**: CI/CD changes

### Examples

**Feature commit**:
```
feat(auth): add JWT token validation

Implement JWT token validation middleware for API endpoints.
Tokens are verified using RS256 algorithm with public key.

Closes #123
```

**Bug fix commit**:
```
fix(database): resolve connection pool exhaustion

Previously, connections were not being properly returned to the pool
after errors. This commit ensures cleanup happens in all code paths.

Fixes #456
```

**Documentation commit**:
```
docs(readme): update installation instructions

Add instructions for macOS ARM64 installation and clarify
prerequisites for CGO support.
```

**Breaking change commit**:
```
feat(api): redesign user authentication flow

BREAKING CHANGE: The authentication API now requires OAuth 2.0
instead of basic authentication. Clients must be updated to use
the new OAuth flow.

Refs #789
```

### Commit Best Practices

1. **Atomic commits**: Each commit should represent one logical change
2. **Descriptive subject**: Clear, concise description (50 chars or less)
3. **Detailed body**: Explain what and why, not how (wrap at 72 chars)
4. **Reference issues**: Link to issue tracker (#123)
5. **Present tense**: "add feature" not "added feature"
6. **Imperative mood**: "fix bug" not "fixes bug"

### Bad vs Good Commits

**Bad**:
```
fix stuff
WIP
update
Fixed bugs and added features and updated docs
```

**Good**:
```
feat(logger): add structured logging support
fix(config): handle missing configuration file gracefully
docs(api): update authentication endpoint documentation
refactor(database): simplify connection pool management
```

## Pull Request Process

### Creating a Pull Request

1. **Ensure your branch is up to date**:
   ```bash
   git checkout develop
   git pull origin develop
   git checkout feature/123-my-feature
   git rebase develop  # or git merge develop
   ```

2. **Run all checks**:
   ```bash
   make pre-commit  # Runs fmt, vet, lint, test
   ```

3. **Push your branch**:
   ```bash
   git push origin feature/123-my-feature
   ```

4. **Create PR on GitHub** with description:
   - What changes are included
   - Why these changes are needed
   - How to test the changes
   - Link to related issues

### PR Title Format

```
<type>(<scope>): <description>
```

Examples:
```
feat(auth): implement OAuth 2.0 authentication
fix(database): resolve connection pool leak
docs(development): add debugging guide
```

### PR Description Template

```markdown
## Summary
Brief description of the changes

## Changes
- List of specific changes
- Another change
- One more change

## Testing
How to test these changes:
1. Step one
2. Step two
3. Expected outcome

## Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] All tests passing
- [ ] Linters passing
- [ ] No breaking changes (or documented)

## Related Issues
Closes #123
Refs #456
```

### PR Review Process

1. **Automated checks** run (CI/CD pipeline)
   - Code formatting
   - Linting
   - Tests
   - Code coverage
   - Security scanning

2. **Code review** by maintainers
   - At least 1 approval required
   - Address review comments
   - Update PR as needed

3. **Merge** after approval
   - Squash and merge (for feature branches)
   - Merge commit (for release branches)
   - Delete branch after merge

### Responding to Review Comments

```bash
# Make requested changes
git add .
git commit -m "refactor(auth): address review comments"
git push origin feature/123-my-feature

# Force push if you've rebased (only for feature branches)
git rebase -i develop
git push --force-with-lease origin feature/123-my-feature
```

## Release Process

### Preparing a Release

1. **Create release branch**:
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b release/1.2.0
   ```

2. **Update version numbers**:
   - Update VERSION file (if exists)
   - Update CHANGELOG.md
   - Update documentation

3. **Commit version bump**:
   ```bash
   git add .
   git commit -m "chore(release): bump version to 1.2.0"
   ```

4. **Create PR to main**:
   ```bash
   git push origin release/1.2.0
   # Create PR: release/1.2.0 -> main
   ```

5. **After PR approval, merge to main**

6. **Tag the release**:
   ```bash
   git checkout main
   git pull origin main
   git tag -a v1.2.0 -m "Release version 1.2.0"
   git push origin v1.2.0
   ```

7. **Merge back to develop**:
   ```bash
   git checkout develop
   git merge main
   git push origin develop
   ```

8. **Delete release branch**:
   ```bash
   git branch -d release/1.2.0
   git push origin --delete release/1.2.0
   ```

### Semantic Versioning

Follow [Semantic Versioning](https://semver.org/): `MAJOR.MINOR.PATCH`

- **MAJOR**: Incompatible API changes
- **MINOR**: Add functionality (backwards compatible)
- **PATCH**: Bug fixes (backwards compatible)

Examples:
- `1.0.0` - Initial stable release
- `1.1.0` - New features added
- `1.1.1` - Bug fixes
- `2.0.0` - Breaking changes

### Creating GitHub Releases

GitHub Actions automatically creates releases when tags are pushed.

Manual release:
1. Go to Releases on GitHub
2. Click "Create a new release"
3. Select tag (e.g., v1.2.0)
4. Write release notes
5. Upload artifacts (if not automated)
6. Publish release

## Common Git Tasks

### Starting a New Feature

```bash
# Update develop
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b feature/123-my-feature

# Make changes and commit
git add .
git commit -m "feat(scope): add new feature"

# Push to remote
git push -u origin feature/123-my-feature
```

### Updating Your Branch

```bash
# Option 1: Merge (preserves history)
git checkout feature/123-my-feature
git merge develop

# Option 2: Rebase (cleaner history)
git checkout feature/123-my-feature
git rebase develop

# If conflicts occur, resolve them:
git add <resolved-files>
git rebase --continue

# Force push after rebase (use with caution)
git push --force-with-lease origin feature/123-my-feature
```

### Squashing Commits

```bash
# Interactive rebase to squash last 3 commits
git rebase -i HEAD~3

# In editor, mark commits to squash:
# pick abc1234 First commit
# squash def5678 Second commit
# squash ghi9012 Third commit

# Edit commit message and save

# Force push (only for feature branches)
git push --force-with-lease origin feature/123-my-feature
```

### Undoing Changes

```bash
# Undo uncommitted changes
git checkout -- <file>
git restore <file>  # Git 2.23+

# Undo staged changes
git reset HEAD <file>
git restore --staged <file>  # Git 2.23+

# Undo last commit (keep changes)
git reset --soft HEAD~1

# Undo last commit (discard changes)
git reset --hard HEAD~1

# Revert a commit (creates new commit)
git revert <commit-hash>
```

### Stashing Changes

```bash
# Stash changes
git stash
git stash save "work in progress on feature X"

# List stashes
git stash list

# Apply stash
git stash apply
git stash apply stash@{1}

# Apply and remove stash
git stash pop

# Drop stash
git stash drop stash@{0}
```

### Cherry-Picking Commits

```bash
# Apply specific commit to current branch
git cherry-pick <commit-hash>

# Cherry-pick without committing
git cherry-pick --no-commit <commit-hash>

# Abort cherry-pick
git cherry-pick --abort
```

### Cleaning Up

```bash
# Delete local branch
git branch -d feature/123-my-feature

# Force delete unmerged branch
git branch -D feature/123-my-feature

# Delete remote branch
git push origin --delete feature/123-my-feature

# Prune deleted remote branches
git fetch --prune

# List merged branches
git branch --merged
git branch --no-merged
```

## Git Hooks

### Pre-commit Hook

Automatically run checks before commits:

Create `.git/hooks/pre-commit`:
```bash
#!/bin/sh

# Format code
make fmt

# Run linters
make lint

# Run tests
make test

# If any command fails, abort commit
if [ $? -ne 0 ]; then
    echo "Pre-commit checks failed. Commit aborted."
    exit 1
fi
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

### Pre-push Hook

Run checks before pushing:

Create `.git/hooks/pre-push`:
```bash
#!/bin/sh

# Run CI checks
make ci

if [ $? -ne 0 ]; then
    echo "Pre-push checks failed. Push aborted."
    exit 1
fi
```

### Commit Message Hook

Validate commit messages:

Create `.git/hooks/commit-msg`:
```bash
#!/bin/sh

commit_msg_file=$1
commit_msg=$(cat "$commit_msg_file")

# Check conventional commit format
if ! echo "$commit_msg" | grep -qE "^(feat|fix|docs|style|refactor|perf|test|chore|ci)(\(.+\))?: .+"; then
    echo "Error: Commit message does not follow conventional commits format"
    echo "Format: <type>(<scope>): <subject>"
    echo "Example: feat(auth): add JWT authentication"
    exit 1
fi
```

## Best Practices

### 1. Commit Early, Commit Often

- Make small, frequent commits
- Each commit should be a logical unit
- Easier to review and debug

### 2. Write Clear Commit Messages

- First line: summary (50 chars)
- Blank line
- Detailed description (72 chars per line)
- Reference issues

### 3. Keep Branches Short-Lived

- Merge feature branches within 1-2 days
- Reduces merge conflicts
- Easier code review

### 4. Sync Regularly

```bash
# Daily or before starting work
git checkout develop
git pull origin develop
```

### 5. Never Commit Secrets

```bash
# Add to .gitignore:
*.env
*.env.local
config.local.yaml
secrets.yaml
*.key
*.pem
```

### 6. Review Your Changes

```bash
# Before committing
git diff
git diff --staged

# Interactive staging
git add -p
```

### 7. Use .gitignore

The project includes a comprehensive `.gitignore`:
- Build artifacts
- Dependencies
- IDE files
- Logs and temp files
- Secret files

### 8. Protect Important Branches

Configure branch protection on GitHub:
- Require PR reviews
- Require status checks
- Prevent force push
- Require signed commits

## Git Configuration

### Recommended Settings

```bash
# User identity
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Default editor
git config --global core.editor "vim"

# Default branch name
git config --global init.defaultBranch main

# Auto-correct typos
git config --global help.autocorrect 20

# Color output
git config --global color.ui auto

# Rebase by default when pulling
git config --global pull.rebase true

# Prune deleted remote branches
git config --global fetch.prune true

# GPG signing (optional but recommended)
git config --global commit.gpgsign true
git config --global user.signingkey YOUR_GPG_KEY_ID
```

### Useful Aliases

```bash
# Common shortcuts
git config --global alias.st status
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.ci commit
git config --global alias.unstage 'reset HEAD --'
git config --global alias.last 'log -1 HEAD'
git config --global alias.visual 'log --oneline --graph --decorate --all'
git config --global alias.amend 'commit --amend --no-edit'
```

## Troubleshooting

### Resolve Merge Conflicts

```bash
# Start merge or rebase
git merge develop  # or git rebase develop

# View conflicts
git status

# Edit conflicting files (look for <<<<<<, =======, >>>>>> markers)
# Choose which changes to keep

# Mark as resolved
git add <resolved-file>

# Continue merge/rebase
git commit  # for merge
git rebase --continue  # for rebase
```

### Recover Lost Commits

```bash
# View reflog
git reflog

# Restore commit
git checkout <commit-hash>
git cherry-pick <commit-hash>
```

### Fix Wrong Commit Message

```bash
# Amend last commit message
git commit --amend -m "New message"

# If already pushed (use with caution)
git push --force-with-lease origin feature-branch
```

## Quick Reference

### Branching

```bash
git checkout -b feature/123-name  # Create branch
git checkout develop              # Switch branch
git branch -d feature-name        # Delete branch
git push origin feature-name      # Push branch
```

### Committing

```bash
git add .                         # Stage all changes
git add -p                        # Interactive staging
git commit -m "message"           # Commit
git commit --amend                # Amend last commit
```

### Syncing

```bash
git pull origin develop           # Pull changes
git push origin feature-name      # Push changes
git fetch --prune                 # Fetch and prune
```

### Undoing

```bash
git reset --soft HEAD~1           # Undo commit, keep changes
git reset --hard HEAD~1           # Undo commit, discard changes
git revert <commit>               # Revert commit
git checkout -- <file>            # Discard changes
```

### Information

```bash
git status                        # Check status
git log                           # View history
git log --oneline --graph         # Visual history
git diff                          # View changes
git blame <file>                  # Who changed what
```

## Additional Resources

- [Pro Git Book](https://git-scm.com/book/en/v2)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)
- [Git Best Practices](https://www.git-tower.com/learn/git/ebook/en/command-line/appendix/best-practices)

---

**Related**: [Code Style Guidelines](code-style.md) | [Testing Guide](testing.md) | [Development Setup](setup.md)
