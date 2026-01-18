# AINative Code Authentication Integration - Quick Start Guide

## Overview
This guide provides a step-by-step roadmap to integrate AINative platform components into ainative-code. Estimated integration time: **2-3 weeks** (70-85% faster than building from scratch).

---

## Day 1-2: Core Authentication Setup

### Step 1: Copy Base Authentication Files
```bash
# Copy authentication core modules
cp /Users/aideveloper/core/src/backend/app/core/auth.py \
   src/core/auth.py

cp /Users/aideveloper/core/src/backend/app/core/security_enhanced.py \
   src/core/security.py

# Copy authentication schemas
cp /Users/aideveloper/core/src/backend/app/schemas/auth.py \
   src/schemas/auth.py
```

### Step 2: Update Import Paths
In the copied files, update:
- `from app.core.database import get_db` → `from src.db import get_db`
- `from app.models.user import User` → `from src.models import User`
- `from app.core.security import ...` → `from src.core.security import ...`

### Step 3: Create Auth Endpoints
```bash
cp /Users/aideveloper/core/src/backend/app/api/v1/endpoints/auth.py \
   src/api/endpoints/auth.py
```

### Step 4: Configure Environment
Create `.env` file with:
```
SECRET_KEY=<generate-with: python -c 'import secrets; print(secrets.token_urlsafe(32))'>
ALGORITHM=HS256
ACCESS_TOKEN_EXPIRE_MINUTES=480
DATABASE_URL=postgresql://user:password@localhost:5432/ainative_code
```

---

## Day 3-5: LLM Provider Integration

### Step 1: Copy Provider Base Class
```bash
cp /Users/aideveloper/core/src/backend/app/providers/base_provider.py \
   src/providers/base.py
```

### Step 2: Copy Anthropic Reference Implementation
```bash
cp /Users/aideveloper/core/src/backend/app/providers/anthropic_provider.py \
   src/providers/anthropic.py
```

### Step 3: Create Provider Configuration
Add to your config.py:
```python
class ProviderConfig:
    ANTHROPIC_API_KEY: str = os.getenv("ANTHROPIC_API_KEY")
    OPENAI_API_KEY: str = os.getenv("OPENAI_API_KEY")
    DEFAULT_PROVIDER: str = "anthropic"
    DEFAULT_MODEL: str = "claude-sonnet-4-5"
```

### Step 4: Implement Additional Providers (Optional)
```bash
# If needed, also copy:
cp /Users/aideveloper/core/src/backend/app/providers/openai_provider.py \
   src/providers/openai.py

cp /Users/aideveloper/core/src/backend/app/providers/google_provider.py \
   src/providers/google.py
```

### Step 5: Create Provider Registry
```python
# src/providers/__init__.py
from .base import BaseAIProvider
from .anthropic import AnthropicProvider
from .openai import OpenAIProvider

PROVIDERS = {
    "anthropic": AnthropicProvider,
    "openai": OpenAIProvider,
}

def get_provider(name: str, api_key: str) -> BaseAIProvider:
    provider_class = PROVIDERS.get(name)
    if not provider_class:
        raise ValueError(f"Unknown provider: {name}")
    return provider_class(api_key=api_key)
```

---

## Day 5-7: Chat Infrastructure

### Step 1: Copy Chat Schemas
```bash
cp /Users/aideveloper/core/src/backend/app/schemas/chat.py \
   src/schemas/chat.py
```

### Step 2: Copy Chat Service
```bash
cp /Users/aideveloper/core/src/backend/app/services/managed_chat_service.py \
   src/services/chat.py
```

### Step 3: Create Chat Endpoints
Create `src/api/endpoints/chat.py`:
```python
from fastapi import APIRouter, Depends
from src.schemas.chat import ChatCompletionRequest, ChatCompletionResponse
from src.services.chat import ChatService
from src.api.deps import get_current_user

router = APIRouter(prefix="/chat", tags=["Chat"])
chat_service = ChatService()

@router.post("/completions", response_model=ChatCompletionResponse)
async def create_completion(
    request: ChatCompletionRequest,
    current_user = Depends(get_current_user)
):
    result = await chat_service.execute_completion(
        user_id=current_user.id,
        messages=request.messages,
        model=request.model,
        temperature=request.temperature,
        max_tokens=request.max_tokens
    )
    return result
```

### Step 4: Database Models
Create user-related tables if not already present:
```python
# Run alembic migrations
alembic upgrade head
```

---

## Day 8-10: Infrastructure Setup

### Step 1: Copy Rate Limiting
```bash
cp /Users/aideveloper/core/src/backend/app/services/rate_limiter.py \
   src/services/rate_limiter.py
```

Add to main.py:
```python
from src.services.rate_limiter import limiter

app = FastAPI()
app.state.limiter = limiter
```

### Step 2: Copy Circuit Breaker
```bash
cp /Users/aideveloper/core/src/backend/app/core/circuit_breaker.py \
   src/core/circuit_breaker.py
```

Use in providers:
```python
from src.core.circuit_breaker import get_circuit_breaker

anthropic_breaker = get_circuit_breaker(
    name="anthropic_api",
    failure_threshold=5,
    timeout_seconds=60.0
)

@anthropic_breaker.protect
async def call_anthropic_api():
    # Your API call here
    pass
```

### Step 3: Copy Error Handling
```bash
cp /Users/aideveloper/core/src/backend/app/core/errors.py \
   src/core/errors.py
```

### Step 4: Configure Caching (Optional)
If needed for performance:
```bash
cp /Users/aideveloper/core/src/backend/app/core/cache.py \
   src/core/cache.py
```

---

## Day 10-14: Streaming & WebSocket (Optional)

### Step 1: WebSocket Support
```bash
cp /Users/aideveloper/core/src/backend/app/websockets/stream_chat.py \
   src/websockets/stream_chat.py
```

### Step 2: Streaming Endpoints
```python
@router.websocket("/ws/chat/{session_id}")
async def websocket_endpoint(websocket: WebSocket, session_id: str):
    # Use StreamChatManager for broadcast handling
    manager = StreamChatManager()
    await manager.connect(session_id, websocket, user)
    # ... message handling
```

---

## Testing Setup

### Copy Test Configuration
```bash
cp /Users/aideveloper/core/tests/conftest.py \
   tests/conftest.py
```

### Create Test Fixtures
```python
# tests/conftest.py
import pytest
from src.core.security import create_access_token

@pytest.fixture
def test_user_token():
    return create_access_token({
        "sub": "test@example.com",
        "role": "USER"
    })

@pytest.fixture
async def auth_headers(test_user_token):
    return {"Authorization": f"Bearer {test_user_token}"}
```

---

## Production Deployment Checklist

### Pre-Deployment
- [ ] All environment variables configured
- [ ] Database migrations run
- [ ] Redis configured (if using caching/rate limiting)
- [ ] API keys secured in secrets manager
- [ ] SSL/TLS certificates configured
- [ ] CORS origins configured
- [ ] Load testing completed

### Security Hardening
- [ ] JWT SECRET_KEY is strong (32+ chars, random)
- [ ] Database connections encrypted (SSL)
- [ ] API rate limits tuned for expected load
- [ ] Circuit breaker thresholds configured
- [ ] Error messages don't leak sensitive info
- [ ] Logging enabled but doesn't log sensitive data
- [ ] HTTPS enforced

### Monitoring
- [ ] Sentry integration configured
- [ ] Logging system set up
- [ ] Metrics collection enabled
- [ ] Health check endpoints implemented
- [ ] Circuit breaker state monitoring

---

## Integration Timeline Summary

| Phase | Duration | Tasks |
|-------|----------|-------|
| Phase 1 | 2-3 days | Core auth, schemas, endpoints |
| Phase 2 | 3-5 days | Provider setup, Anthropic integration |
| Phase 3 | 5-7 days | Chat infrastructure, endpoints |
| Phase 4 | 3-5 days | Rate limiting, circuit breaker, errors |
| Phase 5 | 2-3 days | Testing, configuration |
| Phase 6 | 2-3 days | Deployment, monitoring |
| **Total** | **17-26 days** | **Complete integration** |

---

## Key Files Reference

### Authentication
- `/Users/aideveloper/core/src/backend/app/core/auth.py` - JWT verification
- `/Users/aideveloper/core/src/backend/app/core/security_enhanced.py` - Password hashing
- `/Users/aideveloper/core/src/backend/app/schemas/auth.py` - Schemas
- `/Users/aideveloper/core/src/backend/app/api/v1/endpoints/auth.py` - Endpoints

### Providers
- `/Users/aideveloper/core/src/backend/app/providers/base_provider.py` - Base class
- `/Users/aideveloper/core/src/backend/app/providers/anthropic_provider.py` - Reference

### Infrastructure
- `/Users/aideveloper/core/src/backend/app/services/rate_limiter.py` - Rate limiting
- `/Users/aideveloper/core/src/backend/app/core/circuit_breaker.py` - Resilience
- `/Users/aideveloper/core/src/backend/app/core/errors.py` - Error handling
- `/Users/aideveloper/core/src/backend/app/core/config.py` - Configuration

### Chat
- `/Users/aideveloper/core/src/backend/app/api/api_v1/endpoints/chat.py` - Endpoints
- `/Users/aideveloper/core/src/backend/app/services/managed_chat_service.py` - Service
- `/Users/aideveloper/core/src/backend/app/schemas/chat.py` - Schemas

---

## Common Issues & Solutions

### Issue 1: Import Errors
**Problem:** ModuleNotFoundError after copying files
**Solution:** Update all `from app.*` imports to `from src.*`

### Issue 2: Database Connection
**Problem:** SQLAlchemy connection errors
**Solution:** Ensure DATABASE_URL is set and PostgreSQL is running

### Issue 3: JWT Token Errors
**Problem:** JWT token validation fails
**Solution:** Verify SECRET_KEY matches between token creation and validation

### Issue 4: Rate Limiting Not Working
**Problem:** Requests not being rate limited
**Solution:** Ensure Redis is running and REDIS_HOST is configured

### Issue 5: Circuit Breaker Stays Open
**Problem:** Circuit breaker never recovers
**Solution:** Adjust `timeout_seconds` and `recovery_timeout` parameters

---

## Support & Documentation

For detailed information, see:
- **AINATIVE_PLATFORM_ANALYSIS.md** - Comprehensive code analysis (41KB)
- **AINative Platform Source:** `/Users/aideveloper/core/`
- **FastAPI Docs:** https://fastapi.tiangolo.com
- **Pydantic:** https://docs.pydantic.dev

---

**Estimated Integration Time: 17-26 days (2-3.5 weeks)**
**Estimated Time Savings: 50-70 days vs. building from scratch**
