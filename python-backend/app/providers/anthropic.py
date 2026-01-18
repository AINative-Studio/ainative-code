"""
Anthropic Provider Implementation - Adapted from AINative Core
Following Semantic Seed V2.0 Standards for AI provider interfaces

This module implements the Anthropic provider using the modern Messages API
with full tool use support, extended thinking, and proper error handling.
"""

from typing import Dict, Any, List, Optional
import json
from uuid import UUID, uuid4
import asyncio
import random
from datetime import datetime
from anthropic import AsyncAnthropic, APIStatusError, APIConnectionError, RateLimitError, APITimeoutError
from anthropic.types import Message, ToolUseBlock, TextBlock, ContentBlock

from app.providers.base import BaseAIProvider
from app.schemas.ai import AIRequestBase, AIResponse, AIProviderSchema
import logging

logger = logging.getLogger(__name__)


class AnthropicProvider(BaseAIProvider):
    """
    Anthropic provider implementation using Messages API with tool use

    Features:
    - Modern Messages API (replaces deprecated Completion API)
    - Tool use / function calling support
    - Extended thinking for Claude 3.7 Sonnet
    - Automatic retry logic with exponential backoff
    - Comprehensive error handling
    - Token usage tracking for billing

    Handles interactions with Anthropic API endpoints following
    Semantic Seed V2.0 standards for API integration.
    """

    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: Optional[str] = None,
        model_id: Optional[str] = "claude-sonnet-4-5",
        config: Optional[Dict[str, Any]] = None,
        user_id: Optional[UUID] = None,
    ):
        """
        Initialize the Anthropic provider with proper configuration

        Args:
            api_key: Anthropic API key
            base_url: Base URL for the Anthropic API (optional, uses default)
            model_id: The model ID to use (default: claude-sonnet-4-5)
            config: Configuration options
            user_id: User ID for usage tracking
        """
        super().__init__(
            api_key=api_key,
            base_url=base_url,
            model_id=model_id,
            config=config or {}
        )

        # Store user_id for token tracking
        self.user_id = user_id

        # Initialize the AsyncAnthropic client
        client_kwargs = {"api_key": self.api_key}
        if base_url:
            client_kwargs["base_url"] = base_url

        self.client = AsyncAnthropic(**client_kwargs)

        # Configuration for retries
        self.max_retries = self.config.get("max_retries", 3)
        self.retry_delay = self.config.get("retry_delay", 1.0)

        # Tool registry for function calling
        self.tools: List[Dict[str, Any]] = []
        self.tool_handlers: Dict[str, Any] = {}

        # Log warning if API key is not set
        if not self.api_key:
            logger.warning("Anthropic API key not provided")

    def _get_provider_name(self) -> AIProviderSchema:
        """
        Return the provider name for the Anthropic implementation

        Returns:
            Provider name as AIProviderSchema enum
        """
        return AIProviderSchema.ANTHROPIC

    def register_tool(self, tool_definition: Dict[str, Any], handler: Any):
        """
        Register a tool for function calling

        Args:
            tool_definition: Tool definition following Anthropic's schema
            handler: Async function to execute when tool is called
        """
        self.tools.append(tool_definition)
        self.tool_handlers[tool_definition["name"]] = handler
        logger.info(f"Registered tool: {tool_definition['name']}")

    async def _execute_tool(self, tool_name: str, tool_input: Dict[str, Any]) -> Dict[str, Any]:
        """
        Execute a registered tool

        Args:
            tool_name: Name of the tool to execute
            tool_input: Input parameters for the tool

        Returns:
            Tool execution result
        """
        if tool_name not in self.tool_handlers:
            return {
                "error": f"Tool '{tool_name}' not found",
                "available_tools": list(self.tool_handlers.keys())
            }

        try:
            handler = self.tool_handlers[tool_name]
            result = await handler(**tool_input)
            return {"success": True, "result": result}
        except Exception as e:
            logger.error(f"Error executing tool '{tool_name}': {str(e)}")
            return {"error": str(e), "success": False}

    async def _retry_with_backoff(self, func, *args, **kwargs):
        """
        Retry a function with exponential backoff

        Args:
            func: Async function to retry
            *args, **kwargs: Arguments to pass to the function

        Returns:
            Function result

        Raises:
            Exception: If all retries fail
        """
        last_exception = None

        for attempt in range(self.max_retries):
            try:
                return await func(*args, **kwargs)
            except RateLimitError as e:
                last_exception = e
                if attempt < self.max_retries - 1:
                    wait_time = self.retry_delay * (2 ** attempt)
                    logger.warning(
                        f"Rate limit hit, retrying in {wait_time}s (attempt {attempt + 1}/{self.max_retries})"
                    )
                    await asyncio.sleep(wait_time)
                else:
                    logger.error("Max retries reached for rate limit")
            except APITimeoutError as e:
                last_exception = e
                if attempt < self.max_retries - 1:
                    wait_time = self.retry_delay * (2 ** attempt)
                    logger.warning(
                        f"Request timeout, retrying in {wait_time}s (attempt {attempt + 1}/{self.max_retries})"
                    )
                    await asyncio.sleep(wait_time)
                else:
                    logger.error("Max retries reached for timeout")
            except APIConnectionError as e:
                last_exception = e
                if attempt < self.max_retries - 1:
                    wait_time = self.retry_delay * (2 ** attempt)
                    logger.warning(
                        f"Connection error, retrying in {wait_time}s (attempt {attempt + 1}/{self.max_retries})"
                    )
                    await asyncio.sleep(wait_time)
                else:
                    logger.error("Max retries reached for connection error")
            except Exception as e:
                # Don't retry on other errors
                raise e

        # If we get here, all retries failed
        raise last_exception

    async def _create_message_with_tools(
        self,
        messages: List[Dict[str, str]],
        model: Optional[str] = None,
        max_tokens: Optional[int] = None,
        temperature: Optional[float] = None,
        system: Optional[str] = None,
        tools: Optional[List[Dict[str, Any]]] = None,
        thinking: Optional[Dict[str, Any]] = None,
    ) -> Message:
        """
        Create a message using the Anthropic Messages API with tool use support

        Args:
            messages: List of message dicts with role and content
            model: Model to use (defaults to self.model_id)
            max_tokens: Maximum tokens to generate
            temperature: Sampling temperature
            system: System prompt
            tools: List of tool definitions
            thinking: Extended thinking configuration for Claude 3.7

        Returns:
            Anthropic Message object
        """
        # Validate that API key is set
        if not self.api_key:
            logger.error("Anthropic API key not provided")
            raise ValueError("API key not provided. Please set ANTHROPIC_API_KEY environment variable.")

        # Prepare request parameters
        request_params = {
            "model": model or self.model_id,
            "max_tokens": max_tokens or self.config.get("max_tokens", 4096),
            "messages": messages,
        }

        # Add optional parameters
        if temperature is not None:
            request_params["temperature"] = temperature
        elif "temperature" in self.config:
            request_params["temperature"] = self.config["temperature"]

        if system:
            request_params["system"] = system

        if tools:
            request_params["tools"] = tools

        # Add extended thinking support for Claude 3.7 Sonnet
        if thinking:
            request_params["thinking"] = thinking

        # Make the API call with retry logic
        try:
            response = await self._retry_with_backoff(
                self.client.messages.create,
                **request_params
            )
            return response
        except APIStatusError as e:
            logger.error(f"Anthropic API error: {e.status_code} - {e.message}")
            raise ValueError(f"Anthropic API error: {e.message}")
        except Exception as e:
            logger.error(f"Unexpected error calling Anthropic API: {str(e)}")
            raise Exception(f"Failed to call Anthropic API: {str(e)}")

    def _format_messages(self, request: AIRequestBase) -> List[Dict[str, str]]:
        """
        Format messages from AIRequestBase to Anthropic format

        Args:
            request: The AI request

        Returns:
            List of formatted messages
        """
        messages = []

        # If messages are provided directly, use them
        if request.messages:
            for msg in request.messages:
                # Convert role names if needed
                role = msg.get("role", "user")
                if role == "system":
                    # System messages should be extracted separately
                    continue
                elif role in ["human", "user"]:
                    role = "user"
                elif role == "assistant":
                    role = "assistant"

                messages.append({
                    "role": role,
                    "content": msg.get("content", "")
                })
        else:
            # Create a user message from the prompt
            messages = [{"role": "user", "content": request.prompt}]

        return messages

    def _extract_system_message(self, request: AIRequestBase) -> Optional[str]:
        """
        Extract system message from request messages

        Args:
            request: The AI request

        Returns:
            System message content or None
        """
        if not request.messages:
            return None

        for msg in request.messages:
            if msg.get("role") == "system":
                return msg.get("content", "")

        return None

    def _extract_text_content(self, content: List[ContentBlock]) -> str:
        """
        Extract text content from message content blocks

        Args:
            content: List of content blocks

        Returns:
            Concatenated text content
        """
        text_parts = []
        for block in content:
            if isinstance(block, TextBlock):
                text_parts.append(block.text)
            elif hasattr(block, "text"):
                text_parts.append(block.text)

        return "".join(text_parts)

    async def _generate_completion_impl(self, request: AIRequestBase) -> AIResponse:
        """
        Implementation of completion generation using Messages API

        Args:
            request: The AI request containing prompt and parameters

        Returns:
            AIResponse with the generated content
        """
        logger.info(
            f"Generating Anthropic completion with model {self.model_id}, prompt length: {len(request.prompt) if request.prompt else 0}"
        )

        # Convert to messages format
        messages = [{"role": "user", "content": request.prompt}]

        # Extract configuration
        max_tokens = request.additional_params.get("max_tokens") if request.additional_params else None
        max_tokens = max_tokens or self.config.get("max_tokens", 4096)

        temperature = request.additional_params.get("temperature") if request.additional_params else None
        temperature = temperature if temperature is not None else self.config.get("temperature", 0.7)

        # Check for extended thinking
        thinking = None
        if request.additional_params and "thinking" in request.additional_params:
            thinking = request.additional_params["thinking"]

        # Call the API
        response = await self._create_message_with_tools(
            messages=messages,
            max_tokens=max_tokens,
            temperature=temperature,
            thinking=thinking,
        )

        # Extract text content
        content = self._extract_text_content(response.content)

        # Get usage data
        input_tokens = response.usage.input_tokens
        output_tokens = response.usage.output_tokens
        total_tokens = input_tokens + output_tokens

        # Return the response with all required fields
        return AIResponse(
            content=content,
            model_id=uuid4(),
            model_name=self.model_id,
            provider="anthropic",
            usage={
                "prompt_tokens": input_tokens,
                "completion_tokens": output_tokens,
                "total_tokens": total_tokens
            },
            finish_reason=response.stop_reason or "stop",
            created_at=datetime.now(),
            total_tokens=total_tokens,
            prompt_tokens=input_tokens,
            completion_tokens=output_tokens
        )

    async def _generate_chat_completion_impl(self, request: AIRequestBase) -> AIResponse:
        """
        Implementation of chat completion generation using Messages API

        Args:
            request: The AI request containing messages and parameters

        Returns:
            AIResponse with the generated content
        """
        logger.info(
            f"Generating Anthropic chat completion with model {self.model_id}"
        )

        # Format messages
        messages = self._format_messages(request)
        system = self._extract_system_message(request)

        # Extract configuration
        max_tokens = request.additional_params.get("max_tokens") if request.additional_params else None
        max_tokens = max_tokens or self.config.get("max_tokens", 4096)

        temperature = request.additional_params.get("temperature") if request.additional_params else None
        temperature = temperature if temperature is not None else self.config.get("temperature", 0.7)

        # Check for extended thinking
        thinking = None
        if request.additional_params and "thinking" in request.additional_params:
            thinking = request.additional_params["thinking"]

        # Call the API
        response = await self._create_message_with_tools(
            messages=messages,
            max_tokens=max_tokens,
            temperature=temperature,
            system=system,
            thinking=thinking,
        )

        # Extract text content
        content = self._extract_text_content(response.content)

        # Get usage data
        input_tokens = response.usage.input_tokens
        output_tokens = response.usage.output_tokens
        total_tokens = input_tokens + output_tokens

        # Return the response
        return AIResponse(
            content=content,
            model_id=uuid4(),
            model_name=self.model_id,
            provider="anthropic",
            usage={
                "prompt_tokens": input_tokens,
                "completion_tokens": output_tokens,
                "total_tokens": total_tokens
            },
            finish_reason=response.stop_reason or "stop",
            created_at=datetime.now(),
            total_tokens=total_tokens,
            prompt_tokens=input_tokens,
            completion_tokens=output_tokens
        )

    async def generate_embeddings(self, texts: List[str]) -> List[List[float]]:
        """
        Generate embeddings using Anthropic's API

        Note: Anthropic doesn't natively support embeddings.
        This is a placeholder that should use a different provider.

        Args:
            texts: List of text strings to embed

        Returns:
            List of embedding vectors
        """
        logger.warning(
            "Anthropic does not support embeddings natively. "
            "Consider using OpenAI, Cohere, or a dedicated embedding model."
        )

        # Placeholder embeddings for testing (64-dimensional)
        return [[0.1] * 64 for _ in texts]

    def get_token_count(self, text: str) -> int:
        """
        Count the number of tokens in the given text

        Note: This is an approximation. For accurate counts, use the
        Anthropic tokenizer library.

        Args:
            text: The text to count tokens for

        Returns:
            Estimated number of tokens
        """
        # Simple estimation: ~4 characters per token
        # For more accurate counting, use: anthropic.count_tokens()
        return len(text) // 4

    async def generate(
        self,
        prompt: str,
        max_tokens: Optional[int] = None,
        temperature: float = 0.7,
        **kwargs,
    ) -> str:
        """
        Convenience method for simple text generation.
        Wraps generate_completion and returns just the content string.

        Args:
            prompt: Text prompt to complete
            max_tokens: Maximum tokens to generate (default: 2048)
            temperature: Sampling temperature (0-2)
            **kwargs: Additional provider-specific parameters

        Returns:
            Generated text content as a string
        """
        from app.models.ai import ModelCapabilityType

        # Create a request object
        request = AIRequestBase(
            prompt=prompt,
            capability=ModelCapabilityType.COMPLETION,
            additional_params={
                "max_tokens": max_tokens or 2048,
                "temperature": temperature,
                **kwargs
            }
        )

        # Call the completion method
        response = await self.generate_completion(request)

        # Return just the content string
        return response.content if response else ""
