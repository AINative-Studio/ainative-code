# quaid-scanner Report: /Users/karstenwade/Projects/AINative-Studio/src/ainative-code

**Score:** 🔴 1.8/10 — CRITICAL risk
**Maturity:** sandbox | **Depth:** standard | **Duration:** 1.1s
**Scanned:** 2026-06-01T20:55:53.522Z

## Pillar Scores

| Pillar | Score | Weight | Findings |
|--------|-------|--------|----------|
| Security | 2.0 | 25% | 0C 5W 1I |
| Governance | 0.0 | 20% | 1C 2W 9I |
| Community | 1.5 | 15% | 0C 2W 11I |
| AI Readiness | 2.5 | 15% | 0C 5W 0I |
| Inclusive Language | 0.0 | 15% | 0C 4W 36I |
| Technical Rigor | 6.5 | 10% | 0C 2W 1I |

## Critical Findings

### vendor-neutrality-critical-concentration
**Pillar:** Governance | **Category:** vendor-neutrality

Project is dominated by ainative.studio (99% of commits)

_(source: computed heuristic)_

**Suggestion:** Diversify contributors across multiple organizations to reduce single-vendor risk

**Reference:** https://chaoss.community/metric-project-sponsorship/

## Warnings

- **[TIMEOUT-binary-artifacts]** Scanner "binary-artifacts" timed out after undefinedms *(Increase scannerTimeout in configuration or check network connectivity)*
- **[TIMEOUT-dep-pinning-docker]** Scanner "dep-pinning-docker" timed out after undefinedms *(Increase scannerTimeout in configuration or check network connectivity)*
- **[TIMEOUT-openssf-local-checks]** Scanner "openssf-local-checks" timed out after undefinedms *(Increase scannerTimeout in configuration or check network connectivity)*
- **[TIMEOUT-openssf-scorecard]** Scanner "openssf-scorecard" timed out after undefinedms *(Increase scannerTimeout in configuration or check network connectivity)*
- **[TIMEOUT-token-permissions]** Scanner "token-permissions" timed out after undefinedms *(Increase scannerTimeout in configuration or check network connectivity)*
- **[governance-classification-1]** Unclear governance model — best guess is "Meritocracy" with low confidence (38%) *(Document the governance model explicitly in GOVERNANCE.md for clarity)*
- **[TIMEOUT-license-header-scanner]** Scanner "license-header-scanner" timed out after undefinedms *(Increase scannerTimeout in configuration or check network connectivity)*
- **[contributor-funnel-2]** Conversion rates: casual→regular 0%, regular→core 0% *(Low casual-to-regular conversion suggests contributor onboarding friction)*
- **[support-channels-1]** No SUPPORT.md or .github/SUPPORT.md found *(Add a SUPPORT.md documenting how users can get help)*
- **[agentic-rules-2]** CLAUDE.md lacks recognized structural sections *(Add sections like "Critical Rules", "Project Structure", "Common Tasks" to improve agent guidance.)*
- **[TIMEOUT-ai-repo-detection]** Scanner "ai-repo-detection" timed out after undefinedms *(Increase scannerTimeout in configuration or check network connectivity)*
- **[TIMEOUT-dataset-provenance]** Scanner "dataset-provenance" timed out after undefinedms *(Increase scannerTimeout in configuration or check network connectivity)*
- **[TIMEOUT-model-card-detection]** Scanner "model-card-detection" timed out after undefinedms *(Increase scannerTimeout in configuration or check network connectivity)*
- **[TIMEOUT-model-card-scoring]** Scanner "model-card-scoring" timed out after undefinedms *(Increase scannerTimeout in configuration or check network connectivity)*
- **[TIMEOUT-diminishing-language-scanner]** Scanner "diminishing-language-scanner" timed out after undefinedms *(Increase scannerTimeout in configuration or check network connectivity)*
- **[TIMEOUT-inclusive-code-scanner]** Scanner "inclusive-code-scanner" failed: Cannot read properties of undefined (reading 'termListUrl') *(Check scanner implementation for errors)*
- **[TIMEOUT-inclusive-doc-scanner]** Scanner "inclusive-doc-scanner" failed: Cannot read properties of undefined (reading 'termListUrl') *(Check scanner implementation for errors)*
- **[TIMEOUT-inclusive-naming-scanner]** Scanner "inclusive-naming-scanner" failed: Cannot read properties of undefined (reading 'termListUrl') *(Check scanner implementation for errors)*
- **[interaction-templates-1]** No issue templates configured *(Add .github/ISSUE_TEMPLATE/ with bug report and feature request templates)*
- **[semver-validation-2]** CHANGELOG missing entries for: v0.1.4, v0.1.6, v0.1.7 *(Update CHANGELOG.md to document all releases — automated tools like release-please can help)*

## Info

- **[branch-protection-1]** GitHub token not provided. Cannot check branch protection settings.
- **[asset-protection-1]** No trademark policy found (optional)
- **[asset-protection-2]** No export control documentation found (optional)
- **[asset-protection-4]** Contributor friction level: Medium
- **[bus-factor-1]** Bus factor: 1, Elephant factor: 99% (2 contributors, 157 commits in last 12 months)
- **[dep-license-scanning-1]** Go dependencies detected (go.mod) — license scanning requires downloaded modules
- **[governance-detection-1]** No governance documentation found
- **[license-compatibility-1]** Project license is MIT — no installed dependencies to check compatibility
- **[vendor-neutrality-domain-count]** Found 2 unique email domain(s) across 157 commits
- **[vendor-neutrality-no-succession]** No succession planning documentation found
- **[burnout-detection-1]** Burnout detection requires a GitHub token
- **[contributor-data-1]** 2 unique contributors with 157 commits in the last 12 months
- **[contributor-data-2]** Contributor emails span 2 domains
- **[contributor-funnel-1]** Contributor funnel: 1 core, 0 regular, 1 casual (2 total)
- **[funding-1]** No funding infrastructure detected
- **[issue-closure-1]** Issue closure analysis requires a GitHub token
- **[response-classification-1]** Response classification requires a GitHub token
- **[response-time-1]** Response time analysis requires a GitHub token
- **[stale-bot-1]** No stale bot configured
- **[support-channels-2]** README contains a support/help section
- **[support-channels-3]** Support channels detected: discussions
- **[AK-GIT-CLONE-README.md:335]** Assumed knowledge: "clone" operation used without explanation
- **[AK-GIT-CLONE-README.md:336]** Assumed knowledge: "clone" operation used without explanation
- **[AK-GIT-FORK-README.md:400]** Assumed knowledge: "fork" operation used without explanation
- **[AK-GIT-BRANCH-README.md:401]** Assumed knowledge: "branch" operation used without explanation
- **[AK-ACRONYM-LICENSE-README.md:7]** Undefined acronym "LICENSE" may confuse newcomers
- **[AK-ACRONYM-GPT-README.md:37]** Undefined acronym "GPT" may confuse newcomers
- **[AK-ACRONYM-TUI-README.md:38]** Undefined acronym "TUI" may confuse newcomers
- **[AK-ACRONYM-CMS-README.md:39]** Undefined acronym "CMS" may confuse newcomers
- **[AK-ACRONYM-RLHF-README.md:39]** Undefined acronym "RLHF" may confuse newcomers
- **[AK-ACRONYM-JWT-README.md:40]** Undefined acronym "JWT" may confuse newcomers
- **[AK-ACRONYM-LLM-README.md:40]** Undefined acronym "LLM" may confuse newcomers
- **[AK-ACRONYM-LOCALAPPDATA-README.md:74]** Undefined acronym "LOCALAPPDATA" may confuse newcomers
- **[AK-ACRONYM-PATH-README.md:75]** Undefined acronym "PATH" may confuse newcomers
- **[AK-ACRONYM-APT-README.md:95]** Undefined acronym "APT" may confuse newcomers
- **[AK-ACRONYM-YUM-README.md:95]** Undefined acronym "YUM" may confuse newcomers
- **[AK-ACRONYM-SELECT-README.md:268]** Undefined acronym "SELECT" may confuse newcomers
- **[AK-ACRONYM-FROM-README.md:268]** Undefined acronym "FROM" may confuse newcomers
- **[AK-ACRONYM-WHERE-README.md:268]** Undefined acronym "WHERE" may confuse newcomers
- **[AK-ACRONYM-TASK-README.md:279]** Undefined acronym "TASK" may confuse newcomers
- **[AK-ACRONYM-DEBUG-README.md:284]** Undefined acronym "DEBUG" may confuse newcomers
- **[AK-ACRONYM-INFO-README.md:284]** Undefined acronym "INFO" may confuse newcomers
- **[AK-ACRONYM-WARN-README.md:284]** Undefined acronym "WARN" may confuse newcomers
- **[AK-ACRONYM-ERROR-README.md:284]** Undefined acronym "ERROR" may confuse newcomers
- **[AK-ACRONYM-FATAL-README.md:284]** Undefined acronym "FATAL" may confuse newcomers
- **[AK-ACRONYM-CONTRIBUTING-README.md:392]** Undefined acronym "CONTRIBUTING" may confuse newcomers
- **[AK-ACRONYM-README-README.md:397]** Undefined acronym "README" may confuse newcomers
- **[AK-GIT-FORK-CONTRIBUTING.md:33]** Assumed knowledge: "fork" operation used without explanation
- **[AK-GIT-CLONE-CONTRIBUTING.md:34]** Assumed knowledge: "clone" operation used without explanation
- **[AK-GIT-CLONE-CONTRIBUTING.md:36]** Assumed knowledge: "clone" operation used without explanation
- **[AK-GIT-BRANCH-CONTRIBUTING.md:86]** Assumed knowledge: "branch" operation used without explanation
- **[AK-GIT-REBASE-CONTRIBUTING.md:338]** Assumed knowledge: "rebase" operation used without explanation
- **[AK-ACRONYM-LLM-CONTRIBUTING.md:158]** Undefined acronym "LLM" may confuse newcomers
- **[AK-ACRONYM-INFO-CONTRIBUTING.md:260]** Undefined acronym "INFO" may confuse newcomers
- **[AK-ACRONYM-JWT-CONTRIBUTING.md:297]** Undefined acronym "JWT" may confuse newcomers
- **[AK-ACRONYM-CHANGELOG-CONTRIBUTING.md:337]** Undefined acronym "CHANGELOG" may confuse newcomers
- **[AK-ACRONYM-README-CONTRIBUTING.md:417]** Undefined acronym "README" may confuse newcomers
- **[linter-config-2]** Linter config found but no lint step detected in CI workflows

## Recommendations

- **[HIGH impact / medium effort]** Diversify contributors across multiple organizations to reduce single-vendor risk
  - https://chaoss.community/metric-project-sponsorship/
- **[MEDIUM impact / low effort]** Increase scannerTimeout in configuration or check network connectivity
- **[MEDIUM impact / low effort]** Document the governance model explicitly in GOVERNANCE.md for clarity
- **[MEDIUM impact / low effort]** Increase scannerTimeout in configuration or check network connectivity
- **[MEDIUM impact / low effort]** Low casual-to-regular conversion suggests contributor onboarding friction
- **[MEDIUM impact / low effort]** Add a SUPPORT.md documenting how users can get help
- **[MEDIUM impact / low effort]** Add sections like "Critical Rules", "Project Structure", "Common Tasks" to improve agent guidance.
- **[MEDIUM impact / low effort]** Increase scannerTimeout in configuration or check network connectivity
- **[MEDIUM impact / low effort]** Increase scannerTimeout in configuration or check network connectivity
- **[MEDIUM impact / low effort]** Check scanner implementation for errors
- **[MEDIUM impact / low effort]** Add .github/ISSUE_TEMPLATE/ with bug report and feature request templates
- **[MEDIUM impact / low effort]** Update CHANGELOG.md to document all releases — automated tools like release-please can help

## Score Rationale

Overall score is a weighted sum of six pillar scores (each scored 0–10).

| Pillar | Weight | Raw Score | Contribution |
|--------|--------|-----------|-------------|
| Security | 25% | 2.0 | 0.50 |
| Governance | 20% | 0.0 | 0.00 |
| Community | 15% | 1.5 | 0.22 |
| AI Readiness | 15% | 2.5 | 0.38 |
| Inclusive Language | 15% | 0.0 | 0.00 |
| Technical Rigor | 10% | 6.5 | 0.65 |
| **Overall** | **100%** | | **1.80** |

---
*quaid-scanner v0.1.2 | 2026-06-01T20:55:53.522Z*
*Commit: e114af87a048fe8d0d0d47c67eed043f17d5555d*