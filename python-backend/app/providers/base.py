"""
Base Provider Implementation - Adapted from AINative Core
Following Semantic Seed V2.0 Standards for AI provider interfaces

This module defines the base abstractions for all AI providers,
ensuring consistent interface and behavior across implementations.
"""

import logging
from abc import ABC, abstractmethod
from typing import Dict, Any, List, Optional
from app.schemas.ai import AIRequestBase, AIResponse, AIProviderSchema

logger = logging.getLogger(__name__)


class BaseAIProvider(ABC):
    """
    Base class for all AI providers

    This abstract class defines the standard interface that all provider
    implementations must adhere to, following Semantic Seed V2.0 standards.
    """

    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: Optional[str] = None,
        model_id: Optional[str] = None,
        config: Optional[Dict[str, Any]] = None,
        cache_service: Optional[Any] = None,
    ):
        """Initialize the provider with configuration"""
        self.api_key = api_key
        self.base_url = base_url
        self.model_id = model_id
        self.config = config or {}
        self.cache_service = cache_service
        self.provider_name: AIProviderSchema = self._get_provider_name()

    @abstractmethod
    def _get_provider_name(self) -> AIProviderSchema:
        """
        Get the provider name for this implementation

        Returns:
            Provider name as AIProviderSchema enum
        """
        pass

    async def generate_completion(self, request: AIRequestBase) -> AIResponse:
        """
        Generate a completion for the given request

        Args:
            request: The AI request containing prompt and parameters

        Returns:
            AIResponse with the generated content
        """
        # Check cache if available
        if self.cache_service and self._is_cacheable(request):
            cached_response = await self.cache_service.get_cached_response(
                request, self.provider_name
            )
            if cached_response:
                logger.info(f"Cache hit for {self.provider_name} completion request")
                # Mark response as cached
                cached_response.is_cached = True
                return cached_response

        # Generate new response
        response = await self._generate_completion_impl(request)

        # Cache the response if appropriate
        if self.cache_service and self._is_cacheable(request):
            await self.cache_service.cache_response(
                request, response, self.provider_name
            )

        return response

    @abstractmethod
    async def _generate_completion_impl(self, request: AIRequestBase) -> AIResponse:
        """
        Implementation of completion generation

        Args:
            request: The AI request containing prompt and parameters

        Returns:
            AIResponse with the generated content
        """
        pass

    async def generate_chat_completion(self, request: AIRequestBase) -> AIResponse:
        """
        Generate a chat completion for the given request

        Args:
            request: The AI request containing messages and parameters

        Returns:
            AIResponse with the generated content
        """
        # Check cache if available
        if self.cache_service and self._is_cacheable(request):
            cached_response = await self.cache_service.get_cached_response(
                request, self.provider_name
            )
            if cached_response:
                logger.info(
                    f"Cache hit for {self.provider_name} chat completion request"
                )
                # Mark response as cached
                cached_response.is_cached = True
                return cached_response

        # Generate new response
        response = await self._generate_chat_completion_impl(request)

        # Cache the response if appropriate
        if self.cache_service and self._is_cacheable(request):
            await self.cache_service.cache_response(
                request, response, self.provider_name
            )

        return response

    @abstractmethod
    async def _generate_chat_completion_impl(
        self, request: AIRequestBase
    ) -> AIResponse:
        """
        Implementation of chat completion generation

        Args:
            request: The AI request containing messages and parameters

        Returns:
            AIResponse with the generated content
        """
        pass

    @abstractmethod
    async def generate_embeddings(self, texts: List[str]) -> List[List[float]]:
        """
        Generate embeddings for the given texts

        Args:
            texts: List of text strings to embed

        Returns:
            List of embedding vectors
        """
        pass

    @abstractmethod
    def get_token_count(self, text: str) -> int:
        """
        Count the number of tokens in the given text

        Args:
            text: The text to count tokens for

        Returns:
            Number of tokens
        """
        pass

    def _is_cacheable(self, request: AIRequestBase) -> bool:
        """
        Determine if a request is cacheable

        Args:
            request: The AI request to check

        Returns:
            True if the request can be cached, False otherwise
        """
        # Only cache deterministic requests (temperature = 0)
        if request.temperature and request.temperature > 0:
            return False

        # Don't cache requests with specific non-repeatable parameters
        if request.additional_params and any(
            k in request.additional_params for k in ["stream", "user", "request_id"]
        ):
            return False

        # Check if caching is globally disabled for this provider
        if self.config.get("disable_caching", False):
            return False

        return True
