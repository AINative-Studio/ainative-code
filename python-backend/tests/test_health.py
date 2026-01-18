"""Tests for health check endpoint.

Following BDD-style test naming:
GIVEN [initial context]
WHEN [action is performed]
THEN [expected outcome]
"""
import pytest
from fastapi.testclient import TestClient


def test_health_endpoint_returns_200(client):
    """GIVEN a FastAPI application
    WHEN the /health endpoint is called
    THEN it should return 200 OK
    """
    response = client.get("/health")
    assert response.status_code == 200


def test_health_endpoint_returns_json(client):
    """GIVEN a FastAPI application
    WHEN the /health endpoint is called
    THEN it should return JSON with status
    """
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert "status" in data
    assert data["status"] == "healthy"


def test_health_endpoint_includes_version(client):
    """GIVEN a FastAPI application
    WHEN the /health endpoint is called
    THEN it should include API version
    """
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert "version" in data
    assert isinstance(data["version"], str)
    assert len(data["version"]) > 0


def test_health_endpoint_returns_correct_content_type(client):
    """GIVEN a FastAPI application
    WHEN the /health endpoint is called
    THEN it should return JSON content type
    """
    response = client.get("/health")
    assert response.status_code == 200
    assert "application/json" in response.headers["content-type"]


def test_health_endpoint_response_structure(client):
    """GIVEN a FastAPI application
    WHEN the /health endpoint is called
    THEN it should return a well-formed response with all required fields
    """
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()

    # Verify all required fields are present
    required_fields = ["status", "version"]
    for field in required_fields:
        assert field in data, f"Missing required field: {field}"

    # Verify field types
    assert isinstance(data["status"], str)
    assert isinstance(data["version"], str)
