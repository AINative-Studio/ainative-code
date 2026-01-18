# Issue #154: Copy and Integrate Authentication System from AINative Platform

## Priority
P0 - CRITICAL

## Description
Copy production-ready authentication code from AINative platform (`/Users/aideveloper/core/src/backend/app/core/`) and integrate into the Python backend microservice. Implement JWT-based authentication with token refresh following TDD principles.

## TDD Workflow (MANDATORY)
**⚠️ CRITICAL: Tests MUST be written FIRST before copying/adapting code**

### Test-First Development Steps:
1. Write failing tests for JWT token creation
2. Write failing tests for JWT token validation
3. Write failing tests for password hashing
4. Write failing tests for token refresh
5. Copy and adapt code to make tests pass
6. Achieve 80%+ coverage

## Acceptance Criteria

### Authentication Core (TDD - Tests First!)
- [ ] **FIRST**: Write test for password hashing (test_password_hash.py)
- [ ] **THEN**: Copy `security_enhanced.py` and adapt
- [ ] **FIRST**: Write test for JWT token creation
- [ ] **THEN**: Adapt token creation logic
- [ ] **FIRST**: Write test for JWT token validation
- [ ] **THEN**: Adapt token validation logic
- [ ] **FIRST**: Write test for token refresh
- [ ] **THEN**: Implement token refresh endpoint

### Authentication Endpoints (TDD - Tests First!)
- [ ] **FIRST**: Write test for `/auth/login` endpoint
- [ ] **THEN**: Implement login endpoint
- [ ] **FIRST**: Write test for `/auth/refresh` endpoint
- [ ] **THEN**: Implement refresh endpoint
- [ ] **FIRST**: Write test for `/auth/logout` endpoint
- [ ] **THEN**: Implement logout endpoint
- [ ] **FIRST**: Write test for authentication dependency
- [ ] **THEN**: Implement `get_current_user` dependency

### Security Requirements
- [ ] Passwords hashed with bcrypt (cost factor 12)
- [ ] JWT tokens signed with RS256 or HS256
- [ ] Access tokens expire after 8 hours
- [ ] Refresh tokens expire after 30 days
- [ ] Secret key loaded from environment (never hardcoded)
- [ ] HTTPS enforced in production

### Testing Requirements (80%+ Coverage MANDATORY)
- [ ] Unit tests for password hashing (bcrypt)
- [ ] Unit tests for JWT creation and validation
- [ ] Integration tests for login flow
- [ ] Integration tests for token refresh
- [ ] Integration tests for protected endpoints
- [ ] Negative tests for invalid tokens
- [ ] Negative tests for expired tokens
- [ ] Coverage >= 80% measured with pytest-cov

## Technical Requirements

### Files to Copy from AINative Platform

**Source Files:**
```
/Users/aideveloper/core/src/backend/app/core/auth.py
/Users/aideveloper/core/src/backend/app/core/security_enhanced.py
/Users/aideveloper/core/src/backend/app/schemas/auth.py
/Users/aideveloper/core/src/backend/app/api/v1/endpoints/auth.py
```

**Destination Structure:**
```
python-backend/
├── app/
│   ├── core/
│   │   ├── auth.py           ← Copy from AINative platform
│   │   └── security.py       ← Copy from security_enhanced.py
│   ├── schemas/
│   │   └── auth.py           ← Copy from AINative platform
│   └── api/
│       └── v1/
│           └── endpoints/
│               └── auth.py   ← Copy from AINative platform
└── tests/
    ├── core/
    │   ├── test_security.py  ← Write FIRST!
    │   └── test_auth.py      ← Write FIRST!
    └── api/
        └── test_auth_endpoints.py  ← Write FIRST!
```

### TDD Test Examples (Write FIRST!)

**Test 1: Password Hashing** (tests/core/test_security.py)
```python
import pytest
from app.core.security import hash_password, verify_password

def test_hash_password_creates_bcrypt_hash():
    """GIVEN a plain text password
    WHEN hash_password is called
    THEN it should return a bcrypt hash"""
    password = "SecurePassword123!"
    hashed = hash_password(password)

    # Bcrypt hashes start with $2b$
    assert hashed.startswith("$2b$")
    # Should not equal the original password
    assert hashed != password

def test_verify_password_with_correct_password():
    """GIVEN a password and its hash
    WHEN verify_password is called with correct password
    THEN it should return True"""
    password = "SecurePassword123!"
    hashed = hash_password(password)

    assert verify_password(password, hashed) is True

def test_verify_password_with_incorrect_password():
    """GIVEN a password hash
    WHEN verify_password is called with wrong password
    THEN it should return False"""
    password = "SecurePassword123!"
    wrong_password = "WrongPassword456!"
    hashed = hash_password(password)

    assert verify_password(wrong_password, hashed) is False

def test_hash_password_produces_different_hashes():
    """GIVEN the same password
    WHEN hash_password is called twice
    THEN it should produce different hashes (salt)"""
    password = "SecurePassword123!"
    hash1 = hash_password(password)
    hash2 = hash_password(password)

    assert hash1 != hash2
    # But both should verify correctly
    assert verify_password(password, hash1) is True
    assert verify_password(password, hash2) is True
```

**Test 2: JWT Token Creation** (tests/core/test_auth.py)
```python
import pytest
from datetime import datetime, timedelta
from jose import jwt
from app.core.auth import create_access_token, decode_token
from app.core.config import settings

def test_create_access_token_includes_subject():
    """GIVEN user data
    WHEN create_access_token is called
    THEN token should contain subject"""
    data = {"sub": "user@example.com"}
    token = create_access_token(data)

    decoded = jwt.decode(
        token,
        settings.SECRET_KEY,
        algorithms=[settings.ALGORITHM]
    )
    assert decoded["sub"] == "user@example.com"

def test_create_access_token_includes_expiration():
    """GIVEN user data
    WHEN create_access_token is called
    THEN token should have expiration time"""
    data = {"sub": "user@example.com"}
    token = create_access_token(data)

    decoded = jwt.decode(
        token,
        settings.SECRET_KEY,
        algorithms=[settings.ALGORITHM]
    )
    assert "exp" in decoded

    # Expiration should be in the future
    exp = datetime.fromtimestamp(decoded["exp"])
    assert exp > datetime.utcnow()

def test_create_access_token_with_custom_expiration():
    """GIVEN custom expiration delta
    WHEN create_access_token is called
    THEN token should expire after specified time"""
    data = {"sub": "user@example.com"}
    expires_delta = timedelta(minutes=15)
    token = create_access_token(data, expires_delta=expires_delta)

    decoded = jwt.decode(
        token,
        settings.SECRET_KEY,
        algorithms=[settings.ALGORITHM]
    )

    exp = datetime.fromtimestamp(decoded["exp"])
    expected_exp = datetime.utcnow() + expires_delta

    # Allow 5 second tolerance
    assert abs((exp - expected_exp).total_seconds()) < 5

def test_decode_token_with_valid_token():
    """GIVEN a valid JWT token
    WHEN decode_token is called
    THEN it should return the payload"""
    data = {"sub": "user@example.com", "role": "admin"}
    token = create_access_token(data)

    payload = decode_token(token)
    assert payload["sub"] == "user@example.com"
    assert payload["role"] == "admin"

def test_decode_token_with_expired_token():
    """GIVEN an expired JWT token
    WHEN decode_token is called
    THEN it should raise an exception"""
    data = {"sub": "user@example.com"}
    expires_delta = timedelta(seconds=-10)  # Expired 10 seconds ago
    token = create_access_token(data, expires_delta=expires_delta)

    with pytest.raises(jwt.ExpiredSignatureError):
        decode_token(token)

def test_decode_token_with_invalid_signature():
    """GIVEN a token with invalid signature
    WHEN decode_token is called
    THEN it should raise an exception"""
    data = {"sub": "user@example.com"}
    # Create token with different key
    token = jwt.encode(data, "wrong-secret-key", algorithm="HS256")

    with pytest.raises(jwt.JWTError):
        decode_token(token)
```

**Test 3: Login Endpoint** (tests/api/test_auth_endpoints.py)
```python
import pytest
from fastapi.testclient import TestClient
from app.main import app
from app.core.security import hash_password

client = TestClient(app)

@pytest.fixture
def test_user(db_session):
    """Create a test user in the database"""
    from app.models.user import User

    user = User(
        email="test@example.com",
        hashed_password=hash_password("TestPassword123!"),
        is_active=True
    )
    db_session.add(user)
    db_session.commit()
    return user

def test_login_with_valid_credentials(test_user):
    """GIVEN a user with valid credentials
    WHEN POST /auth/login is called
    THEN it should return access and refresh tokens"""
    response = client.post(
        "/api/v1/auth/login",
        json={
            "email": "test@example.com",
            "password": "TestPassword123!"
        }
    )

    assert response.status_code == 200
    data = response.json()
    assert "access_token" in data
    assert "refresh_token" in data
    assert "token_type" in data
    assert data["token_type"] == "bearer"

def test_login_with_invalid_password(test_user):
    """GIVEN incorrect password
    WHEN POST /auth/login is called
    THEN it should return 401 Unauthorized"""
    response = client.post(
        "/api/v1/auth/login",
        json={
            "email": "test@example.com",
            "password": "WrongPassword"
        }
    )

    assert response.status_code == 401
    assert "Incorrect email or password" in response.json()["detail"]

def test_login_with_nonexistent_user():
    """GIVEN nonexistent user email
    WHEN POST /auth/login is called
    THEN it should return 401 Unauthorized"""
    response = client.post(
        "/api/v1/auth/login",
        json={
            "email": "nonexistent@example.com",
            "password": "SomePassword"
        }
    )

    assert response.status_code == 401

def test_login_with_inactive_user(test_user, db_session):
    """GIVEN an inactive user account
    WHEN POST /auth/login is called
    THEN it should return 401 Unauthorized"""
    test_user.is_active = False
    db_session.commit()

    response = client.post(
        "/api/v1/auth/login",
        json={
            "email": "test@example.com",
            "password": "TestPassword123!"
        }
    )

    assert response.status_code == 401
    assert "Inactive account" in response.json()["detail"]

def test_protected_endpoint_with_valid_token(test_user):
    """GIVEN a valid access token
    WHEN accessing a protected endpoint
    THEN it should allow access"""
    # Login first
    login_response = client.post(
        "/api/v1/auth/login",
        json={
            "email": "test@example.com",
            "password": "TestPassword123!"
        }
    )
    token = login_response.json()["access_token"]

    # Access protected endpoint
    response = client.get(
        "/api/v1/users/me",
        headers={"Authorization": f"Bearer {token}"}
    )

    assert response.status_code == 200
    data = response.json()
    assert data["email"] == "test@example.com"

def test_protected_endpoint_without_token():
    """GIVEN no authentication token
    WHEN accessing a protected endpoint
    THEN it should return 401 Unauthorized"""
    response = client.get("/api/v1/users/me")

    assert response.status_code == 401

def test_protected_endpoint_with_invalid_token():
    """GIVEN an invalid token
    WHEN accessing a protected endpoint
    THEN it should return 401 Unauthorized"""
    response = client.get(
        "/api/v1/users/me",
        headers={"Authorization": "Bearer invalid_token_here"}
    )

    assert response.status_code == 401
```

**Test 4: Token Refresh** (tests/api/test_token_refresh.py)
```python
import pytest
from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)

def test_refresh_token_with_valid_refresh_token(test_user):
    """GIVEN a valid refresh token
    WHEN POST /auth/refresh is called
    THEN it should return new access token"""
    # Login to get refresh token
    login_response = client.post(
        "/api/v1/auth/login",
        json={
            "email": "test@example.com",
            "password": "TestPassword123!"
        }
    )
    refresh_token = login_response.json()["refresh_token"]

    # Refresh the token
    response = client.post(
        "/api/v1/auth/refresh",
        json={"refresh_token": refresh_token}
    )

    assert response.status_code == 200
    data = response.json()
    assert "access_token" in data
    assert "token_type" in data

def test_refresh_token_with_invalid_token():
    """GIVEN an invalid refresh token
    WHEN POST /auth/refresh is called
    THEN it should return 401 Unauthorized"""
    response = client.post(
        "/api/v1/auth/refresh",
        json={"refresh_token": "invalid_token"}
    )

    assert response.status_code == 401
```

### Copy and Adaptation Steps

**Step 1: Copy Files**
```bash
# Create directory structure
mkdir -p python-backend/app/core
mkdir -p python-backend/app/schemas
mkdir -p python-backend/app/api/v1/endpoints

# Copy authentication files
cp /Users/aideveloper/core/src/backend/app/core/auth.py \
   python-backend/app/core/auth.py

cp /Users/aideveloper/core/src/backend/app/core/security_enhanced.py \
   python-backend/app/core/security.py

cp /Users/aideveloper/core/src/backend/app/schemas/auth.py \
   python-backend/app/schemas/auth.py

cp /Users/aideveloper/core/src/backend/app/api/v1/endpoints/auth.py \
   python-backend/app/api/v1/endpoints/auth.py
```

**Step 2: Update Import Paths**
```python
# In all copied files, update:
from app.core.database import get_db
# TO:
from app.core.database import get_db

from app.models.user import User
# TO:
from app.models.user import User

from app.core.security import ...
# TO:
from app.core.security import ...
```

**Step 3: Run Tests and Fix**
```bash
# Run tests (should fail at first)
pytest tests/core/test_security.py -v
pytest tests/core/test_auth.py -v
pytest tests/api/test_auth_endpoints.py -v

# Fix code until all tests pass
# Check coverage
pytest --cov=app.core --cov=app.api.v1.endpoints.auth \
       --cov-report=term-missing

# Coverage should be >= 80%
```

## Dependencies
- Issue #153: Python backend setup must be complete

## Definition of Done
- [ ] All tests written FIRST and passing
- [ ] Code coverage >= 80% for auth module
- [ ] Authentication files copied and adapted
- [ ] Import paths updated correctly
- [ ] Login endpoint working (POST /api/v1/auth/login)
- [ ] Token refresh endpoint working (POST /api/v1/auth/refresh)
- [ ] Protected endpoints require valid JWT
- [ ] Invalid/expired tokens rejected
- [ ] Environment variable SECRET_KEY configured
- [ ] Code follows project coding standards
- [ ] PR approved and merged

## Estimated Effort
**Size:** Medium (1-2 days)

**Breakdown:**
- TDD Test Writing: 4 hours
- Copy and adapt files: 2 hours
- Fix import paths: 1 hour
- Integration and debugging: 2 hours
- Documentation: 1 hour

## Labels
- `P0` - Critical priority
- `feature` - New feature
- `backend` - Backend component
- `auth` - Authentication
- `size:M` - Medium effort
- `tdd` - Test-driven development

## Milestone
Foundation Complete (Week 1)

## Notes
- **TDD is MANDATORY**: Write tests before copying/adapting code
- **80% coverage is NON-NEGOTIABLE**: Measure on every commit
- Use existing AINative platform code as reference
- DO NOT modify copied code until tests are written
- Follow BDD-style test naming
- Test both success and failure cases
- Bcrypt cost factor: 12 (production standard)
