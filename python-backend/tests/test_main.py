"""Tests for main FastAPI application.

Following BDD-style test naming:
GIVEN [initial context]
WHEN [action is performed]
THEN [expected outcome]
"""
import pytest


def test_app_is_fastapi_instance(client):
    """GIVEN a FastAPI application
    WHEN the app is instantiated
    THEN it should be a valid FastAPI instance
    """
    from app.main import app
    from fastapi import FastAPI

    assert isinstance(app, FastAPI)


def test_app_has_title(client):
    """GIVEN a FastAPI application
    WHEN the app is instantiated
    THEN it should have a title
    """
    from app.main import app

    assert hasattr(app, "title")
    assert app.title is not None
    assert len(app.title) > 0


def test_app_has_version(client):
    """GIVEN a FastAPI application
    WHEN the app is instantiated
    THEN it should have a version
    """
    from app.main import app

    assert hasattr(app, "version")
    assert app.version is not None


def test_app_has_cors_middleware(client):
    """GIVEN a FastAPI application
    WHEN the app is instantiated
    THEN it should have CORS middleware configured
    """
    # Test CORS behavior by making a request with an allowed origin
    response = client.get("/health", headers={
        "Origin": "http://localhost:3000"
    })

    # CORS middleware should add the access-control-allow-origin header
    assert response.status_code == 200
    assert "access-control-allow-origin" in response.headers


def test_root_endpoint_not_found():
    """GIVEN a FastAPI application
    WHEN the root endpoint is called
    THEN it should return 404 (no root endpoint defined yet)
    """
    from app.main import app
    from fastapi.testclient import TestClient

    client = TestClient(app)
    response = client.get("/")
    assert response.status_code == 404


def test_app_openapi_docs_available(client):
    """GIVEN a FastAPI application
    WHEN the /docs endpoint is called
    THEN it should return the OpenAPI documentation
    """
    response = client.get("/docs")
    assert response.status_code == 200


def test_app_openapi_json_available(client):
    """GIVEN a FastAPI application
    WHEN the /openapi.json endpoint is called
    THEN it should return the OpenAPI schema
    """
    response = client.get("/openapi.json")
    assert response.status_code == 200
    schema = response.json()
    assert "openapi" in schema
    assert "info" in schema
    assert "paths" in schema
