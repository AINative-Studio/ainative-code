# Issue #153: Set Up Python Backend Microservice with FastAPI

## Priority
P0 - CRITICAL

## Description
Create a new Python FastAPI microservice that will serve as the authentication and chat completion backend for ainative-code. This service will reuse production-ready code from the AINative platform.

## TDD Workflow (MANDATORY)
**⚠️ CRITICAL: Tests MUST be written FIRST before implementation code**

### Test-First Development Steps:
1. Write failing tests for each feature
2. Run tests and verify they fail (Red)
3. Implement minimal code to make tests pass (Green)
4. Refactor while keeping tests green (Refactor)
5. Achieve 80%+ coverage before moving to next feature

## Acceptance Criteria

### Service Structure
- [ ] FastAPI project created with proper directory structure
- [ ] Virtual environment configured with Python 3.11+
- [ ] Dependencies installed (FastAPI, uvicorn, pytest, httpx)
- [ ] Health check endpoint `/health` returns 200 OK
- [ ] API versioning implemented (`/api/v1/`)
- [ ] CORS configuration for development

### Testing Requirements (TDD - Write Tests FIRST)
- [ ] **FIRST**: Write test for health endpoint (test_health_endpoint.py)
- [ ] **THEN**: Implement health endpoint
- [ ] **FIRST**: Write test for CORS configuration
- [ ] **THEN**: Configure CORS middleware
- [ ] **FIRST**: Write test for API versioning
- [ ] **THEN**: Implement API router structure
- [ ] Coverage: 80%+ (MANDATORY - measured with pytest-cov)
- [ ] All tests passing before PR submission

### Configuration
- [ ] Environment variable management with Pydantic Settings
- [ ] `.env.example` file with all required variables
- [ ] Configuration validation on startup
- [ ] Separate configs for dev/staging/prod

### Documentation
- [ ] README.md with setup instructions
- [ ] API documentation auto-generated (FastAPI Swagger)
- [ ] Environment variable documentation
- [ ] Development workflow guide

## Technical Requirements

### Project Structure
```
python-backend/
├── app/
│   ├── __init__.py
│   ├── main.py              # FastAPI application
│   ├── core/
│   │   ├── __init__.py
│   │   ├── config.py        # Pydantic Settings
│   │   └── logging.py       # Logging configuration
│   ├── api/
│   │   ├── __init__.py
│   │   └── v1/
│   │       ├── __init__.py
│   │       └── router.py    # API router
│   └── schemas/
│       └── __init__.py
├── tests/
│   ├── __init__.py
│   ├── conftest.py          # pytest fixtures
│   └── test_health.py       # Health endpoint tests
├── .env.example
├── .gitignore
├── pytest.ini
├── pyproject.toml           # or requirements.txt
└── README.md
```

### Dependencies (pyproject.toml or requirements.txt)
```toml
[tool.poetry.dependencies]
python = "^3.11"
fastapi = "^0.109.0"
uvicorn = {extras = ["standard"], version = "^0.27.0"}
pydantic = "^2.5.0"
pydantic-settings = "^2.1.0"
python-dotenv = "^1.0.0"

[tool.poetry.dev-dependencies]
pytest = "^7.4.0"
pytest-cov = "^4.1.0"
pytest-asyncio = "^0.21.0"
httpx = "^0.26.0"
black = "^24.0.0"
ruff = "^0.1.0"
mypy = "^1.8.0"
```

### TDD Test Examples (Write FIRST!)

**Test 1: Health Endpoint** (tests/test_health.py)
```python
import pytest
from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)

def test_health_endpoint_returns_200():
    """GIVEN a FastAPI application
    WHEN the /health endpoint is called
    THEN it should return 200 OK"""
    response = client.get("/health")
    assert response.status_code == 200

def test_health_endpoint_returns_json():
    """GIVEN a FastAPI application
    WHEN the /health endpoint is called
    THEN it should return JSON with status"""
    response = client.get("/health")
    data = response.json()
    assert "status" in data
    assert data["status"] == "healthy"

def test_health_endpoint_includes_version():
    """GIVEN a FastAPI application
    WHEN the /health endpoint is called
    THEN it should include API version"""
    response = client.get("/health")
    data = response.json()
    assert "version" in data
```

**Test 2: API Versioning** (tests/test_api_versioning.py)
```python
import pytest
from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)

def test_api_v1_prefix_exists():
    """GIVEN a FastAPI application
    WHEN accessing /api/v1/
    THEN it should be accessible"""
    response = client.get("/api/v1/")
    # Should not return 404
    assert response.status_code != 404

def test_api_root_redirects_to_docs():
    """GIVEN a FastAPI application
    WHEN accessing root /
    THEN it should provide API documentation"""
    response = client.get("/")
    assert response.status_code == 200
```

**Test 3: Configuration** (tests/test_config.py)
```python
import pytest
from app.core.config import Settings

def test_settings_loads_from_env(monkeypatch):
    """GIVEN environment variables set
    WHEN Settings is instantiated
    THEN it should load configuration"""
    monkeypatch.setenv("API_VERSION", "v1")
    monkeypatch.setenv("DEBUG", "true")

    settings = Settings()
    assert settings.API_VERSION == "v1"
    assert settings.DEBUG is True

def test_settings_validates_required_fields():
    """GIVEN missing required environment variables
    WHEN Settings is instantiated
    THEN it should raise validation error"""
    with pytest.raises(ValueError):
        Settings(SECRET_KEY="")  # Empty secret key should fail
```

### Implementation Code (Write AFTER Tests Pass)

**app/main.py**
```python
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.core.config import settings
from app.api.v1.router import api_router

app = FastAPI(
    title="AINative Code Backend",
    version=settings.API_VERSION,
    description="Authentication and Chat Completion Backend"
)

# CORS configuration
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.ALLOWED_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Health check
@app.get("/health")
async def health_check():
    return {
        "status": "healthy",
        "version": settings.API_VERSION
    }

# API routes
app.include_router(api_router, prefix="/api/v1")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

**app/core/config.py**
```python
from pydantic_settings import BaseSettings
from typing import List

class Settings(BaseSettings):
    API_VERSION: str = "v1"
    DEBUG: bool = False
    SECRET_KEY: str
    ALLOWED_ORIGINS: List[str] = ["http://localhost:3000"]

    class Config:
        env_file = ".env"
        case_sensitive = True

settings = Settings()
```

### Running Tests (TDD Cycle)
```bash
# 1. Write tests first (Red)
# tests/test_health.py created with failing tests

# 2. Run tests and verify they fail
pytest tests/test_health.py -v
# Expected: FAILED (no implementation yet)

# 3. Implement minimal code (Green)
# app/main.py created with health endpoint

# 4. Run tests again
pytest tests/test_health.py -v
# Expected: PASSED

# 5. Check coverage
pytest --cov=app --cov-report=term-missing
# Expected: 80%+ coverage

# 6. Refactor if needed, tests stay green
```

## Dependencies
None (first issue in the new approach)

## Definition of Done
- [ ] All tests written FIRST and passing
- [ ] Code coverage >= 80% (measured with pytest-cov)
- [ ] FastAPI application runs successfully
- [ ] Health endpoint accessible
- [ ] API documentation generated
- [ ] Configuration validated
- [ ] README.md complete with setup steps
- [ ] Code follows project coding standards
- [ ] PR approved and merged
- [ ] Tests run in CI/CD pipeline

## Estimated Effort
**Size:** Small (4-8 hours)

**Breakdown:**
- TDD Test Writing: 2 hours
- Implementation: 2 hours
- Configuration & Setup: 2 hours
- Documentation: 1 hour
- Review & Refinement: 1 hour

## Labels
- `P0` - Critical priority
- `feature` - New feature
- `backend` - Backend component
- `size:S` - Small effort
- `tdd` - Test-driven development

## Milestone
Foundation Complete (Week 1)

## Notes
- **TDD is MANDATORY**: Write tests before implementation
- **80% coverage is NON-NEGOTIABLE**: Check coverage on every commit
- Start simple, add complexity incrementally
- Use pytest fixtures for test setup
- Follow BDD-style test naming (GIVEN/WHEN/THEN)
- Use type hints throughout
- Run `black` and `ruff` for code formatting
