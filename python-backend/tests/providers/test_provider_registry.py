"""
Test Provider Registry - Written FIRST following TDD (RED phase)

Tests for provider factory and registry system.
"""
import pytest


def test_get_provider_returns_anthropic_instance():
    """
    GIVEN registered Anthropic provider
    WHEN get_provider is called with 'anthropic'
    THEN it should return AnthropicProvider instance
    """
    from app.providers import get_provider
    from app.providers.anthropic import AnthropicProvider

    provider = get_provider("anthropic", api_key="sk-ant-test")

    assert isinstance(provider, AnthropicProvider)
    assert provider.api_key == "sk-ant-test"


def test_get_provider_with_unknown_provider():
    """
    GIVEN unknown provider name
    WHEN get_provider is called
    THEN it should raise ValueError
    """
    from app.providers import get_provider

    with pytest.raises(ValueError, match="Unknown provider"):
        get_provider("unknown-provider", api_key="test")


def test_get_provider_without_api_key():
    """
    GIVEN provider name without API key
    WHEN get_provider is called
    THEN it should raise ValueError
    """
    from app.providers import get_provider

    with pytest.raises(ValueError, match="API key is required"):
        get_provider("anthropic", api_key="")


def test_get_provider_with_none_api_key():
    """
    GIVEN provider name with None API key
    WHEN get_provider is called
    THEN it should raise ValueError
    """
    from app.providers import get_provider

    with pytest.raises(ValueError, match="API key is required"):
        get_provider("anthropic", api_key=None)


def test_providers_registry_contains_anthropic():
    """
    GIVEN PROVIDERS registry
    WHEN checking available providers
    THEN it should contain 'anthropic'
    """
    from app.providers import PROVIDERS

    assert "anthropic" in PROVIDERS
    assert PROVIDERS["anthropic"] is not None


def test_get_provider_with_custom_config():
    """
    GIVEN provider with custom configuration
    WHEN get_provider is called with kwargs
    THEN it should pass configuration to provider
    """
    from app.providers import get_provider

    provider = get_provider(
        "anthropic",
        api_key="sk-ant-test",
        model_id="claude-sonnet-4-5",
        config={"max_retries": 5}
    )

    assert provider.model_id == "claude-sonnet-4-5"
    assert provider.config["max_retries"] == 5


def test_get_provider_with_base_url():
    """
    GIVEN provider with custom base_url
    WHEN get_provider is called
    THEN it should configure provider with base_url
    """
    from app.providers import get_provider

    provider = get_provider(
        "anthropic",
        api_key="sk-ant-test",
        base_url="https://custom.api.com"
    )

    assert provider.base_url == "https://custom.api.com"


def test_providers_registry_is_dict():
    """
    GIVEN PROVIDERS constant
    WHEN checking type
    THEN it should be a dictionary
    """
    from app.providers import PROVIDERS

    assert isinstance(PROVIDERS, dict)


def test_all_registered_providers_are_classes():
    """
    GIVEN PROVIDERS registry
    WHEN checking all values
    THEN they should all be classes (not instances)
    """
    from app.providers import PROVIDERS
    from inspect import isclass

    for name, provider_class in PROVIDERS.items():
        assert isclass(provider_class), f"Provider '{name}' should be a class"
