"""Application configuration management.

This module handles all configuration settings using pydantic-settings
for type-safe environment variable loading and validation.
"""
import secrets
from typing import List

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application settings loaded from environment variables.

    Attributes:
        API_VERSION: Current API version
        DEBUG: Enable debug mode
        SECRET_KEY: Secret key for cryptographic operations
        ALLOWED_ORIGINS: List of allowed CORS origins
        AINATIVE_API_BASE_URL: Base URL for AINative API
        AINATIVE_API_TIMEOUT: Request timeout for AINative API calls
    """

    API_VERSION: str = Field(default="v1", description="API version identifier")
    DEBUG: bool = Field(default=False, description="Enable debug mode")
    SECRET_KEY: str = Field(
        default_factory=lambda: secrets.token_urlsafe(32),
        description="Secret key for cryptographic operations",
    )
    ALLOWED_ORIGINS: List[str] = Field(
        default=["http://localhost:3000"],
        description="Allowed CORS origins",
    )

    # AINative API Integration Settings
    AINATIVE_API_BASE_URL: str = Field(
        default="https://api.ainative.studio/v1",
        description="Base URL for AINative API integration",
    )
    AINATIVE_API_TIMEOUT: float = Field(
        default=30.0,
        description="HTTP request timeout for AINative API calls (seconds)",
    )

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=True,
        extra="ignore",
    )


# Global settings instance
settings = Settings()
