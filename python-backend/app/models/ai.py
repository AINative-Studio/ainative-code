"""AI Models - Minimal implementation for testing."""
from enum import Enum


class ModelCapabilityType(str, Enum):
    """Model capability types."""

    COMPLETION = "completion"
    CHAT_COMPLETION = "chat_completion"
    EMBEDDING = "embedding"
    IMAGE_GENERATION = "image_generation"
