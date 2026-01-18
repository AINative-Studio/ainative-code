"""Tests for configuration management.

Following BDD-style test naming:
GIVEN [initial context]
WHEN [action is performed]
THEN [expected outcome]
"""
import pytest
import os


def test_settings_has_default_api_version():
    """GIVEN no environment variables set
    WHEN Settings is instantiated
    THEN it should have a default API_VERSION
    """
    from app.core.config import Settings

    settings = Settings()
    assert hasattr(settings, "API_VERSION")
    assert settings.API_VERSION is not None
    assert isinstance(settings.API_VERSION, str)


def test_settings_has_default_debug():
    """GIVEN no environment variables set
    WHEN Settings is instantiated
    THEN it should have a default DEBUG setting
    """
    from app.core.config import Settings

    settings = Settings()
    assert hasattr(settings, "DEBUG")
    assert isinstance(settings.DEBUG, bool)


def test_settings_has_secret_key():
    """GIVEN no environment variables set
    WHEN Settings is instantiated
    THEN it should have a SECRET_KEY (default or from env)
    """
    from app.core.config import Settings

    settings = Settings()
    assert hasattr(settings, "SECRET_KEY")
    assert settings.SECRET_KEY is not None
    assert len(settings.SECRET_KEY) > 0


def test_settings_has_allowed_origins():
    """GIVEN no environment variables set
    WHEN Settings is instantiated
    THEN it should have default ALLOWED_ORIGINS
    """
    from app.core.config import Settings

    settings = Settings()
    assert hasattr(settings, "ALLOWED_ORIGINS")
    assert isinstance(settings.ALLOWED_ORIGINS, list)
    assert len(settings.ALLOWED_ORIGINS) > 0


def test_settings_loads_from_env(monkeypatch):
    """GIVEN environment variables set
    WHEN Settings is instantiated
    THEN it should load configuration from environment
    """
    from app.core.config import Settings

    monkeypatch.setenv("API_VERSION", "v2")
    monkeypatch.setenv("DEBUG", "true")

    settings = Settings()
    assert settings.API_VERSION == "v2"
    assert settings.DEBUG is True


def test_settings_debug_false_by_default():
    """GIVEN no DEBUG environment variable
    WHEN Settings is instantiated
    THEN DEBUG should default to False
    """
    from app.core.config import Settings

    # Ensure DEBUG is not set
    if "DEBUG" in os.environ:
        del os.environ["DEBUG"]

    settings = Settings()
    assert settings.DEBUG is False


def test_settings_secret_key_is_secure():
    """GIVEN Settings instantiated
    WHEN SECRET_KEY is accessed
    THEN it should be a secure random string of adequate length
    """
    from app.core.config import Settings

    settings = Settings()
    # Should be at least 32 characters for security
    assert len(settings.SECRET_KEY) >= 32


def test_settings_singleton_behavior():
    """GIVEN a settings module
    WHEN settings are imported multiple times
    THEN it should provide consistent configuration
    """
    from app.core.config import settings as settings1
    from app.core.config import settings as settings2

    assert settings1.API_VERSION == settings2.API_VERSION
    assert settings1.SECRET_KEY == settings2.SECRET_KEY
