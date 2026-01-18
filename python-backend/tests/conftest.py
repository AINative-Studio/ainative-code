"""Pytest configuration and shared fixtures."""
import pytest
from fastapi.testclient import TestClient


@pytest.fixture
def client():
    """Create a test client for the FastAPI application.

    This fixture will be used across all tests that need to make
    HTTP requests to the API endpoints.
    """
    from app.main import app
    return TestClient(app)


@pytest.fixture
def mock_settings(monkeypatch):
    """Create mock settings for testing.

    This fixture allows tests to override configuration settings
    without affecting other tests.
    """
    def _set_env(**kwargs):
        for key, value in kwargs.items():
            monkeypatch.setenv(key, str(value))
    return _set_env
