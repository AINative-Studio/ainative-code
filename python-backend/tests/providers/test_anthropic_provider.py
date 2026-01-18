"""
Test Anthropic Provider - Written FIRST following TDD (RED phase)

These tests define the expected behavior before implementation.
"""
import pytest
from unittest.mock import AsyncMock, patch, MagicMock, Mock
from uuid import UUID


@pytest.fixture
def anthropic_provider():
    """Create Anthropic provider with test API key."""
    from app.providers.anthropic import AnthropicProvider

    return AnthropicProvider(api_key="sk-ant-test123")


@pytest.fixture
def mock_anthropic_client():
    """Create a mock Anthropic client."""
    mock_client = MagicMock()
    mock_client.messages = MagicMock()
    mock_client.messages.create = AsyncMock()
    return mock_client


def test_anthropic_provider_initialization():
    """
    GIVEN valid API key and configuration
    WHEN AnthropicProvider is initialized
    THEN it should set up client correctly
    """
    from app.providers.anthropic import AnthropicProvider

    provider = AnthropicProvider(
        api_key="sk-ant-test123",
        model_id="claude-sonnet-4-5",
        config={"max_retries": 5}
    )

    assert provider.api_key == "sk-ant-test123"
    assert provider.model_id == "claude-sonnet-4-5"
    assert provider.config["max_retries"] == 5
    assert provider.max_retries == 5
    assert provider.client is not None


def test_anthropic_provider_default_model():
    """
    GIVEN AnthropicProvider without model_id
    WHEN initialized
    THEN it should use default model
    """
    from app.providers.anthropic import AnthropicProvider

    provider = AnthropicProvider(api_key="sk-ant-test")

    assert provider.model_id == "claude-sonnet-4-5"


def test_anthropic_provider_custom_base_url():
    """
    GIVEN AnthropicProvider with custom base_url
    WHEN initialized
    THEN it should configure client with custom URL
    """
    from app.providers.anthropic import AnthropicProvider

    provider = AnthropicProvider(
        api_key="sk-ant-test",
        base_url="https://custom.anthropic.ai"
    )

    assert provider.base_url == "https://custom.anthropic.ai"


def test_anthropic_provider_user_id_tracking():
    """
    GIVEN AnthropicProvider with user_id
    WHEN initialized
    THEN it should store user_id for token tracking
    """
    from app.providers.anthropic import AnthropicProvider
    from uuid import uuid4

    user_id = uuid4()
    provider = AnthropicProvider(
        api_key="sk-ant-test",
        user_id=user_id
    )

    assert provider.user_id == user_id


@pytest.mark.asyncio
async def test_chat_completion_with_valid_request(anthropic_provider, mock_anthropic_client):
    """
    GIVEN a valid chat request
    WHEN chat completion is generated
    THEN it should return proper response
    """
    from app.schemas.ai import AIRequestBase
    from app.models.ai import ModelCapabilityType

    # Mock the Anthropic response
    mock_response = MagicMock()
    mock_response.id = "msg_test123"
    mock_response.role = "assistant"
    mock_response.model = "claude-sonnet-4-5"
    mock_response.stop_reason = "end_turn"

    # Mock content blocks
    mock_text_block = MagicMock()
    mock_text_block.text = "Hello! How can I help you today?"
    mock_response.content = [mock_text_block]

    # Mock usage
    mock_usage = MagicMock()
    mock_usage.input_tokens = 15
    mock_usage.output_tokens = 10
    mock_response.usage = mock_usage

    mock_anthropic_client.messages.create.return_value = mock_response
    anthropic_provider.client = mock_anthropic_client

    # Create request
    request = AIRequestBase(
        prompt="Hello, Claude!",
        capability=ModelCapabilityType.CHAT_COMPLETION
    )

    # Generate completion
    response = await anthropic_provider.generate_completion(request)

    # Verify response
    assert response.content == "Hello! How can I help you today?"
    assert response.provider == "anthropic"
    assert response.usage["prompt_tokens"] == 15
    assert response.usage["completion_tokens"] == 10
    assert response.usage["total_tokens"] == 25


@pytest.mark.asyncio
async def test_chat_completion_with_messages(anthropic_provider, mock_anthropic_client):
    """
    GIVEN a request with conversation messages
    WHEN chat completion is generated
    THEN it should format and send messages correctly
    """
    from app.schemas.ai import AIRequestBase
    from app.models.ai import ModelCapabilityType

    # Mock response
    mock_response = MagicMock()
    mock_response.id = "msg_test123"
    mock_response.role = "assistant"
    mock_response.model = "claude-sonnet-4-5"
    mock_response.stop_reason = "end_turn"

    mock_text_block = MagicMock()
    mock_text_block.text = "Response to conversation"
    mock_response.content = [mock_text_block]

    mock_usage = MagicMock()
    mock_usage.input_tokens = 20
    mock_usage.output_tokens = 15
    mock_response.usage = mock_usage

    mock_anthropic_client.messages.create.return_value = mock_response
    anthropic_provider.client = mock_anthropic_client

    # Create request with messages
    request = AIRequestBase(
        capability=ModelCapabilityType.CHAT_COMPLETION,
        messages=[
            {"role": "user", "content": "Hello"},
            {"role": "assistant", "content": "Hi there!"},
            {"role": "user", "content": "How are you?"}
        ]
    )

    response = await anthropic_provider.generate_chat_completion(request)

    assert response.content == "Response to conversation"
    assert mock_anthropic_client.messages.create.called


@pytest.mark.asyncio
async def test_chat_completion_handles_api_error(anthropic_provider, mock_anthropic_client):
    """
    GIVEN an API error from Anthropic
    WHEN chat completion is called
    THEN it should raise appropriate exception
    """
    from app.schemas.ai import AIRequestBase
    from app.models.ai import ModelCapabilityType
    from anthropic import APIStatusError
    from httpx import Response, Request

    # Mock API error with proper signature
    mock_request = Request("POST", "https://api.anthropic.com/v1/messages")
    mock_response = Response(429, request=mock_request)
    mock_error = APIStatusError(
        message="Rate limit exceeded",
        response=mock_response,
        body=None
    )
    mock_anthropic_client.messages.create.side_effect = mock_error
    anthropic_provider.client = mock_anthropic_client

    request = AIRequestBase(
        prompt="Test",
        capability=ModelCapabilityType.COMPLETION
    )

    with pytest.raises(ValueError, match="Anthropic API error"):
        await anthropic_provider.generate_completion(request)


@pytest.mark.asyncio
async def test_retry_logic_on_rate_limit(anthropic_provider, mock_anthropic_client):
    """
    GIVEN a rate limit error from Anthropic
    WHEN chat completion is called
    THEN it should retry with exponential backoff
    """
    from app.schemas.ai import AIRequestBase
    from app.models.ai import ModelCapabilityType
    from anthropic import RateLimitError
    from httpx import Response, Request

    # Mock successful response after retries
    mock_response = MagicMock()
    mock_response.id = "msg_test123"
    mock_response.role = "assistant"
    mock_response.model = "claude-sonnet-4-5"
    mock_response.stop_reason = "end_turn"

    mock_text_block = MagicMock()
    mock_text_block.text = "Success after retry"
    mock_response.content = [mock_text_block]

    mock_usage = MagicMock()
    mock_usage.input_tokens = 10
    mock_usage.output_tokens = 8
    mock_response.usage = mock_usage

    # Create proper RateLimitError
    mock_request = Request("POST", "https://api.anthropic.com/v1/messages")
    mock_http_response = Response(429, request=mock_request)
    rate_limit_error = RateLimitError(
        message="Rate limit exceeded",
        response=mock_http_response,
        body=None
    )

    # First call raises rate limit, second succeeds
    mock_anthropic_client.messages.create.side_effect = [
        rate_limit_error,
        mock_response
    ]
    anthropic_provider.client = mock_anthropic_client
    anthropic_provider.config["max_retries"] = 3
    anthropic_provider.max_retries = 3
    anthropic_provider.retry_delay = 0.01  # Fast retry for testing

    request = AIRequestBase(
        prompt="Test retry",
        capability=ModelCapabilityType.COMPLETION
    )

    response = await anthropic_provider.generate_completion(request)

    assert response.content == "Success after retry"
    assert mock_anthropic_client.messages.create.call_count == 2


@pytest.mark.asyncio
async def test_tool_registration(anthropic_provider):
    """
    GIVEN a tool definition and handler
    WHEN registering a tool
    THEN it should be added to tool registry
    """
    tool_definition = {
        "name": "test_tool",
        "description": "A test tool",
        "input_schema": {
            "type": "object",
            "properties": {
                "param": {"type": "string"}
            }
        }
    }

    async def tool_handler(param: str):
        return f"Processed: {param}"

    anthropic_provider.register_tool(tool_definition, tool_handler)

    assert len(anthropic_provider.tools) == 1
    assert "test_tool" in anthropic_provider.tool_handlers
    assert anthropic_provider.tools[0]["name"] == "test_tool"


@pytest.mark.asyncio
async def test_tool_execution(anthropic_provider):
    """
    GIVEN a registered tool
    WHEN executing the tool
    THEN it should call handler and return result
    """
    async def calculator_tool(operation: str, a: int, b: int):
        if operation == "add":
            return a + b
        return 0

    tool_def = {"name": "calculator", "description": "Calculator"}
    anthropic_provider.register_tool(tool_def, calculator_tool)

    result = await anthropic_provider._execute_tool(
        "calculator",
        {"operation": "add", "a": 5, "b": 3}
    )

    assert result["success"] is True
    assert result["result"] == 8


@pytest.mark.asyncio
async def test_tool_execution_error_handling(anthropic_provider):
    """
    GIVEN a tool that raises an error
    WHEN executing the tool
    THEN it should return error response
    """
    async def failing_tool(param: str):
        raise ValueError("Tool failed")

    tool_def = {"name": "failing_tool", "description": "Fails"}
    anthropic_provider.register_tool(tool_def, failing_tool)

    result = await anthropic_provider._execute_tool(
        "failing_tool",
        {"param": "test"}
    )

    assert result["success"] is False
    assert "Tool failed" in result["error"]


@pytest.mark.asyncio
async def test_generate_embeddings_placeholder(anthropic_provider):
    """
    GIVEN Anthropic provider (which doesn't support embeddings)
    WHEN generate_embeddings is called
    THEN it should return placeholder embeddings
    """
    texts = ["text1", "text2", "text3"]

    embeddings = await anthropic_provider.generate_embeddings(texts)

    assert len(embeddings) == 3
    assert all(len(emb) == 64 for emb in embeddings)


def test_get_token_count_estimation(anthropic_provider):
    """
    GIVEN a text string
    WHEN getting token count
    THEN it should return reasonable estimation
    """
    text = "This is a test string with multiple words"

    token_count = anthropic_provider.get_token_count(text)

    # ~4 characters per token estimation
    assert token_count > 0
    assert token_count == len(text) // 4


@pytest.mark.asyncio
async def test_extract_text_content_from_blocks(anthropic_provider):
    """
    GIVEN response with multiple content blocks
    WHEN extracting text content
    THEN it should concatenate all text blocks
    """
    from anthropic.types import TextBlock

    content_blocks = [
        TextBlock(text="Hello ", type="text"),
        TextBlock(text="World!", type="text")
    ]

    text = anthropic_provider._extract_text_content(content_blocks)

    assert text == "Hello World!"


@pytest.mark.asyncio
async def test_format_messages_from_request(anthropic_provider):
    """
    GIVEN AIRequestBase with messages
    WHEN formatting messages
    THEN it should convert to Anthropic format
    """
    from app.schemas.ai import AIRequestBase
    from app.models.ai import ModelCapabilityType

    request = AIRequestBase(
        capability=ModelCapabilityType.CHAT_COMPLETION,
        messages=[
            {"role": "system", "content": "You are helpful"},
            {"role": "user", "content": "Hello"},
            {"role": "assistant", "content": "Hi!"}
        ]
    )

    messages = anthropic_provider._format_messages(request)

    # System messages should be filtered out
    assert len(messages) == 2
    assert messages[0]["role"] == "user"
    assert messages[0]["content"] == "Hello"
    assert messages[1]["role"] == "assistant"


@pytest.mark.asyncio
async def test_extract_system_message(anthropic_provider):
    """
    GIVEN AIRequestBase with system message
    WHEN extracting system message
    THEN it should return system content
    """
    from app.schemas.ai import AIRequestBase
    from app.models.ai import ModelCapabilityType

    request = AIRequestBase(
        capability=ModelCapabilityType.CHAT_COMPLETION,
        messages=[
            {"role": "system", "content": "You are a helpful assistant"},
            {"role": "user", "content": "Hello"}
        ]
    )

    system = anthropic_provider._extract_system_message(request)

    assert system == "You are a helpful assistant"


@pytest.mark.asyncio
async def test_no_system_message_returns_none(anthropic_provider):
    """
    GIVEN AIRequestBase without system message
    WHEN extracting system message
    THEN it should return None
    """
    from app.schemas.ai import AIRequestBase
    from app.models.ai import ModelCapabilityType

    request = AIRequestBase(
        capability=ModelCapabilityType.CHAT_COMPLETION,
        messages=[
            {"role": "user", "content": "Hello"}
        ]
    )

    system = anthropic_provider._extract_system_message(request)

    assert system is None


@pytest.mark.asyncio
async def test_convenience_generate_method(anthropic_provider, mock_anthropic_client):
    """
    GIVEN simple text prompt
    WHEN using convenience generate method
    THEN it should return text string directly
    """
    # Mock response
    mock_response = MagicMock()
    mock_response.id = "msg_test"
    mock_response.role = "assistant"
    mock_response.model = "claude-sonnet-4-5"
    mock_response.stop_reason = "end_turn"

    mock_text_block = MagicMock()
    mock_text_block.text = "Generated response"
    mock_response.content = [mock_text_block]

    mock_usage = MagicMock()
    mock_usage.input_tokens = 5
    mock_usage.output_tokens = 3
    mock_response.usage = mock_usage

    mock_anthropic_client.messages.create.return_value = mock_response
    anthropic_provider.client = mock_anthropic_client

    result = await anthropic_provider.generate(
        prompt="Test prompt",
        max_tokens=100,
        temperature=0.7
    )

    assert result == "Generated response"
    assert isinstance(result, str)


@pytest.mark.asyncio
async def test_api_key_validation(anthropic_provider, mock_anthropic_client):
    """
    GIVEN AnthropicProvider without API key
    WHEN making API call
    THEN it should raise ValueError
    """
    from app.schemas.ai import AIRequestBase
    from app.models.ai import ModelCapabilityType
    from app.providers.anthropic import AnthropicProvider

    # Create provider without API key
    provider_no_key = AnthropicProvider(api_key=None)

    request = AIRequestBase(
        prompt="Test",
        capability=ModelCapabilityType.COMPLETION
    )

    with pytest.raises(ValueError, match="API key not provided"):
        await provider_no_key._create_message_with_tools(
            messages=[{"role": "user", "content": "test"}]
        )
