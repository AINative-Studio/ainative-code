---
name: weekly-report
description: Generate comprehensive weekly progress reports for AINative platform. Use when (1) Creating end-of-week status reports, (2) Summarizing commits and features, (3) Documenting bug fixes and improvements, (4) Tracking sprint progress, (5) Preparing stakeholder updates. Analyzes git commits, GitHub issues, and PRs to produce structured markdown reports.
---

# Weekly Report Generation

Generate comprehensive weekly progress reports that summarize development activity across all AINative repositories.

## Report Structure

### Required Sections

1. **Executive Summary** - High-level overview of the week
2. **Major Features Implemented** - New functionality with commit refs
3. **Critical Bug Fixes** - Issues resolved with root cause analysis
4. **Security Improvements** - Vulnerability fixes, hardening
5. **Infrastructure & DevOps** - Deployment, configuration changes
6. **Frontend Improvements** - UI/UX changes across web apps
7. **Work In Progress** - Ongoing work, backlog updates
8. **Commit Statistics** - Quantitative analysis
9. **Success Metrics** - KPIs and feature completion tracking
10. **Next Week Priorities** - Upcoming focus areas

## Data Collection Process

### Step 1: Gather Git Commits

```bash
# Get commits from the past week for core repo
cd /Users/ranveerdeshmukh/AINative-core/core
git log --since="7 days ago" --pretty=format:"%h %s" --no-merges | head -100

# Count commits by date
git log --since="7 days ago" --format="%ad" --date=short | sort | uniq -c

# Get commit authors
git log --since="7 days ago" --format="%an" | sort | uniq -c | sort -rn
```

### Step 2: Gather GitHub Issues

```bash
# List closed issues this week
gh issue list --repo AINative-Studio/core --state closed --limit 100

# List created issues this week
gh issue list --repo AINative-Studio/core --state all --limit 100

# Get issue details
gh issue view <issue-number> --repo AINative-Studio/core
```

### Step 3: Analyze PRs

```bash
# List merged PRs
gh pr list --repo AINative-Studio/core --state merged --limit 50

# Get PR details
gh pr view <pr-number> --repo AINative-Studio/core
```

## Report Template

```markdown
# AINative Platform - Weekly Progress Report
## [Start Date] - [End Date]

---

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [Major Features Implemented](#major-features-implemented)
3. [Critical Bug Fixes](#critical-bug-fixes)
4. [Security Improvements](#security-improvements)
5. [Infrastructure & DevOps](#infrastructure--devops)
6. [Frontend Improvements](#frontend-improvements)
7. [Work In Progress](#work-in-progress)
8. [Commit Statistics](#commit-statistics)
9. [Success Metrics](#success-metrics)
10. [Next Week Priorities](#next-week-priorities)

---

## Executive Summary

This reporting period saw **[X] commits** across [N] repositories. Major accomplishments include:

- **[Feature 1]**: Brief description
- **[Feature 2]**: Brief description
- **[Bug Fix]**: Brief description
- **[Security]**: Brief description

**Status**: [Overall assessment]

---

## Major Features Implemented

### 1. [Feature Name]
**Commits**: [commit_hash1, commit_hash2]
**Status**: Complete/In Progress
**Story Points**: [X]

#### Problem Solved:
[Description of what problem this feature addresses]

#### Implementation Details:

**Backend** (`src/backend/app/services/`):
- `service_file.py` - Description of service

**API Endpoints** (`src/backend/app/api/v1/endpoints/`):
- `endpoint.py` - Description of endpoint

**Impact**: HIGH/MEDIUM/LOW - [Why this matters]

---

## Critical Bug Fixes

### 1. [Bug Title] (CRITICAL/HIGH/MEDIUM/LOW)
**Commit**: [commit_hash]
**Issue #[number]**: [Brief description]

**Root Causes**:
1. [Cause 1]
2. [Cause 2]

**Fixes**:
- [What was fixed]
- [How it was fixed]

**Impact**: [Effect of the fix]

---

## Security Improvements

### 1. [Security Issue Title]
**Commits**: [commit_hashes]
**Issues**: #[numbers]

#### Vulnerabilities Fixed:
- **[Package]**: [Vulnerability type] (SEVERITY)

**Impact**: [Security improvement]

---

## Infrastructure & DevOps

### 1. [Infrastructure Change]
**Commits**: [commit_hashes]

**Changes**:
- [Change 1]
- [Change 2]

---

## Frontend Improvements

### [App Name] ([X] commits)

#### 1. [Improvement Title]
**Commits**: [commit_hashes]

**Changes**:
- [Change 1]
- [Change 2]

---

## Work In Progress

### [Feature/Project Name]
**Issues**: #[range]
**Status**: [Current status]

[Description of ongoing work]

---

## Commit Statistics

**Total Commits**: [X]
**Period**: [Start Date] - [End Date] ([N] days)
**Daily Average**: [X] commits/day

**By Repository**:
| Repository | Commits | Focus Area |
|------------|---------|------------|
| AINative-Studio/core | [X] | [Focus] |
| relycapital/AINative-website | [X] | [Focus] |

**Categories**:
| Type | Count | Percentage |
|------|-------|------------|
| Features | [X] | [X]% |
| Bug Fixes | [X] | [X]% |
| Security | [X] | [X]% |
| Tests | [X] | [X]% |
| DevOps | [X] | [X]% |
| Docs | [X] | [X]% |

**Commits by Date**:
| Date | Commits | Key Changes |
|------|---------|-------------|
| [Date] | [X] | [Summary] |

---

## Success Metrics

### Technical Metrics

| Metric | Target | Status |
|--------|--------|--------|
| [Metric 1] | [Target] | Achieved/Pending |

### Feature Completion

| Feature | Status | Impact |
|---------|--------|--------|
| [Feature] | Complete/In Progress | CRITICAL/HIGH/MEDIUM/LOW |

---

## Next Week Priorities

1. **[Priority 1]**: [Description]
2. **[Priority 2]**: [Description]
3. **[Priority 3]**: [Description]

---

## Document Version

**Version**: 1.0
**Date**: [Date]
**Author**: [Username]
**Status**: Ready for Review

---

*End of Weekly Progress Report*
```

## File Naming Convention

```
docs/reports/WEEKLY_REPORT_YYYY-MM-DD_username.md
```

Example: `WEEKLY_REPORT_2026-01-04_ranveerd11.md`

## Quality Checklist

Before finalizing the report:

- [ ] All commits from the week are analyzed
- [ ] GitHub issues are linked with # notation
- [ ] Commit hashes are included for traceability
- [ ] Impact levels (CRITICAL/HIGH/MEDIUM/LOW) assigned
- [ ] Statistics are accurate
- [ ] Next week priorities are actionable
- [ ] No sensitive data (credentials, PII) included
- [ ] File placed in `docs/reports/` directory

## Repositories to Include

1. **AINative-Studio/core** - Main backend and API
2. **relycapital/AINative-website** - Marketing website
3. **urbantech/live.ainative.studio** - Live streaming platform
4. **AINative-Studio/chatwoot-zerodb** - Customer support integration

## Categorizing Commits

| Prefix/Keywords | Category |
|-----------------|----------|
| feat, add, implement | Features |
| fix, resolve, correct | Bug Fixes |
| security, CVE, vulnerability | Security |
| test, spec, coverage | Tests |
| deploy, CI, CD, config | DevOps |
| doc, readme, comment | Docs |
| refactor, clean, optimize | Refactor |

## Impact Assessment

- **CRITICAL**: Core functionality, data integrity, security vulnerabilities
- **HIGH**: Major features, significant bug fixes, performance improvements
- **MEDIUM**: Enhancements, moderate fixes, integrations
- **LOW**: Minor fixes, UI polish, documentation

## Example Usage

```bash
# Invoke the weekly report skill
/weekly-report

# Generate report for specific date range
/weekly-report --start 2026-01-01 --end 2026-01-07

# Generate report for specific author
/weekly-report --author ranveerd11
```
