"""Authentication request/response schemas.

This module defines Pydantic models for authentication-related
API requests and responses when integrating with AINative API.
"""
from typing import Optional
from pydantic import BaseModel, Field, EmailStr


class LoginRequest(BaseModel):
    """Login request schema.

    Attributes:
        email: User email address
        password: User password
    """

    email: EmailStr = Field(
        ...,
        description="User email address",
        examples=["user@example.com"]
    )
    password: str = Field(
        ...,
        min_length=8,
        description="User password (minimum 8 characters)",
        examples=["SecurePassword123!"]
    )


class RegisterRequest(BaseModel):
    """User registration request schema.

    Attributes:
        email: User email address
        password: User password
        name: User display name
    """

    email: EmailStr = Field(
        ...,
        description="User email address",
        examples=["newuser@example.com"]
    )
    password: str = Field(
        ...,
        min_length=8,
        description="User password (minimum 8 characters)",
        examples=["SecurePassword123!"]
    )
    name: str = Field(
        ...,
        min_length=1,
        max_length=100,
        description="User display name",
        examples=["John Doe"]
    )


class UserInfo(BaseModel):
    """User information schema.

    Attributes:
        id: Unique user identifier
        email: User email address
        name: User display name
        email_verified: Whether email is verified
    """

    id: str = Field(..., description="Unique user identifier")
    email: str = Field(..., description="User email address")
    name: str = Field(..., description="User display name")
    email_verified: bool = Field(
        default=False,
        description="Email verification status"
    )


class TokenResponse(BaseModel):
    """Authentication token response schema.

    Attributes:
        access_token: JWT access token
        refresh_token: JWT refresh token (optional)
        token_type: Token type (typically 'bearer')
        user: User information
    """

    access_token: str = Field(..., description="JWT access token")
    refresh_token: Optional[str] = Field(
        None,
        description="JWT refresh token"
    )
    token_type: str = Field(
        default="bearer",
        description="Token type"
    )
    user: UserInfo = Field(..., description="User information")


class RefreshTokenRequest(BaseModel):
    """Token refresh request schema.

    Attributes:
        refresh_token: Valid refresh token
    """

    refresh_token: str = Field(
        ...,
        description="Valid refresh token",
        examples=["eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."]
    )


class RefreshTokenResponse(BaseModel):
    """Token refresh response schema.

    Attributes:
        access_token: New JWT access token
        token_type: Token type (typically 'bearer')
    """

    access_token: str = Field(..., description="New JWT access token")
    token_type: str = Field(
        default="bearer",
        description="Token type"
    )


class LogoutResponse(BaseModel):
    """Logout response schema.

    Attributes:
        message: Success message
    """

    message: str = Field(
        default="Successfully logged out",
        description="Logout success message"
    )


class ChatMessage(BaseModel):
    """Chat message schema.

    Attributes:
        role: Message role (user, assistant, system)
        content: Message content
    """

    role: str = Field(
        ...,
        description="Message role",
        pattern="^(user|assistant|system)$",
        examples=["user"]
    )
    content: str = Field(
        ...,
        description="Message content",
        examples=["Hello, how can you help me?"]
    )


class ChatCompletionRequest(BaseModel):
    """Chat completion request schema.

    Attributes:
        messages: List of chat messages
        model: Model identifier
        temperature: Sampling temperature (0.0-1.0)
        max_tokens: Maximum tokens to generate
        stream: Whether to stream the response
    """

    messages: list[ChatMessage] = Field(
        ...,
        description="List of chat messages",
        min_length=1
    )
    model: str = Field(
        default="llama-3.3-70b-instruct",
        description="Model identifier",
        examples=["llama-3.3-70b-instruct", "claude-3-5-sonnet-20241022"]
    )
    temperature: float = Field(
        default=0.7,
        ge=0.0,
        le=1.0,
        description="Sampling temperature"
    )
    max_tokens: int = Field(
        default=1000,
        ge=1,
        le=4000,
        description="Maximum tokens to generate"
    )
    stream: bool = Field(
        default=False,
        description="Stream response"
    )


class ChatCompletionChoice(BaseModel):
    """Chat completion choice schema.

    Attributes:
        message: Generated message
        finish_reason: Reason for completion finish
    """

    message: dict = Field(..., description="Generated message")
    finish_reason: str = Field(..., description="Finish reason")


class ChatCompletionUsage(BaseModel):
    """Chat completion usage statistics.

    Attributes:
        prompt_tokens: Number of tokens in prompt
        completion_tokens: Number of tokens in completion
        total_tokens: Total tokens used
    """

    prompt_tokens: Optional[int] = Field(None, description="Prompt tokens")
    completion_tokens: Optional[int] = Field(None, description="Completion tokens")
    total_tokens: int = Field(..., description="Total tokens")


class ChatCompletionResponse(BaseModel):
    """Chat completion response schema.

    Attributes:
        id: Completion identifier
        model: Model used
        choices: List of completion choices
        usage: Token usage statistics
    """

    id: str = Field(..., description="Completion identifier")
    model: str = Field(..., description="Model used")
    choices: list[ChatCompletionChoice] = Field(
        ...,
        description="Completion choices"
    )
    usage: ChatCompletionUsage = Field(..., description="Usage statistics")
