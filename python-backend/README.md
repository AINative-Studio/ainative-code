# AINative Code Backend

A FastAPI-based backend microservice for AINative Code platform, providing authentication and chat completion services.

## Features

- FastAPI framework with automatic OpenAPI documentation
- Pydantic for data validation and settings management
- CORS middleware for cross-origin requests
- Health check endpoint
- Comprehensive test suite with 83% coverage
- Type hints throughout the codebase

## Project Structure

```
python-backend/
├── app/
│   ├── __init__.py
│   ├── main.py                 # FastAPI application entry point
│   ├── core/
│   │   ├── __init__.py
│   │   ├── config.py           # Configuration management
│   │   └── logging.py          # Logging setup (placeholder)
│   ├── api/
│   │   ├── __init__.py
│   │   └── v1/                 # API version 1 endpoints
│   │       └── __init__.py
│   └── schemas/
│       └── __init__.py         # Pydantic schemas
├── tests/
│   ├── __init__.py
│   ├── conftest.py             # Pytest configuration and fixtures
│   ├── test_config.py          # Configuration tests
│   ├── test_health.py          # Health endpoint tests
│   └── test_main.py            # Main application tests
├── venv/                       # Virtual environment (not in git)
├── .env.example                # Example environment variables
├── .gitignore                  # Git ignore rules
├── pytest.ini                  # Pytest configuration
├── pyproject.toml              # Poetry configuration
├── requirements.txt            # Python dependencies
└── README.md                   # This file
```

## Prerequisites

- Python 3.11 or higher
- pip (Python package installer)

## Installation

### 1. Clone the repository

```bash
cd /Users/aideveloper/AINative-Code/python-backend
```

### 2. Create a virtual environment

```bash
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

### 3. Install dependencies

```bash
pip install --upgrade pip
pip install -r requirements.txt
```

### 4. Configure environment variables

```bash
cp .env.example .env
# Edit .env and update values as needed
```

Generate a secure SECRET_KEY in Python:

```python
import secrets
print(secrets.token_urlsafe(32))
```

## Running the Application

### Development Mode

```bash
# Activate virtual environment
source venv/bin/activate

# Run with auto-reload
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

### Production Mode

```bash
# Activate virtual environment
source venv/bin/activate

# Run with production settings
uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 4
```

The API will be available at:
- API: http://localhost:8000
- Interactive docs (Swagger UI): http://localhost:8000/docs
- Alternative docs (ReDoc): http://localhost:8000/redoc
- OpenAPI schema: http://localhost:8000/openapi.json

## API Endpoints

### Health Check

```bash
curl http://localhost:8000/health
```

Response:
```json
{
  "status": "healthy",
  "version": "v1"
}
```

## Running Tests

### Run all tests

```bash
source venv/bin/activate
pytest
```

### Run with coverage report

```bash
pytest --cov=app.core --cov=app.main --cov-report=term-missing
```

### Run specific test file

```bash
pytest tests/test_health.py -v
```

### Run tests with detailed output

```bash
pytest -v --tb=short
```

### View HTML coverage report

```bash
pytest --cov=app.core --cov=app.main --cov-report=html
open htmlcov/index.html
```

## Test-Driven Development (TDD)

This project was built following strict TDD methodology:

1. **RED Phase**: Wrote failing tests first
   - `tests/test_config.py` - Configuration tests
   - `tests/test_health.py` - Health endpoint tests
   - `tests/test_main.py` - Application tests

2. **GREEN Phase**: Implemented minimal code to pass tests
   - `app/core/config.py` - Configuration management
   - `app/main.py` - FastAPI application

3. **REFACTOR Phase**: Code is clean and maintainable
   - 83% test coverage (exceeds 80% requirement)
   - All 20 tests passing
   - Type hints throughout
   - Clear documentation

## Development Workflow

### Adding New Features (TDD Approach)

1. **Write tests first** (RED phase):
```bash
# Create test file
touch tests/test_new_feature.py

# Write BDD-style tests
# GIVEN [context]
# WHEN [action]
# THEN [expected outcome]
```

2. **Run tests to ensure they fail**:
```bash
pytest tests/test_new_feature.py -v
```

3. **Implement feature** (GREEN phase):
```bash
# Create feature module
touch app/api/v1/new_feature.py

# Implement minimal code to pass tests
```

4. **Run tests to ensure they pass**:
```bash
pytest tests/test_new_feature.py -v
```

5. **Refactor and verify coverage**:
```bash
pytest --cov=app --cov-fail-under=80
```

### Code Quality

Format code with Black:
```bash
black app/ tests/
```

Lint code with Ruff:
```bash
ruff check app/ tests/
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `API_VERSION` | API version identifier | `v1` |
| `DEBUG` | Enable debug mode | `false` |
| `SECRET_KEY` | Secret key for cryptographic operations | Auto-generated |
| `ALLOWED_ORIGINS` | Allowed CORS origins (JSON array) | `["http://localhost:3000"]` |

### Configuration File

Configuration is managed in `app/core/config.py` using Pydantic Settings:

- Type-safe configuration
- Environment variable loading
- Default values
- Validation

## Testing

### Test Structure

Tests follow BDD (Behavior-Driven Development) style:

```python
def test_feature_name():
    """GIVEN initial context
    WHEN action is performed
    THEN expected outcome occurs
    """
    # Arrange
    # Act
    # Assert
```

### Test Coverage

Current coverage: **83%**

Coverage breakdown:
- `app/core/config.py`: 100%
- `app/main.py`: 82%

### Fixtures

Common test fixtures in `tests/conftest.py`:
- `client`: TestClient for making HTTP requests
- `mock_settings`: Mock configuration settings

## Deployment

### Docker (Future)

```dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app/ ./app/

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

## Troubleshooting

### Import Errors

If you encounter import errors, ensure:
1. Virtual environment is activated
2. All dependencies are installed: `pip install -r requirements.txt`
3. You're running commands from the project root

### Test Failures

If tests fail:
1. Check virtual environment is activated
2. Verify dependencies are up to date
3. Check environment variables are set correctly
4. Review test output for specific errors

### Coverage Warnings

Coverage warnings about "module-not-measured" can be ignored if:
- They relate to test setup/teardown
- Coverage percentage still meets requirements (80%+)

## Contributing

1. Follow TDD methodology (write tests first)
2. Maintain test coverage above 80%
3. Use type hints
4. Follow PEP 8 style guide (enforced by Black and Ruff)
5. Write clear docstrings and comments

## License

Proprietary - AINative Studio

## Support

For questions or issues, contact the AINative development team.

---

**Built with FastAPI and Test-Driven Development**
