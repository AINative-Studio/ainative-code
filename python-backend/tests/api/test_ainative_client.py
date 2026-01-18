"""Test suite for AINative API client integration.

This test suite follows TDD principles - tests are written FIRST before implementation.
All tests use mocked HTTP calls to ensure fast, reliable testing without external dependencies.
"""
import pytest
from unittest.mock import patch, MagicMock, AsyncMock
import httpx
from fastapi import HTTPException


@pytest.fixture
def client():
    """Create AINative API client instance for testing.

    Returns:
        AINativeClient: Configured client instance
    """
    # Import will fail initially (RED phase) - this is expected in TDD
    from app.api.ainative_client import AINativeClient
    return AINativeClient(base_url="https://api.ainative.studio/v1")


class TestAINativeClientLogin:
    """Test suite for login functionality."""

    @pytest.mark.asyncio
    async def test_login_with_valid_credentials_returns_tokens(self, client):
        """GIVEN valid email and password
        WHEN login is called
        THEN it should return access token, refresh token, and user info
        """
        mock_response = {
            "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test_access",
            "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test_refresh",
            "token_type": "bearer",
            "user": {
                "id": "user_123",
                "email": "test@example.com",
                "name": "Test User",
                "email_verified": True
            }
        }

        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = MagicMock(
                status_code=200,
                json=lambda: mock_response
            )

            result = await client.login("test@example.com", "SecurePass123!")

            # Assertions
            assert result["access_token"] == mock_response["access_token"]
            assert result["refresh_token"] == mock_response["refresh_token"]
            assert result["token_type"] == "bearer"
            assert result["user"]["email"] == "test@example.com"
            assert result["user"]["id"] == "user_123"

            # Verify correct endpoint was called
            mock_post.assert_called_once()
            call_args = mock_post.call_args
            assert "auth/login" in str(call_args[0][0])

            # Verify correct payload was sent
            assert call_args[1]["json"]["email"] == "test@example.com"
            assert call_args[1]["json"]["password"] == "SecurePass123!"

    @pytest.mark.asyncio
    async def test_login_with_invalid_credentials_raises_401(self, client):
        """GIVEN invalid credentials
        WHEN login is called
        THEN it should raise HTTPException with 401 status
        """
        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = MagicMock(
                status_code=401,
                json=lambda: {"detail": "Incorrect email or password"}
            )

            with pytest.raises(HTTPException) as exc_info:
                await client.login("test@example.com", "wrong_password")

            assert exc_info.value.status_code == 401
            assert "Incorrect email or password" in str(exc_info.value.detail)

    @pytest.mark.asyncio
    async def test_login_with_network_error_raises_exception(self, client):
        """GIVEN network connectivity issues
        WHEN login is called
        THEN it should raise appropriate exception
        """
        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.side_effect = httpx.ConnectError("Connection failed")

            with pytest.raises(httpx.ConnectError):
                await client.login("test@example.com", "password123")


class TestAINativeClientRegister:
    """Test suite for user registration functionality."""

    @pytest.mark.asyncio
    async def test_register_with_valid_data_creates_user(self, client):
        """GIVEN valid registration data
        WHEN register is called
        THEN it should create user and return tokens
        """
        mock_response = {
            "access_token": "new_user_access_token",
            "refresh_token": "new_user_refresh_token",
            "token_type": "bearer",
            "user": {
                "id": "user_new_123",
                "email": "newuser@example.com",
                "name": "New User",
                "email_verified": False
            }
        }

        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = MagicMock(
                status_code=201,
                json=lambda: mock_response
            )

            result = await client.register(
                email="newuser@example.com",
                password="SecurePass123!",
                name="New User"
            )

            assert result["access_token"] == mock_response["access_token"]
            assert result["user"]["email"] == "newuser@example.com"
            assert result["user"]["name"] == "New User"

            # Verify endpoint
            call_args = mock_post.call_args
            assert "auth/register" in str(call_args[0][0])

    @pytest.mark.asyncio
    async def test_register_with_existing_email_raises_400(self, client):
        """GIVEN email already exists
        WHEN register is called
        THEN it should raise HTTPException with 400 status
        """
        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = MagicMock(
                status_code=400,
                json=lambda: {"detail": "Email already registered"}
            )

            with pytest.raises(HTTPException) as exc_info:
                await client.register(
                    email="existing@example.com",
                    password="password123",
                    name="Test"
                )

            assert exc_info.value.status_code == 400
            assert "Email already registered" in str(exc_info.value.detail)


class TestAINativeClientUserInfo:
    """Test suite for getting current user information."""

    @pytest.mark.asyncio
    async def test_get_current_user_with_valid_token_returns_user_info(self, client):
        """GIVEN a valid access token
        WHEN get_current_user is called
        THEN it should return user information
        """
        mock_response = {
            "id": "user_123",
            "email": "test@example.com",
            "name": "Test User",
            "email_verified": True,
            "created_at": "2025-01-01T00:00:00Z",
            "plan": "free"
        }

        with patch('httpx.AsyncClient.get', new_callable=AsyncMock) as mock_get:
            mock_get.return_value = MagicMock(
                status_code=200,
                json=lambda: mock_response
            )

            result = await client.get_current_user("valid_access_token")

            assert result["email"] == "test@example.com"
            assert result["id"] == "user_123"
            assert result["email_verified"] is True

            # Verify Authorization header was included
            call_args = mock_get.call_args
            assert "auth/me" in str(call_args[0][0])
            assert call_args[1]["headers"]["Authorization"] == "Bearer valid_access_token"

    @pytest.mark.asyncio
    async def test_get_current_user_with_invalid_token_raises_401(self, client):
        """GIVEN an invalid or expired token
        WHEN get_current_user is called
        THEN it should raise HTTPException with 401 status
        """
        with patch('httpx.AsyncClient.get', new_callable=AsyncMock) as mock_get:
            mock_get.return_value = MagicMock(
                status_code=401,
                json=lambda: {"detail": "Invalid or expired token"}
            )

            with pytest.raises(HTTPException) as exc_info:
                await client.get_current_user("invalid_token")

            assert exc_info.value.status_code == 401


class TestAINativeClientTokenRefresh:
    """Test suite for token refresh functionality."""

    @pytest.mark.asyncio
    async def test_refresh_token_with_valid_refresh_token_returns_new_access_token(self, client):
        """GIVEN a valid refresh token
        WHEN refresh_token is called
        THEN it should return new access token
        """
        mock_response = {
            "access_token": "new_refreshed_access_token",
            "token_type": "bearer"
        }

        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = MagicMock(
                status_code=200,
                json=lambda: mock_response
            )

            result = await client.refresh_token("valid_refresh_token")

            assert result["access_token"] == "new_refreshed_access_token"
            assert result["token_type"] == "bearer"

            # Verify endpoint
            call_args = mock_post.call_args
            assert "auth/refresh" in str(call_args[0][0])
            assert call_args[1]["json"]["refresh_token"] == "valid_refresh_token"

    @pytest.mark.asyncio
    async def test_refresh_token_with_expired_refresh_token_raises_401(self, client):
        """GIVEN an expired refresh token
        WHEN refresh_token is called
        THEN it should raise HTTPException with 401 status
        """
        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = MagicMock(
                status_code=401,
                json=lambda: {"detail": "Invalid or expired refresh token"}
            )

            with pytest.raises(HTTPException) as exc_info:
                await client.refresh_token("expired_refresh_token")

            assert exc_info.value.status_code == 401


class TestAINativeClientLogout:
    """Test suite for logout functionality."""

    @pytest.mark.asyncio
    async def test_logout_with_valid_token_blacklists_token(self, client):
        """GIVEN a valid access token
        WHEN logout is called
        THEN it should successfully blacklist the token
        """
        mock_response = {
            "message": "Successfully logged out"
        }

        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = MagicMock(
                status_code=200,
                json=lambda: mock_response
            )

            result = await client.logout("valid_access_token")

            assert result["message"] == "Successfully logged out"

            # Verify Authorization header
            call_args = mock_post.call_args
            assert "auth/logout" in str(call_args[0][0])
            assert call_args[1]["headers"]["Authorization"] == "Bearer valid_access_token"


class TestAINativeClientChatCompletion:
    """Test suite for chat completion functionality."""

    @pytest.mark.asyncio
    async def test_chat_completion_with_valid_request_returns_response(self, client):
        """GIVEN valid chat messages and token
        WHEN chat_completion is called
        THEN it should return completion response
        """
        messages = [
            {"role": "user", "content": "Hello, how are you?"}
        ]

        mock_response = {
            "id": "chatcmpl-abc123",
            "model": "llama-3.3-70b-instruct",
            "choices": [
                {
                    "message": {
                        "role": "assistant",
                        "content": ["I'm doing well, thank you! How can I help you today?"]
                    },
                    "finish_reason": "stop"
                }
            ],
            "usage": {
                "prompt_tokens": 20,
                "completion_tokens": 30,
                "total_tokens": 50
            }
        }

        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = MagicMock(
                status_code=200,
                json=lambda: mock_response
            )

            result = await client.chat_completion(
                messages=messages,
                model="llama-3.3-70b-instruct",
                token="valid_access_token"
            )

            assert result["id"] == "chatcmpl-abc123"
            assert result["choices"][0]["message"]["content"] == ["I'm doing well, thank you! How can I help you today?"]
            assert result["usage"]["total_tokens"] == 50

            # Verify request
            call_args = mock_post.call_args
            assert "chat/completions" in str(call_args[0][0])
            assert call_args[1]["headers"]["Authorization"] == "Bearer valid_access_token"
            assert call_args[1]["json"]["messages"] == messages
            assert call_args[1]["json"]["model"] == "llama-3.3-70b-instruct"

    @pytest.mark.asyncio
    async def test_chat_completion_with_insufficient_credits_raises_402(self, client):
        """GIVEN user has insufficient credits
        WHEN chat_completion is called
        THEN it should raise HTTPException with 402 status
        """
        messages = [{"role": "user", "content": "Hello"}]

        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = MagicMock(
                status_code=402,
                json=lambda: {"detail": "Insufficient credits"}
            )

            with pytest.raises(HTTPException) as exc_info:
                await client.chat_completion(
                    messages=messages,
                    model="llama-3.3-70b-instruct",
                    token="valid_token"
                )

            assert exc_info.value.status_code == 402
            assert "Insufficient credits" in str(exc_info.value.detail)

    @pytest.mark.asyncio
    async def test_chat_completion_with_unavailable_model_raises_403(self, client):
        """GIVEN model not available for user's plan
        WHEN chat_completion is called
        THEN it should raise HTTPException with 403 status
        """
        messages = [{"role": "user", "content": "Hello"}]

        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = MagicMock(
                status_code=403,
                json=lambda: {"detail": "Model not available for your plan"}
            )

            with pytest.raises(HTTPException) as exc_info:
                await client.chat_completion(
                    messages=messages,
                    model="gpt-4",
                    token="valid_token"
                )

            assert exc_info.value.status_code == 403

    @pytest.mark.asyncio
    async def test_chat_completion_with_custom_parameters(self, client):
        """GIVEN custom temperature and max_tokens parameters
        WHEN chat_completion is called
        THEN it should include parameters in request
        """
        messages = [{"role": "user", "content": "Test"}]

        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = MagicMock(
                status_code=200,
                json=lambda: {"id": "test", "choices": [], "usage": {}}
            )

            await client.chat_completion(
                messages=messages,
                model="llama-3.3-70b-instruct",
                token="token",
                temperature=0.9,
                max_tokens=2000
            )

            call_args = mock_post.call_args
            assert call_args[1]["json"]["temperature"] == 0.9
            assert call_args[1]["json"]["max_tokens"] == 2000


class TestAINativeClientConfiguration:
    """Test suite for client configuration and initialization."""

    def test_client_initialization_with_default_base_url(self):
        """GIVEN no base URL provided
        WHEN client is initialized
        THEN it should use default AINative API URL
        """
        from app.api.ainative_client import AINativeClient

        client = AINativeClient()

        assert client.base_url == "https://api.ainative.studio/v1"

    def test_client_initialization_with_custom_base_url(self):
        """GIVEN custom base URL provided
        WHEN client is initialized
        THEN it should use custom URL
        """
        from app.api.ainative_client import AINativeClient

        client = AINativeClient(base_url="https://custom.api.com/v1")

        assert client.base_url == "https://custom.api.com/v1"

    def test_client_has_appropriate_timeout(self):
        """GIVEN client is initialized
        WHEN checking timeout configuration
        THEN it should have reasonable timeout value
        """
        from app.api.ainative_client import AINativeClient

        client = AINativeClient()

        assert client.timeout > 0
        assert client.timeout <= 60  # Reasonable maximum timeout
