"""
Test Base Provider - Written FIRST following TDD (RED phase)

These tests are written BEFORE implementation to drive the design.
"""
import pytest
from abc import ABC


def test_base_provider_cannot_be_instantiated():
    """
    GIVEN BaseAIProvider abstract class
    WHEN trying to instantiate directly
    THEN it should raise TypeError
    """
    from app.providers.base import BaseAIProvider

    with pytest.raises(TypeError, match="Can't instantiate abstract class"):
        BaseAIProvider(api_key="test")


def test_base_provider_requires_provider_name_implementation():
    """
    GIVEN a concrete provider without _get_provider_name implemented
    WHEN trying to instantiate
    THEN it should raise TypeError
    """
    from app.providers.base import BaseAIProvider

    class IncompleteProvider(BaseAIProvider):
        pass

    with pytest.raises(TypeError):
        IncompleteProvider(api_key="test")


def test_base_provider_stores_configuration():
    """
    GIVEN a complete provider implementation
    WHEN initialized with configuration
    THEN it should store all configuration parameters
    """
    from app.providers.base import BaseAIProvider
    from app.schemas.ai import AIProviderSchema

    class TestProvider(BaseAIProvider):
        def _get_provider_name(self) -> AIProviderSchema:
            return AIProviderSchema.ANTHROPIC

        async def _generate_completion_impl(self, request):
            return {}

        async def _generate_chat_completion_impl(self, request):
            return {}

        async def generate_embeddings(self, texts):
            return []

        def get_token_count(self, text):
            return 0

    provider = TestProvider(
        api_key="test-key",
        base_url="https://test.api",
        model_id="test-model",
        config={"max_tokens": 1000}
    )

    assert provider.api_key == "test-key"
    assert provider.base_url == "https://test.api"
    assert provider.model_id == "test-model"
    assert provider.config["max_tokens"] == 1000


def test_base_provider_abstract_methods_required():
    """
    GIVEN BaseAIProvider
    WHEN creating a subclass
    THEN it must implement all abstract methods
    """
    from app.providers.base import BaseAIProvider
    from app.schemas.ai import AIProviderSchema

    # Missing all abstract methods
    class BadProvider(BaseAIProvider):
        pass

    with pytest.raises(TypeError):
        BadProvider(api_key="test")

    # Missing some abstract methods
    class PartialProvider(BaseAIProvider):
        def _get_provider_name(self) -> AIProviderSchema:
            return AIProviderSchema.ANTHROPIC

    with pytest.raises(TypeError):
        PartialProvider(api_key="test")


@pytest.mark.asyncio
async def test_base_provider_cacheable_logic():
    """
    GIVEN a provider with caching logic
    WHEN request has temperature > 0
    THEN request should not be cacheable
    """
    from app.providers.base import BaseAIProvider
    from app.schemas.ai import AIRequestBase, AIProviderSchema
    from app.models.ai import ModelCapabilityType

    class TestProvider(BaseAIProvider):
        def _get_provider_name(self) -> AIProviderSchema:
            return AIProviderSchema.ANTHROPIC

        async def _generate_completion_impl(self, request):
            return {"content": "test"}

        async def _generate_chat_completion_impl(self, request):
            return {"content": "test"}

        async def generate_embeddings(self, texts):
            return []

        def get_token_count(self, text):
            return 0

    provider = TestProvider(api_key="test")

    # Temperature > 0 should not be cacheable
    request_with_temp = AIRequestBase(
        prompt="test",
        capability=ModelCapabilityType.COMPLETION,
        temperature=0.7
    )
    assert not provider._is_cacheable(request_with_temp)

    # Temperature = 0 should be cacheable
    request_deterministic = AIRequestBase(
        prompt="test",
        capability=ModelCapabilityType.COMPLETION,
        temperature=0.0
    )
    assert provider._is_cacheable(request_deterministic)


@pytest.mark.asyncio
async def test_base_provider_caching_disabled_config():
    """
    GIVEN a provider with caching disabled in config
    WHEN checking if request is cacheable
    THEN it should return False regardless of temperature
    """
    from app.providers.base import BaseAIProvider
    from app.schemas.ai import AIRequestBase, AIProviderSchema
    from app.models.ai import ModelCapabilityType

    class TestProvider(BaseAIProvider):
        def _get_provider_name(self) -> AIProviderSchema:
            return AIProviderSchema.ANTHROPIC

        async def _generate_completion_impl(self, request):
            return {"content": "test"}

        async def _generate_chat_completion_impl(self, request):
            return {"content": "test"}

        async def generate_embeddings(self, texts):
            return []

        def get_token_count(self, text):
            return 0

    provider = TestProvider(
        api_key="test",
        config={"disable_caching": True}
    )

    request = AIRequestBase(
        prompt="test",
        capability=ModelCapabilityType.COMPLETION,
        temperature=0.0
    )

    assert not provider._is_cacheable(request)


@pytest.mark.asyncio
async def test_base_provider_non_cacheable_params():
    """
    GIVEN a request with stream or user parameters
    WHEN checking if cacheable
    THEN it should return False
    """
    from app.providers.base import BaseAIProvider
    from app.schemas.ai import AIRequestBase, AIProviderSchema
    from app.models.ai import ModelCapabilityType

    class TestProvider(BaseAIProvider):
        def _get_provider_name(self) -> AIProviderSchema:
            return AIProviderSchema.ANTHROPIC

        async def _generate_completion_impl(self, request):
            return {"content": "test"}

        async def _generate_chat_completion_impl(self, request):
            return {"content": "test"}

        async def generate_embeddings(self, texts):
            return []

        def get_token_count(self, text):
            return 0

    provider = TestProvider(api_key="test")

    # Test with stream parameter
    request_stream = AIRequestBase(
        prompt="test",
        capability=ModelCapabilityType.COMPLETION,
        temperature=0.0,
        additional_params={"stream": True}
    )
    assert not provider._is_cacheable(request_stream)

    # Test with user parameter
    request_user = AIRequestBase(
        prompt="test",
        capability=ModelCapabilityType.COMPLETION,
        temperature=0.0,
        additional_params={"user": "user123"}
    )
    assert not provider._is_cacheable(request_user)

    # Test with request_id parameter
    request_id = AIRequestBase(
        prompt="test",
        capability=ModelCapabilityType.COMPLETION,
        temperature=0.0,
        additional_params={"request_id": "req_123"}
    )
    assert not provider._is_cacheable(request_id)


@pytest.mark.asyncio
async def test_base_provider_with_cache_service():
    """
    GIVEN a provider with cache service
    WHEN generating completion
    THEN it should check cache and store result
    """
    from app.providers.base import BaseAIProvider
    from app.schemas.ai import AIRequestBase, AIResponse, AIProviderSchema
    from app.models.ai import ModelCapabilityType
    from unittest.mock import AsyncMock, MagicMock
    from datetime import datetime
    from uuid import uuid4

    class TestProvider(BaseAIProvider):
        def _get_provider_name(self) -> AIProviderSchema:
            return AIProviderSchema.ANTHROPIC

        async def _generate_completion_impl(self, request):
            return AIResponse(
                content="generated",
                model_id=uuid4(),
                model_name="test-model",
                provider="test",
                usage={"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15},
                created_at=datetime.now(),
                total_tokens=15,
                prompt_tokens=10,
                completion_tokens=5
            )

        async def _generate_chat_completion_impl(self, request):
            return AIResponse(
                content="chat response",
                model_id=uuid4(),
                model_name="test-model",
                provider="test",
                usage={"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15},
                created_at=datetime.now(),
                total_tokens=15,
                prompt_tokens=10,
                completion_tokens=5
            )

        async def generate_embeddings(self, texts):
            return []

        def get_token_count(self, text):
            return 0

    # Mock cache service
    mock_cache = MagicMock()
    mock_cache.get_cached_response = AsyncMock(return_value=None)
    mock_cache.cache_response = AsyncMock()

    provider = TestProvider(api_key="test", cache_service=mock_cache)

    request = AIRequestBase(
        prompt="test",
        capability=ModelCapabilityType.COMPLETION,
        temperature=0.0
    )

    response = await provider.generate_completion(request)

    assert response.content == "generated"
    assert mock_cache.get_cached_response.called
    assert mock_cache.cache_response.called
