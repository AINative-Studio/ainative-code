"""AINative API HTTP Client.

This module provides an HTTP client for integrating with the AINative API
at https://api.ainative.studio/v1. It handles authentication, user management,
and chat completion requests.

All methods are designed to work with the existing AINative API endpoints,
NOT to replicate authentication logic.
"""
import httpx
from typing import Dict, List, Optional
from fastapi import HTTPException
import logging

logger = logging.getLogger(__name__)


class AINativeClient:
    """HTTP client for AINative API integration.

    This client provides methods to interact with the AINative API endpoints
    for authentication, user management, and chat completions.

    Attributes:
        base_url: Base URL for the AINative API
        timeout: Request timeout in seconds
    """

    def __init__(
        self,
        base_url: str = "https://api.ainative.studio/v1",
        timeout: float = 30.0
    ):
        """Initialize the AINative API client.

        Args:
            base_url: Base URL for the AINative API (default: production URL)
            timeout: Request timeout in seconds (default: 30.0)
        """
        self.base_url = base_url
        self.timeout = timeout
        logger.info(f"AINative client initialized with base_url: {base_url}")

    async def login(self, email: str, password: str) -> Dict:
        """Login to AINative API.

        Args:
            email: User email address
            password: User password

        Returns:
            Dict containing access_token, refresh_token, token_type, and user info

        Raises:
            HTTPException: If login fails (401 for invalid credentials)
        """
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            try:
                response = await client.post(
                    f"{self.base_url}/auth/login",
                    json={"email": email, "password": password}
                )

                if response.status_code == 401:
                    error_detail = response.json().get("detail", "Incorrect email or password")
                    logger.warning(f"Login failed for {email}: {error_detail}")
                    raise HTTPException(
                        status_code=401,
                        detail=error_detail
                    )

                response.raise_for_status()
                logger.info(f"Successfully logged in user: {email}")
                return response.json()

            except httpx.HTTPStatusError as e:
                logger.error(f"HTTP error during login: {e}")
                raise HTTPException(
                    status_code=e.response.status_code,
                    detail=str(e)
                )

    async def register(
        self,
        email: str,
        password: str,
        name: str
    ) -> Dict:
        """Register new user with AINative API.

        Args:
            email: User email address
            password: User password
            name: User display name

        Returns:
            Dict containing access_token, refresh_token, token_type, and user info

        Raises:
            HTTPException: If registration fails (400 for validation errors)
        """
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            try:
                response = await client.post(
                    f"{self.base_url}/auth/register",
                    json={
                        "email": email,
                        "password": password,
                        "name": name
                    }
                )

                if response.status_code == 400:
                    error_detail = response.json().get("detail", "Registration failed")
                    logger.warning(f"Registration failed for {email}: {error_detail}")
                    raise HTTPException(
                        status_code=400,
                        detail=error_detail
                    )

                response.raise_for_status()
                logger.info(f"Successfully registered user: {email}")
                return response.json()

            except httpx.HTTPStatusError as e:
                logger.error(f"HTTP error during registration: {e}")
                raise HTTPException(
                    status_code=e.response.status_code,
                    detail=str(e)
                )

    async def get_current_user(self, token: str) -> Dict:
        """Get current user information from AINative API.

        Args:
            token: Valid access token

        Returns:
            Dict containing user information (id, email, name, etc.)

        Raises:
            HTTPException: If token is invalid or expired (401)
        """
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            try:
                response = await client.get(
                    f"{self.base_url}/auth/me",
                    headers={"Authorization": f"Bearer {token}"}
                )

                if response.status_code == 401:
                    error_detail = response.json().get("detail", "Invalid or expired token")
                    logger.warning(f"Failed to get user info: {error_detail}")
                    raise HTTPException(
                        status_code=401,
                        detail=error_detail
                    )

                response.raise_for_status()
                return response.json()

            except httpx.HTTPStatusError as e:
                logger.error(f"HTTP error getting user info: {e}")
                raise HTTPException(
                    status_code=e.response.status_code,
                    detail=str(e)
                )

    async def refresh_token(self, refresh_token: str) -> Dict:
        """Refresh access token using refresh token.

        Args:
            refresh_token: Valid refresh token

        Returns:
            Dict containing new access_token and token_type

        Raises:
            HTTPException: If refresh token is invalid or expired (401)
        """
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            try:
                response = await client.post(
                    f"{self.base_url}/auth/refresh",
                    json={"refresh_token": refresh_token}
                )

                if response.status_code == 401:
                    error_detail = response.json().get("detail", "Invalid or expired refresh token")
                    logger.warning(f"Failed to refresh token: {error_detail}")
                    raise HTTPException(
                        status_code=401,
                        detail=error_detail
                    )

                response.raise_for_status()
                logger.info("Successfully refreshed access token")
                return response.json()

            except httpx.HTTPStatusError as e:
                logger.error(f"HTTP error refreshing token: {e}")
                raise HTTPException(
                    status_code=e.response.status_code,
                    detail=str(e)
                )

    async def logout(self, token: str) -> Dict:
        """Logout and blacklist token.

        Args:
            token: Valid access token to blacklist

        Returns:
            Dict with success message

        Raises:
            HTTPException: If logout fails
        """
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            try:
                response = await client.post(
                    f"{self.base_url}/auth/logout",
                    headers={"Authorization": f"Bearer {token}"}
                )

                response.raise_for_status()
                logger.info("Successfully logged out user")
                return response.json()

            except httpx.HTTPStatusError as e:
                logger.error(f"HTTP error during logout: {e}")
                raise HTTPException(
                    status_code=e.response.status_code,
                    detail=str(e)
                )

    async def chat_completion(
        self,
        messages: List[Dict],
        model: str,
        token: str,
        temperature: float = 0.7,
        max_tokens: int = 1000,
        stream: bool = False
    ) -> Dict:
        """Get chat completion from AINative API.

        Args:
            messages: List of message dictionaries with 'role' and 'content'
            model: Model identifier (e.g., 'llama-3.3-70b-instruct')
            token: Valid access token
            temperature: Sampling temperature (0.0-1.0)
            max_tokens: Maximum tokens to generate
            stream: Whether to stream the response

        Returns:
            Dict containing completion response with choices and usage

        Raises:
            HTTPException:
                - 402 if insufficient credits
                - 403 if model not available for plan
                - 401 if token invalid
        """
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            try:
                response = await client.post(
                    f"{self.base_url}/chat/completions",
                    headers={"Authorization": f"Bearer {token}"},
                    json={
                        "messages": messages,
                        "model": model,
                        "temperature": temperature,
                        "max_tokens": max_tokens,
                        "stream": stream
                    }
                )

                if response.status_code == 402:
                    logger.warning("Chat completion failed: Insufficient credits")
                    raise HTTPException(
                        status_code=402,
                        detail="Insufficient credits"
                    )

                if response.status_code == 403:
                    logger.warning(f"Chat completion failed: Model {model} not available")
                    raise HTTPException(
                        status_code=403,
                        detail="Model not available for your plan"
                    )

                if response.status_code == 401:
                    logger.warning("Chat completion failed: Invalid token")
                    raise HTTPException(
                        status_code=401,
                        detail="Invalid or expired token"
                    )

                response.raise_for_status()
                logger.info(f"Successfully generated chat completion with model: {model}")
                return response.json()

            except httpx.HTTPStatusError as e:
                logger.error(f"HTTP error during chat completion: {e}")
                raise HTTPException(
                    status_code=e.response.status_code,
                    detail=str(e)
                )
