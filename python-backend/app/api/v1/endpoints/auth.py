"""Authentication endpoints using AINative API integration.

This module provides FastAPI endpoints that integrate with the AINative API
for user authentication, registration, and token management.
"""
from fastapi import APIRouter, Depends, HTTPException, Header
from typing import Optional
import logging

from app.api.ainative_client import AINativeClient
from app.schemas.auth import (
    LoginRequest,
    RegisterRequest,
    TokenResponse,
    RefreshTokenRequest,
    RefreshTokenResponse,
    LogoutResponse,
    UserInfo,
    ChatCompletionRequest,
    ChatCompletionResponse,
)
from app.core.config import settings

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/auth", tags=["Authentication"])


def get_ainative_client() -> AINativeClient:
    """Dependency to get AINative API client instance.

    Returns:
        AINativeClient: Configured client instance
    """
    return AINativeClient(
        base_url=settings.AINATIVE_API_BASE_URL,
        timeout=settings.AINATIVE_API_TIMEOUT
    )


async def get_token_from_header(
    authorization: Optional[str] = Header(None)
) -> str:
    """Extract and validate bearer token from Authorization header.

    Args:
        authorization: Authorization header value

    Returns:
        str: Extracted token

    Raises:
        HTTPException: If token is missing or invalid format
    """
    if not authorization:
        logger.warning("Missing Authorization header")
        raise HTTPException(
            status_code=401,
            detail="Missing Authorization header"
        )

    parts = authorization.split()
    if len(parts) != 2 or parts[0].lower() != "bearer":
        logger.warning(f"Invalid Authorization header format: {authorization}")
        raise HTTPException(
            status_code=401,
            detail="Invalid Authorization header format. Expected: Bearer <token>"
        )

    return parts[1]


@router.post("/login", response_model=TokenResponse)
async def login(
    request: LoginRequest,
    client: AINativeClient = Depends(get_ainative_client)
):
    """Login via AINative API.

    Args:
        request: Login credentials (email and password)
        client: AINative API client instance

    Returns:
        TokenResponse: Access token, refresh token, and user info

    Raises:
        HTTPException: If login fails (401 for invalid credentials)
    """
    try:
        logger.info(f"Login attempt for user: {request.email}")
        result = await client.login(request.email, request.password)
        logger.info(f"Login successful for user: {request.email}")
        return result
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Unexpected error during login: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Internal server error: {str(e)}"
        )


@router.post("/register", response_model=TokenResponse, status_code=201)
async def register(
    request: RegisterRequest,
    client: AINativeClient = Depends(get_ainative_client)
):
    """Register new user via AINative API.

    Args:
        request: Registration data (email, password, name)
        client: AINative API client instance

    Returns:
        TokenResponse: Access token, refresh token, and user info

    Raises:
        HTTPException: If registration fails (400 for validation errors)
    """
    try:
        logger.info(f"Registration attempt for user: {request.email}")
        result = await client.register(
            email=request.email,
            password=request.password,
            name=request.name
        )
        logger.info(f"Registration successful for user: {request.email}")
        return result
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Unexpected error during registration: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Internal server error: {str(e)}"
        )


@router.get("/me", response_model=UserInfo)
async def get_current_user(
    token: str = Depends(get_token_from_header),
    client: AINativeClient = Depends(get_ainative_client)
):
    """Get current user information from AINative API.

    Args:
        token: Valid access token from Authorization header
        client: AINative API client instance

    Returns:
        UserInfo: Current user information

    Raises:
        HTTPException: If token is invalid or expired (401)
    """
    try:
        logger.info("Fetching current user information")
        result = await client.get_current_user(token)
        logger.info(f"Successfully retrieved user info for: {result.get('email')}")
        return result
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Unexpected error getting user info: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Internal server error: {str(e)}"
        )


@router.post("/refresh", response_model=RefreshTokenResponse)
async def refresh_token(
    request: RefreshTokenRequest,
    client: AINativeClient = Depends(get_ainative_client)
):
    """Refresh access token using refresh token.

    Args:
        request: Refresh token request
        client: AINative API client instance

    Returns:
        RefreshTokenResponse: New access token

    Raises:
        HTTPException: If refresh token is invalid or expired (401)
    """
    try:
        logger.info("Refreshing access token")
        result = await client.refresh_token(request.refresh_token)
        logger.info("Successfully refreshed access token")
        return result
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Unexpected error refreshing token: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Internal server error: {str(e)}"
        )


@router.post("/logout", response_model=LogoutResponse)
async def logout(
    token: str = Depends(get_token_from_header),
    client: AINativeClient = Depends(get_ainative_client)
):
    """Logout and blacklist current token.

    Args:
        token: Valid access token from Authorization header
        client: AINative API client instance

    Returns:
        LogoutResponse: Success message

    Raises:
        HTTPException: If logout fails
    """
    try:
        logger.info("User logout attempt")
        result = await client.logout(token)
        logger.info("Successfully logged out user")
        return result
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Unexpected error during logout: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Internal server error: {str(e)}"
        )


# Chat completion endpoint
chat_router = APIRouter(prefix="/chat", tags=["Chat"])


@chat_router.post("/completions", response_model=ChatCompletionResponse)
async def chat_completion(
    request: ChatCompletionRequest,
    token: str = Depends(get_token_from_header),
    client: AINativeClient = Depends(get_ainative_client)
):
    """Get chat completion from AINative API.

    Args:
        request: Chat completion request with messages and parameters
        token: Valid access token from Authorization header
        client: AINative API client instance

    Returns:
        ChatCompletionResponse: Chat completion with choices and usage

    Raises:
        HTTPException:
            - 401 if token is invalid
            - 402 if insufficient credits
            - 403 if model not available for plan
    """
    try:
        logger.info(f"Chat completion request for model: {request.model}")
        result = await client.chat_completion(
            messages=[msg.model_dump() for msg in request.messages],
            model=request.model,
            token=token,
            temperature=request.temperature,
            max_tokens=request.max_tokens,
            stream=request.stream
        )
        logger.info(f"Chat completion successful for model: {request.model}")
        return result
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Unexpected error during chat completion: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Internal server error: {str(e)}"
        )
