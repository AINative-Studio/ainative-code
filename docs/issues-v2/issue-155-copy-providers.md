# Issue #155: Copy and Integrate LLM Provider System (TDD)

## Priority
P0 - CRITICAL

## Description
Copy production-ready LLM provider implementations from AINative platform and integrate into Python backend. Focus on Anthropic provider initially, with framework for adding OpenAI, Google, etc. **All code must be written following strict TDD principles.**

## TDD Workflow (MANDATORY)
⚠️ **CRITICAL: Tests MUST be written FIRST before copying/adapting provider code**

### Test-First Steps:
1. Write failing tests for provider base class interface
2. Write failing tests for Anthropic provider initialization
3. Write failing tests for chat completion (non-streaming)
4. Write failing tests for streaming chat completion
5. Copy and adapt code to make tests pass
6. Achieve 80%+ coverage before PR

## Acceptance Criteria

### Provider Base Class (TDD - Tests First!)
- [ ] **FIRST**: Write test for provider interface contract
- [ ] **THEN**: Copy base_provider.py from AINative platform
- [ ] **FIRST**: Write test for provider initialization
- [ ] **THEN**: Adapt base provider class
- [ ] **FIRST**: Write test for error handling in providers
- [ ] **THEN**: Implement error handling

### Anthropic Provider (TDD - Tests First!)
- [ ] **FIRST**: Write test for Anthropic provider initialization
- [ ] **THEN**: Copy and adapt anthropic_provider.py
- [ ] **FIRST**: Write test for non-streaming chat completion
- [ ] **THEN**: Implement non-streaming chat
- [ ] **FIRST**: Write test for streaming chat completion  
- [ ] **THEN**: Implement streaming chat
- [ ] **FIRST**: Write test for API key validation
- [ ] **THEN**: Implement API key validation
- [ ] **FIRST**: Write test for rate limit handling
- [ ] **THEN**: Integrate rate limiting

### Provider Registry (TDD - Tests First!)
- [ ] **FIRST**: Write test for provider registration
- [ ] **THEN**: Implement provider registry
- [ ] **FIRST**: Write test for provider lookup by name
- [ ] **THEN**: Implement provider lookup
- [ ] **FIRST**: Write test for unknown provider handling
- [ ] **THEN**: Implement error handling for unknown providers

### Testing Requirements (80%+ Coverage MANDATORY)
- [ ] Unit tests for provider base class
- [ ] Unit tests for Anthropic provider methods
- [ ] Integration tests with mock Anthropic API
- [ ] Tests for streaming response handling
- [ ] Tests for error scenarios (API errors, timeouts)
- [ ] Tests for retry logic
- [ ] Coverage >= 80% measured with pytest-cov

## Technical Requirements

### Files to Copy
**Source:**
```
/Users/aideveloper/core/src/backend/app/providers/base_provider.py
/Users/aideveloper/core/src/backend/app/providers/anthropic_provider.py
```

**Destination:**
```
python-backend/
├── app/
│   └── providers/
│       ├── __init__.py
│       ├── base.py       ← Copy from base_provider.py
│       └── anthropic.py  ← Copy from anthropic_provider.py
└── tests/
    └── providers/
        ├── test_base_provider.py      ← Write FIRST!
        ├── test_anthropic_provider.py ← Write FIRST!
        └── test_provider_registry.py  ← Write FIRST!
```

### TDD Test Examples (Write FIRST!)

**Test 1: Provider Base Class** (tests/providers/test_base_provider.py)
```python
import pytest
from app.providers.base import BaseProvider

def test_base_provider_is_abstract():
    """GIVEN BaseProvider class
    WHEN trying to instantiate directly
    THEN it should raise TypeError"""
    with pytest.raises(TypeError):
        BaseProvider(api_key="test")

def test_provider_must_implement_chat():
    """GIVEN a provider subclass
    WHEN chat method is not implemented
    THEN instantiation should fail"""
    class IncompleteProvider(BaseProvider):
        pass

    with pytest.raises(TypeError):
        IncompleteProvider(api_key="test")

def test_provider_requires_api_key():
    """GIVEN a provider subclass
    WHEN initialized without API key
    THEN it should raise ValueError"""
    class TestProvider(BaseProvider):
        async def chat(self, messages, **kwargs):
            pass

    with pytest.raises(ValueError, match="API key is required"):
        TestProvider(api_key="")
```

**Test 2: Anthropic Provider** (tests/providers/test_anthropic_provider.py)
```python
import pytest
from unittest.mock import AsyncMock, patch, MagicMock
from app.providers.anthropic import AnthropicProvider

@pytest.fixture
def anthropic_provider():
    """Create Anthropic provider with test API key"""
    return AnthropicProvider(api_key="sk-ant-test123")

def test_anthropic_provider_initialization(anthropic_provider):
    """GIVEN valid API key
    WHEN AnthropicProvider is initialized
    THEN it should set up client correctly"""
    assert anthropic_provider.api_key == "sk-ant-test123"
    assert anthropic_provider.name == "anthropic"

@pytest.mark.asyncio
async def test_chat_completion_with_valid_request(anthropic_provider):
    """GIVEN a valid chat request
    WHEN chat method is called
    THEN it should return completion response"""
    messages = [
        {"role": "user", "content": "Hello, Claude!"}
    ]

    with patch.object(
        anthropic_provider.client.messages,
        'create',
        new_callable=AsyncMock
    ) as mock_create:
        mock_create.return_value = MagicMock(
            content=[MagicMock(text="Hello! How can I help you?")],
            model="claude-sonnet-4-5",
            usage=MagicMock(input_tokens=10, output_tokens=8)
        )

        response = await anthropic_provider.chat(
            messages=messages,
            model="claude-sonnet-4-5",
            max_tokens=1024
        )

        assert response["content"] == "Hello! How can I help you?"
        assert response["model"] == "claude-sonnet-4-5"
        assert "usage" in response

@pytest.mark.asyncio
async def test_chat_completion_handles_api_error(anthropic_provider):
    """GIVEN an API error from Anthropic
    WHEN chat method is called
    THEN it should raise appropriate exception"""
    messages = [{"role": "user", "content": "Test"}]

    with patch.object(
        anthropic_provider.client.messages,
        'create',
        side_effect=Exception("API Error")
    ):
        with pytest.raises(Exception, match="API Error"):
            await anthropic_provider.chat(messages=messages)

@pytest.mark.asyncio
async def test_streaming_chat_completion(anthropic_provider):
    """GIVEN a streaming chat request
    WHEN chat_stream method is called
    THEN it should yield response chunks"""
    messages = [{"role": "user", "content": "Count to 3"}]

    # Mock streaming response
    async def mock_stream():
        yield MagicMock(delta=MagicMock(text="1"))
        yield MagicMock(delta=MagicMock(text="2"))
        yield MagicMock(delta=MagicMock(text="3"))

    with patch.object(
        anthropic_provider.client.messages,
        'create',
        return_value=mock_stream()
    ):
        chunks = []
        async for chunk in anthropic_provider.chat_stream(messages=messages):
            chunks.append(chunk)

        assert len(chunks) == 3
        assert "".join(chunks) == "123"

def test_anthropic_provider_validates_model():
    """GIVEN invalid model name
    WHEN chat is called
    THEN it should raise ValueError"""
    provider = AnthropicProvider(api_key="sk-ant-test")

    with pytest.raises(ValueError, match="Invalid model"):
        provider.chat(
            messages=[{"role": "user", "content": "test"}],
            model="invalid-model-name"
        )
```

**Test 3: Provider Registry** (tests/providers/test_provider_registry.py)
```python
import pytest
from app.providers import get_provider, register_provider, PROVIDERS
from app.providers.anthropic import AnthropicProvider

def test_register_provider():
    """GIVEN a provider class
    WHEN register_provider is called
    THEN it should be added to registry"""
    class TestProvider:
        pass

    register_provider("test", TestProvider)
    assert "test" in PROVIDERS
    assert PROVIDERS["test"] == TestProvider

def test_get_provider_returns_correct_instance():
    """GIVEN registered Anthropic provider
    WHEN get_provider is called with 'anthropic'
    THEN it should return AnthropicProvider instance"""
    provider = get_provider("anthropic", api_key="sk-ant-test")

    assert isinstance(provider, AnthropicProvider)
    assert provider.api_key == "sk-ant-test"

def test_get_provider_with_unknown_provider():
    """GIVEN unknown provider name
    WHEN get_provider is called
    THEN it should raise ValueError"""
    with pytest.raises(ValueError, match="Unknown provider"):
        get_provider("unknown-provider", api_key="test")

def test_get_provider_without_api_key():
    """GIVEN provider name without API key
    WHEN get_provider is called
    THEN it should raise ValueError"""
    with pytest.raises(ValueError, match="API key is required"):
        get_provider("anthropic", api_key="")
```

### Implementation Guide (After Tests Pass)

**app/providers/__init__.py**
```python
from typing import Dict, Type
from .base import BaseProvider
from .anthropic import AnthropicProvider

PROVIDERS: Dict[str, Type[BaseProvider]] = {
    "anthropic": AnthropicProvider,
}

def register_provider(name: str, provider_class: Type[BaseProvider]):
    """Register a new provider"""
    PROVIDERS[name] = provider_class

def get_provider(name: str, api_key: str) -> BaseProvider:
    """Get provider instance by name"""
    if not api_key:
        raise ValueError("API key is required")

    provider_class = PROVIDERS.get(name)
    if not provider_class:
        raise ValueError(f"Unknown provider: {name}")

    return provider_class(api_key=api_key)
```

### TDD Cycle
```bash
# 1. Write tests first (Red)
pytest tests/providers/test_anthropic_provider.py -v
# Expected: FAILED

# 2. Copy and adapt code (Green)
# Copy files, update imports

# 3. Run tests again
pytest tests/providers/test_anthropic_provider.py -v
# Expected: PASSED

# 4. Check coverage
pytest --cov=app.providers --cov-report=term-missing
# Expected: >= 80%
```

## Dependencies
- Issue #153: Python backend setup

## Definition of Done
- [ ] All tests written FIRST and passing
- [ ] Code coverage >= 80% for providers module
- [ ] Base provider class copied and adapted
- [ ] Anthropic provider copied and working
- [ ] Provider registry implemented
- [ ] Mock tests for API calls
- [ ] Streaming and non-streaming both work
- [ ] Error handling tested
- [ ] Code follows coding standards
- [ ] PR approved and merged

## Estimated Effort
**Size:** Medium (1-2 days)

## Labels
- `P0`, `feature`, `backend`, `provider`, `size:M`, `tdd`

## Milestone
Foundation Complete (Week 1)
