#!/bin/bash
# Verification script for Issue #154: AINative API Integration

echo "=================================================="
echo "Issue #154 Verification Script"
echo "AINative API Authentication Integration"
echo "=================================================="
echo ""

# Navigate to project directory
cd "$(dirname "$0")"

echo "1. Running test suite..."
echo "---------------------------------------------------"
./venv/bin/python -m pytest tests/api/test_ainative_client.py -v --tb=short 2>&1 | grep -E "(test_|PASSED|FAILED|passed|failed)"
echo ""

echo "2. Checking code coverage..."
echo "---------------------------------------------------"
./venv/bin/python -m pytest tests/api/test_ainative_client.py \
  --cov=app.api.ainative_client \
  --cov-report=term \
  --cov-fail-under=80 2>&1 | grep -E "(Stmts|Miss|Cover|TOTAL|Required test coverage)"
echo ""

echo "3. Verifying FastAPI endpoints..."
echo "---------------------------------------------------"
./venv/bin/python -c "
from app.main import app
routes = [r.path for r in app.routes if hasattr(r, 'path')]
auth_routes = [r for r in routes if 'auth' in r]
chat_routes = [r for r in routes if 'chat' in r]
print(f'Total routes: {len(routes)}')
print(f'Auth routes ({len(auth_routes)}): {auth_routes}')
print(f'Chat routes ({len(chat_routes)}): {chat_routes}')
"
echo ""

echo "4. Checking file structure..."
echo "---------------------------------------------------"
echo "New files created:"
[ -f "tests/api/test_ainative_client.py" ] && echo "  ✓ tests/api/test_ainative_client.py" || echo "  ✗ tests/api/test_ainative_client.py"
[ -f "app/api/ainative_client.py" ] && echo "  ✓ app/api/ainative_client.py" || echo "  ✗ app/api/ainative_client.py"
[ -f "app/schemas/auth.py" ] && echo "  ✓ app/schemas/auth.py" || echo "  ✗ app/schemas/auth.py"
[ -f "app/api/v1/endpoints/auth.py" ] && echo "  ✓ app/api/v1/endpoints/auth.py" || echo "  ✗ app/api/v1/endpoints/auth.py"
echo ""

echo "5. Final Status"
echo "---------------------------------------------------"
./venv/bin/python -c "
import sys
sys.path.insert(0, '.')
from app.api.ainative_client import AINativeClient

# Verify class exists
client = AINativeClient()
print(f'✓ AINativeClient initialized successfully')
print(f'✓ Base URL: {client.base_url}')
print(f'✓ Timeout: {client.timeout}s')

# Verify all methods exist
methods = ['login', 'register', 'get_current_user', 'refresh_token', 'logout', 'chat_completion']
for method in methods:
    if hasattr(client, method):
        print(f'✓ Method {method}() exists')
    else:
        print(f'✗ Method {method}() missing')
"
echo ""

echo "=================================================="
echo "Verification Complete!"
echo "=================================================="
