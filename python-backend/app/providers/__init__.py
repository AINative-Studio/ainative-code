"""LLM Provider System - Provider Registry and Factory.

This module provides a factory pattern for creating LLM provider instances.
"""
from typing import Dict, Type, Optional, Any
from .base import BaseAIProvider
from .anthropic import AnthropicProvider

# Registry of available providers
PROVIDERS: Dict[str, Type[BaseAIProvider]] = {
    "anthropic": AnthropicProvider,
}


def get_provider(name: str, api_key: Optional[str] = None, **kwargs) -> BaseAIProvider:
    """
    Get provider instance by name.

    Args:
        name: Provider name (e.g., 'anthropic')
        api_key: API key for the provider
        **kwargs: Additional provider configuration (base_url, model_id, config, etc.)

    Returns:
        Provider instance

    Raises:
        ValueError: If provider not found or API key not provided

    Example:
        >>> provider = get_provider("anthropic", api_key="sk-ant-xxx")
        >>> provider = get_provider("anthropic", api_key="sk-ant-xxx", model_id="claude-sonnet-4-5")
    """
    # Validate API key
    if not api_key:
        raise ValueError("API key is required")

    # Get provider class from registry
    provider_class = PROVIDERS.get(name)
    if not provider_class:
        raise ValueError(f"Unknown provider: {name}")

    # Instantiate and return provider
    return provider_class(api_key=api_key, **kwargs)
