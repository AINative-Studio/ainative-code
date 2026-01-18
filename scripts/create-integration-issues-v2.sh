#!/bin/bash
# Create integration-focused issues with strict TDD requirements

set -e

echo "Creating integration-focused issues with TDD requirements..."

cd /Users/aideveloper/AINative-Code

# Issue #153
echo "Creating #153: Python Backend Setup..."
sleep 2
gh issue create \
  --title "[TDD] Set Up Python Backend Microservice with FastAPI" \
  --body-file "docs/issues-v2/issue-153-python-backend-setup.md" \
  --label "P0,feature,backend,size:S,tdd" \
  --milestone "Foundation Complete"

# Issue #154
echo "Creating #154: Authentication Integration..."
sleep 2
gh issue create \
  --title "[TDD] Copy and Integrate Authentication System" \
  --body-file "docs/issues-v2/issue-154-copy-authentication.md" \
  --label "P0,feature,backend,auth,size:M,tdd" \
  --milestone "Foundation Complete"

# Issue #155
echo "Creating #155: Provider Integration..."
sleep 2
gh issue create \
  --title "[TDD] Copy and Integrate LLM Provider System" \
  --body-file "docs/issues-v2/issue-155-copy-providers.md" \
  --label "P0,feature,backend,provider,size:M,tdd" \
  --milestone "Foundation Complete"

echo "✓ Created 3 new integration-focused issues with strict TDD requirements"
echo "✓ Closed 7 issues that are no longer needed (code already exists)"
echo ""
echo "Next: Review issues at https://github.com/AINative-Studio/ainative-code/issues"
