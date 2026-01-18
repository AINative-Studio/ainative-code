"""AI Provider Schemas - Minimal implementation for testing."""
from pydantic import BaseModel, Field
from typing import Optional, Dict, Any, List
from enum import Enum
from datetime import datetime
from uuid import UUID


class AIProviderSchema(str, Enum):
    """Supported AI providers."""

    ANTHROPIC = "anthropic"
    OPENAI = "openai"
    COHERE = "cohere"


class AIRequestBase(BaseModel):
    """Base AI request schema."""

    prompt: Optional[str] = None
    messages: Optional[List[Dict[str, str]]] = None
    capability: str  # Will be linked to ModelCapabilityType
    temperature: Optional[float] = None
    max_tokens: Optional[int] = None
    additional_params: Optional[Dict[str, Any]] = None

    class Config:
        arbitrary_types_allowed = True


class AIUsage(BaseModel):
    """Token usage information."""

    prompt_tokens: int = 0
    completion_tokens: int = 0
    total_tokens: int = 0


class AIResponse(BaseModel):
    """AI response schema."""

    content: str
    model_id: UUID
    model_name: str
    provider: str
    usage: Dict[str, int]
    finish_reason: str = "stop"
    created_at: datetime = Field(default_factory=datetime.now)
    total_tokens: int = 0
    prompt_tokens: int = 0
    completion_tokens: int = 0
    is_cached: bool = False

    class Config:
        arbitrary_types_allowed = True
