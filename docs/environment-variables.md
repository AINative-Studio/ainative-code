# Environment Variables Reference

This document provides a comprehensive reference for all environment variables supported by AINative Code.

## Table of Contents

1. [Overview](#overview)
2. [Naming Convention](#naming-convention)
3. [Configuration Precedence](#configuration-precedence)
4. [Application Settings](#application-settings)
5. [LLM Provider Variables](#llm-provider-variables)
6. [Platform Authentication](#platform-authentication)
7. [Service Endpoints](#service-endpoints)
8. [Tool Configuration](#tool-configuration)
9. [Performance Settings](#performance-settings)
10. [Logging Configuration](#logging-configuration)
11. [Security Settings](#security-settings)
12. [Complete Variable Reference](#complete-variable-reference)

## Overview

AINative Code supports configuration through environment variables using the `AINATIVE_` prefix. This allows for flexible deployment across different environments without modifying configuration files.

## Naming Convention

Environment variables follow this pattern:

```
AINATIVE_<SECTION>_<SUBSECTION>_<KEY>
```

Where:
- All uppercase letters
- Sections separated by underscores
- Nested configuration uses additional underscores
- Dots and dashes in YAML keys become underscores

**Examples:**
- YAML: `app.environment` → ENV: `AINATIVE_APP_ENVIRONMENT`
- YAML: `llm.anthropic.api_key` → ENV: `AINATIVE_LLM_ANTHROPIC_API_KEY`
- YAML: `services.zerodb.max_connections` → ENV: `AINATIVE_SERVICES_ZERODB_MAX_CONNECTIONS`

## Configuration Precedence

Configuration is loaded in the following order (highest to lowest priority):

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration file** (config.yaml)
4. **Default values** (lowest priority)

This means environment variables will override configuration file settings, but command-line flags override everything.

## Application Settings

### AINATIVE_APP_NAME
- **Type:** String
- **Default:** `ainative-code`
- **Description:** Application name
- **Example:** `AINATIVE_APP_NAME=my-custom-app`

### AINATIVE_APP_VERSION
- **Type:** String
- **Default:** `0.1.0`
- **Description:** Application version
- **Example:** `AINATIVE_APP_VERSION=1.0.0`

### AINATIVE_APP_ENVIRONMENT
- **Type:** String
- **Default:** `development`
- **Valid Values:** `development`, `staging`, `production`
- **Description:** Runtime environment
- **Example:** `AINATIVE_APP_ENVIRONMENT=production`

### AINATIVE_APP_DEBUG
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable debug mode
- **Example:** `AINATIVE_APP_DEBUG=true`

## LLM Provider Variables

### Default Provider

#### AINATIVE_LLM_DEFAULT_PROVIDER
- **Type:** String
- **Default:** `anthropic`
- **Valid Values:** `anthropic`, `openai`, `google`, `bedrock`, `azure`, `ollama`
- **Description:** Primary LLM provider to use
- **Example:** `AINATIVE_LLM_DEFAULT_PROVIDER=openai`

### Anthropic Claude

#### AINATIVE_LLM_ANTHROPIC_API_KEY
- **Type:** String
- **Required:** Yes (if using Anthropic)
- **Description:** Anthropic API key
- **Example:** `AINATIVE_LLM_ANTHROPIC_API_KEY=sk-ant-api03-...`
- **Get Key:** https://console.anthropic.com/

#### AINATIVE_LLM_ANTHROPIC_MODEL
- **Type:** String
- **Default:** `claude-3-5-sonnet-20241022`
- **Description:** Claude model to use
- **Example:** `AINATIVE_LLM_ANTHROPIC_MODEL=claude-3-opus-20240229`

#### AINATIVE_LLM_ANTHROPIC_MAX_TOKENS
- **Type:** Integer
- **Default:** `4096`
- **Description:** Maximum tokens in response
- **Example:** `AINATIVE_LLM_ANTHROPIC_MAX_TOKENS=8192`

#### AINATIVE_LLM_ANTHROPIC_TEMPERATURE
- **Type:** Float
- **Default:** `0.7`
- **Range:** `0.0` - `1.0`
- **Description:** Sampling temperature (higher = more random)
- **Example:** `AINATIVE_LLM_ANTHROPIC_TEMPERATURE=0.5`

#### AINATIVE_LLM_ANTHROPIC_TOP_P
- **Type:** Float
- **Default:** `1.0`
- **Range:** `0.0` - `1.0`
- **Description:** Nucleus sampling parameter
- **Example:** `AINATIVE_LLM_ANTHROPIC_TOP_P=0.9`

#### AINATIVE_LLM_ANTHROPIC_TOP_K
- **Type:** Integer
- **Default:** `0`
- **Description:** Top-K sampling parameter
- **Example:** `AINATIVE_LLM_ANTHROPIC_TOP_K=40`

#### AINATIVE_LLM_ANTHROPIC_TIMEOUT
- **Type:** Duration
- **Default:** `30s`
- **Description:** Request timeout
- **Example:** `AINATIVE_LLM_ANTHROPIC_TIMEOUT=60s`

#### AINATIVE_LLM_ANTHROPIC_RETRY_ATTEMPTS
- **Type:** Integer
- **Default:** `3`
- **Description:** Number of retry attempts
- **Example:** `AINATIVE_LLM_ANTHROPIC_RETRY_ATTEMPTS=5`

#### AINATIVE_LLM_ANTHROPIC_BASE_URL
- **Type:** String
- **Default:** (empty, uses official API)
- **Description:** Custom base URL for API
- **Example:** `AINATIVE_LLM_ANTHROPIC_BASE_URL=https://api.custom.com`

#### AINATIVE_LLM_ANTHROPIC_API_VERSION
- **Type:** String
- **Default:** `2023-06-01`
- **Description:** API version to use
- **Example:** `AINATIVE_LLM_ANTHROPIC_API_VERSION=2023-06-01`

### OpenAI

#### AINATIVE_LLM_OPENAI_API_KEY
- **Type:** String
- **Required:** Yes (if using OpenAI)
- **Description:** OpenAI API key
- **Example:** `AINATIVE_LLM_OPENAI_API_KEY=sk-...`
- **Get Key:** https://platform.openai.com/api-keys

#### AINATIVE_LLM_OPENAI_MODEL
- **Type:** String
- **Default:** `gpt-4-turbo-preview`
- **Description:** OpenAI model to use
- **Example:** `AINATIVE_LLM_OPENAI_MODEL=gpt-4`

#### AINATIVE_LLM_OPENAI_ORGANIZATION
- **Type:** String
- **Default:** (empty)
- **Description:** OpenAI organization ID
- **Example:** `AINATIVE_LLM_OPENAI_ORGANIZATION=org-...`

#### AINATIVE_LLM_OPENAI_MAX_TOKENS
- **Type:** Integer
- **Default:** `4096`
- **Description:** Maximum tokens in response
- **Example:** `AINATIVE_LLM_OPENAI_MAX_TOKENS=8192`

#### AINATIVE_LLM_OPENAI_TEMPERATURE
- **Type:** Float
- **Default:** `0.7`
- **Range:** `0.0` - `2.0`
- **Description:** Sampling temperature
- **Example:** `AINATIVE_LLM_OPENAI_TEMPERATURE=0.8`

#### AINATIVE_LLM_OPENAI_TOP_P
- **Type:** Float
- **Default:** `1.0`
- **Range:** `0.0` - `1.0`
- **Description:** Nucleus sampling parameter
- **Example:** `AINATIVE_LLM_OPENAI_TOP_P=0.95`

#### AINATIVE_LLM_OPENAI_FREQUENCY_PENALTY
- **Type:** Float
- **Default:** `0.0`
- **Range:** `-2.0` - `2.0`
- **Description:** Frequency penalty
- **Example:** `AINATIVE_LLM_OPENAI_FREQUENCY_PENALTY=0.5`

#### AINATIVE_LLM_OPENAI_PRESENCE_PENALTY
- **Type:** Float
- **Default:** `0.0`
- **Range:** `-2.0` - `2.0`
- **Description:** Presence penalty
- **Example:** `AINATIVE_LLM_OPENAI_PRESENCE_PENALTY=0.5`

#### AINATIVE_LLM_OPENAI_TIMEOUT
- **Type:** Duration
- **Default:** `30s`
- **Description:** Request timeout
- **Example:** `AINATIVE_LLM_OPENAI_TIMEOUT=45s`

#### AINATIVE_LLM_OPENAI_RETRY_ATTEMPTS
- **Type:** Integer
- **Default:** `3`
- **Description:** Number of retry attempts
- **Example:** `AINATIVE_LLM_OPENAI_RETRY_ATTEMPTS=5`

#### AINATIVE_LLM_OPENAI_BASE_URL
- **Type:** String
- **Default:** (empty, uses official API)
- **Description:** Custom base URL for API
- **Example:** `AINATIVE_LLM_OPENAI_BASE_URL=https://api.custom.com/v1`

### Google Gemini

#### AINATIVE_LLM_GOOGLE_API_KEY
- **Type:** String
- **Required:** Yes (if using Google)
- **Description:** Google API key
- **Example:** `AINATIVE_LLM_GOOGLE_API_KEY=AIza...`
- **Get Key:** https://makersuite.google.com/app/apikey

#### AINATIVE_LLM_GOOGLE_MODEL
- **Type:** String
- **Default:** `gemini-pro`
- **Description:** Gemini model to use
- **Example:** `AINATIVE_LLM_GOOGLE_MODEL=gemini-pro-vision`

#### AINATIVE_LLM_GOOGLE_PROJECT_ID
- **Type:** String
- **Default:** (empty)
- **Description:** Google Cloud project ID (for Vertex AI)
- **Example:** `AINATIVE_LLM_GOOGLE_PROJECT_ID=my-project-123`

#### AINATIVE_LLM_GOOGLE_LOCATION
- **Type:** String
- **Default:** `us-central1`
- **Description:** Google Cloud region (for Vertex AI)
- **Example:** `AINATIVE_LLM_GOOGLE_LOCATION=europe-west1`

#### AINATIVE_LLM_GOOGLE_MAX_TOKENS
- **Type:** Integer
- **Default:** `4096`
- **Description:** Maximum tokens in response
- **Example:** `AINATIVE_LLM_GOOGLE_MAX_TOKENS=8192`

#### AINATIVE_LLM_GOOGLE_TEMPERATURE
- **Type:** Float
- **Default:** `0.7`
- **Range:** `0.0` - `1.0`
- **Description:** Sampling temperature
- **Example:** `AINATIVE_LLM_GOOGLE_TEMPERATURE=0.6`

#### AINATIVE_LLM_GOOGLE_TOP_P
- **Type:** Float
- **Default:** `1.0`
- **Range:** `0.0` - `1.0`
- **Description:** Nucleus sampling parameter
- **Example:** `AINATIVE_LLM_GOOGLE_TOP_P=0.95`

#### AINATIVE_LLM_GOOGLE_TOP_K
- **Type:** Integer
- **Default:** `40`
- **Description:** Top-K sampling parameter
- **Example:** `AINATIVE_LLM_GOOGLE_TOP_K=50`

#### AINATIVE_LLM_GOOGLE_TIMEOUT
- **Type:** Duration
- **Default:** `30s`
- **Description:** Request timeout
- **Example:** `AINATIVE_LLM_GOOGLE_TIMEOUT=60s`

#### AINATIVE_LLM_GOOGLE_RETRY_ATTEMPTS
- **Type:** Integer
- **Default:** `3`
- **Description:** Number of retry attempts
- **Example:** `AINATIVE_LLM_GOOGLE_RETRY_ATTEMPTS=5`

### AWS Bedrock

#### AINATIVE_LLM_BEDROCK_REGION
- **Type:** String
- **Required:** Yes (if using Bedrock)
- **Description:** AWS region
- **Example:** `AINATIVE_LLM_BEDROCK_REGION=us-east-1`

#### AINATIVE_LLM_BEDROCK_MODEL
- **Type:** String
- **Default:** `anthropic.claude-3-sonnet-20240229-v1:0`
- **Description:** Bedrock model ARN
- **Example:** `AINATIVE_LLM_BEDROCK_MODEL=anthropic.claude-3-haiku-20240307-v1:0`

#### AINATIVE_LLM_BEDROCK_ACCESS_KEY_ID
- **Type:** String
- **Default:** (empty, uses AWS credentials chain)
- **Description:** AWS access key ID
- **Example:** `AINATIVE_LLM_BEDROCK_ACCESS_KEY_ID=AKIA...`

#### AINATIVE_LLM_BEDROCK_SECRET_ACCESS_KEY
- **Type:** String
- **Default:** (empty, uses AWS credentials chain)
- **Description:** AWS secret access key
- **Example:** `AINATIVE_LLM_BEDROCK_SECRET_ACCESS_KEY=...`

#### AINATIVE_LLM_BEDROCK_SESSION_TOKEN
- **Type:** String
- **Default:** (empty)
- **Description:** AWS session token (for temporary credentials)
- **Example:** `AINATIVE_LLM_BEDROCK_SESSION_TOKEN=...`

#### AINATIVE_LLM_BEDROCK_PROFILE
- **Type:** String
- **Default:** `default`
- **Description:** AWS profile name
- **Example:** `AINATIVE_LLM_BEDROCK_PROFILE=production`

#### AINATIVE_LLM_BEDROCK_MAX_TOKENS
- **Type:** Integer
- **Default:** `4096`
- **Description:** Maximum tokens in response
- **Example:** `AINATIVE_LLM_BEDROCK_MAX_TOKENS=8192`

#### AINATIVE_LLM_BEDROCK_TEMPERATURE
- **Type:** Float
- **Default:** `0.7`
- **Range:** `0.0` - `1.0`
- **Description:** Sampling temperature
- **Example:** `AINATIVE_LLM_BEDROCK_TEMPERATURE=0.5`

#### AINATIVE_LLM_BEDROCK_TOP_P
- **Type:** Float
- **Default:** `1.0`
- **Range:** `0.0` - `1.0`
- **Description:** Nucleus sampling parameter
- **Example:** `AINATIVE_LLM_BEDROCK_TOP_P=0.9`

#### AINATIVE_LLM_BEDROCK_TIMEOUT
- **Type:** Duration
- **Default:** `60s`
- **Description:** Request timeout
- **Example:** `AINATIVE_LLM_BEDROCK_TIMEOUT=120s`

#### AINATIVE_LLM_BEDROCK_RETRY_ATTEMPTS
- **Type:** Integer
- **Default:** `3`
- **Description:** Number of retry attempts
- **Example:** `AINATIVE_LLM_BEDROCK_RETRY_ATTEMPTS=5`

### Azure OpenAI

#### AINATIVE_LLM_AZURE_API_KEY
- **Type:** String
- **Required:** Yes (if using Azure)
- **Description:** Azure OpenAI API key
- **Example:** `AINATIVE_LLM_AZURE_API_KEY=...`

#### AINATIVE_LLM_AZURE_ENDPOINT
- **Type:** String
- **Required:** Yes (if using Azure)
- **Description:** Azure OpenAI endpoint URL
- **Example:** `AINATIVE_LLM_AZURE_ENDPOINT=https://my-resource.openai.azure.com`

#### AINATIVE_LLM_AZURE_DEPLOYMENT_NAME
- **Type:** String
- **Required:** Yes (if using Azure)
- **Description:** Azure deployment name
- **Example:** `AINATIVE_LLM_AZURE_DEPLOYMENT_NAME=gpt-4-deployment`

#### AINATIVE_LLM_AZURE_API_VERSION
- **Type:** String
- **Default:** `2023-05-15`
- **Description:** Azure API version
- **Example:** `AINATIVE_LLM_AZURE_API_VERSION=2023-12-01`

#### AINATIVE_LLM_AZURE_MAX_TOKENS
- **Type:** Integer
- **Default:** `4096`
- **Description:** Maximum tokens in response
- **Example:** `AINATIVE_LLM_AZURE_MAX_TOKENS=8192`

#### AINATIVE_LLM_AZURE_TEMPERATURE
- **Type:** Float
- **Default:** `0.7`
- **Range:** `0.0` - `2.0`
- **Description:** Sampling temperature
- **Example:** `AINATIVE_LLM_AZURE_TEMPERATURE=0.8`

#### AINATIVE_LLM_AZURE_TOP_P
- **Type:** Float
- **Default:** `1.0`
- **Range:** `0.0` - `1.0`
- **Description:** Nucleus sampling parameter
- **Example:** `AINATIVE_LLM_AZURE_TOP_P=0.95`

#### AINATIVE_LLM_AZURE_TIMEOUT
- **Type:** Duration
- **Default:** `30s`
- **Description:** Request timeout
- **Example:** `AINATIVE_LLM_AZURE_TIMEOUT=60s`

#### AINATIVE_LLM_AZURE_RETRY_ATTEMPTS
- **Type:** Integer
- **Default:** `3`
- **Description:** Number of retry attempts
- **Example:** `AINATIVE_LLM_AZURE_RETRY_ATTEMPTS=5`

### Ollama (Local Models)

#### AINATIVE_LLM_OLLAMA_BASE_URL
- **Type:** String
- **Default:** `http://localhost:11434`
- **Description:** Ollama server URL
- **Example:** `AINATIVE_LLM_OLLAMA_BASE_URL=http://192.168.1.100:11434`

#### AINATIVE_LLM_OLLAMA_MODEL
- **Type:** String
- **Required:** Yes (if using Ollama)
- **Description:** Ollama model name
- **Example:** `AINATIVE_LLM_OLLAMA_MODEL=llama2`

#### AINATIVE_LLM_OLLAMA_MAX_TOKENS
- **Type:** Integer
- **Default:** `4096`
- **Description:** Maximum tokens in response
- **Example:** `AINATIVE_LLM_OLLAMA_MAX_TOKENS=8192`

#### AINATIVE_LLM_OLLAMA_TEMPERATURE
- **Type:** Float
- **Default:** `0.7`
- **Range:** `0.0` - `1.0`
- **Description:** Sampling temperature
- **Example:** `AINATIVE_LLM_OLLAMA_TEMPERATURE=0.5`

#### AINATIVE_LLM_OLLAMA_TOP_P
- **Type:** Float
- **Default:** `1.0`
- **Range:** `0.0` - `1.0`
- **Description:** Nucleus sampling parameter
- **Example:** `AINATIVE_LLM_OLLAMA_TOP_P=0.9`

#### AINATIVE_LLM_OLLAMA_TOP_K
- **Type:** Integer
- **Default:** `40`
- **Description:** Top-K sampling parameter
- **Example:** `AINATIVE_LLM_OLLAMA_TOP_K=50`

#### AINATIVE_LLM_OLLAMA_TIMEOUT
- **Type:** Duration
- **Default:** `120s`
- **Description:** Request timeout
- **Example:** `AINATIVE_LLM_OLLAMA_TIMEOUT=180s`

#### AINATIVE_LLM_OLLAMA_RETRY_ATTEMPTS
- **Type:** Integer
- **Default:** `1`
- **Description:** Number of retry attempts
- **Example:** `AINATIVE_LLM_OLLAMA_RETRY_ATTEMPTS=3`

#### AINATIVE_LLM_OLLAMA_KEEP_ALIVE
- **Type:** Duration
- **Default:** `5m`
- **Description:** Model keep-alive duration
- **Example:** `AINATIVE_LLM_OLLAMA_KEEP_ALIVE=10m`

## Platform Authentication

### AINATIVE_PLATFORM_AUTHENTICATION_METHOD
- **Type:** String
- **Default:** `api_key`
- **Valid Values:** `api_key`, `jwt`, `oauth2`
- **Description:** Authentication method
- **Example:** `AINATIVE_PLATFORM_AUTHENTICATION_METHOD=jwt`

### AINATIVE_PLATFORM_AUTHENTICATION_API_KEY
- **Type:** String
- **Description:** API key for authentication
- **Example:** `AINATIVE_PLATFORM_AUTHENTICATION_API_KEY=your-api-key`

### AINATIVE_PLATFORM_AUTHENTICATION_TOKEN
- **Type:** String
- **Description:** JWT token
- **Example:** `AINATIVE_PLATFORM_AUTHENTICATION_TOKEN=eyJ...`

### AINATIVE_PLATFORM_AUTHENTICATION_REFRESH_TOKEN
- **Type:** String
- **Description:** JWT refresh token
- **Example:** `AINATIVE_PLATFORM_AUTHENTICATION_REFRESH_TOKEN=...`

### AINATIVE_PLATFORM_AUTHENTICATION_CLIENT_ID
- **Type:** String
- **Description:** OAuth2 client ID
- **Example:** `AINATIVE_PLATFORM_AUTHENTICATION_CLIENT_ID=client-id`

### AINATIVE_PLATFORM_AUTHENTICATION_CLIENT_SECRET
- **Type:** String
- **Description:** OAuth2 client secret
- **Example:** `AINATIVE_PLATFORM_AUTHENTICATION_CLIENT_SECRET=client-secret`

### AINATIVE_PLATFORM_AUTHENTICATION_TOKEN_URL
- **Type:** String
- **Description:** OAuth2 token URL
- **Example:** `AINATIVE_PLATFORM_AUTHENTICATION_TOKEN_URL=https://auth.example.com/token`

### AINATIVE_PLATFORM_AUTHENTICATION_TIMEOUT
- **Type:** Duration
- **Default:** `10s`
- **Description:** Authentication timeout
- **Example:** `AINATIVE_PLATFORM_AUTHENTICATION_TIMEOUT=15s`

### AINATIVE_PLATFORM_ORGANIZATION_ID
- **Type:** String
- **Description:** Organization ID
- **Example:** `AINATIVE_PLATFORM_ORGANIZATION_ID=org-123`

### AINATIVE_PLATFORM_ORGANIZATION_NAME
- **Type:** String
- **Description:** Organization name
- **Example:** `AINATIVE_PLATFORM_ORGANIZATION_NAME=My Org`

### AINATIVE_PLATFORM_ORGANIZATION_WORKSPACE
- **Type:** String
- **Default:** `default`
- **Description:** Workspace name
- **Example:** `AINATIVE_PLATFORM_ORGANIZATION_WORKSPACE=production`

## Service Endpoints

### ZeroDB

#### AINATIVE_SERVICES_ZERODB_ENABLED
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable ZeroDB service
- **Example:** `AINATIVE_SERVICES_ZERODB_ENABLED=true`

#### AINATIVE_SERVICES_ZERODB_PROJECT_ID
- **Type:** String
- **Description:** ZeroDB project ID
- **Example:** `AINATIVE_SERVICES_ZERODB_PROJECT_ID=proj-abc123`

#### AINATIVE_SERVICES_ZERODB_CONNECTION_STRING
- **Type:** String
- **Description:** ZeroDB connection string
- **Example:** `AINATIVE_SERVICES_ZERODB_CONNECTION_STRING=postgresql://user:pass@host:5432/db`

#### AINATIVE_SERVICES_ZERODB_ENDPOINT
- **Type:** String
- **Description:** ZeroDB endpoint URL
- **Example:** `AINATIVE_SERVICES_ZERODB_ENDPOINT=postgresql://localhost:5432`

#### AINATIVE_SERVICES_ZERODB_DATABASE
- **Type:** String
- **Default:** `ainative_code`
- **Description:** Database name
- **Example:** `AINATIVE_SERVICES_ZERODB_DATABASE=production_db`

#### AINATIVE_SERVICES_ZERODB_USERNAME
- **Type:** String
- **Description:** Database username
- **Example:** `AINATIVE_SERVICES_ZERODB_USERNAME=dbuser`

#### AINATIVE_SERVICES_ZERODB_PASSWORD
- **Type:** String
- **Description:** Database password
- **Example:** `AINATIVE_SERVICES_ZERODB_PASSWORD=secure-password`

#### AINATIVE_SERVICES_ZERODB_SSL
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable SSL
- **Example:** `AINATIVE_SERVICES_ZERODB_SSL=true`

#### AINATIVE_SERVICES_ZERODB_SSL_MODE
- **Type:** String
- **Default:** `disable`
- **Valid Values:** `disable`, `require`, `verify-ca`, `verify-full`
- **Description:** SSL mode
- **Example:** `AINATIVE_SERVICES_ZERODB_SSL_MODE=require`

#### AINATIVE_SERVICES_ZERODB_MAX_CONNECTIONS
- **Type:** Integer
- **Default:** `10`
- **Description:** Maximum connections
- **Example:** `AINATIVE_SERVICES_ZERODB_MAX_CONNECTIONS=20`

#### AINATIVE_SERVICES_ZERODB_IDLE_CONNECTIONS
- **Type:** Integer
- **Default:** `2`
- **Description:** Idle connections
- **Example:** `AINATIVE_SERVICES_ZERODB_IDLE_CONNECTIONS=5`

#### AINATIVE_SERVICES_ZERODB_CONN_MAX_LIFETIME
- **Type:** Duration
- **Default:** `1h`
- **Description:** Connection max lifetime
- **Example:** `AINATIVE_SERVICES_ZERODB_CONN_MAX_LIFETIME=2h`

#### AINATIVE_SERVICES_ZERODB_TIMEOUT
- **Type:** Duration
- **Default:** `5s`
- **Description:** Query timeout
- **Example:** `AINATIVE_SERVICES_ZERODB_TIMEOUT=10s`

#### AINATIVE_SERVICES_ZERODB_RETRY_ATTEMPTS
- **Type:** Integer
- **Default:** `3`
- **Description:** Retry attempts
- **Example:** `AINATIVE_SERVICES_ZERODB_RETRY_ATTEMPTS=5`

#### AINATIVE_SERVICES_ZERODB_RETRY_DELAY
- **Type:** Duration
- **Default:** `1s`
- **Description:** Retry delay
- **Example:** `AINATIVE_SERVICES_ZERODB_RETRY_DELAY=2s`

### Design Service

#### AINATIVE_SERVICES_DESIGN_ENABLED
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable Design service
- **Example:** `AINATIVE_SERVICES_DESIGN_ENABLED=true`

#### AINATIVE_SERVICES_DESIGN_ENDPOINT
- **Type:** String
- **Description:** Design service endpoint
- **Example:** `AINATIVE_SERVICES_DESIGN_ENDPOINT=https://design.ainative.studio/api`

#### AINATIVE_SERVICES_DESIGN_API_KEY
- **Type:** String
- **Description:** Design service API key
- **Example:** `AINATIVE_SERVICES_DESIGN_API_KEY=design-key`

#### AINATIVE_SERVICES_DESIGN_TIMEOUT
- **Type:** Duration
- **Default:** `30s`
- **Description:** Request timeout
- **Example:** `AINATIVE_SERVICES_DESIGN_TIMEOUT=60s`

#### AINATIVE_SERVICES_DESIGN_RETRY_ATTEMPTS
- **Type:** Integer
- **Default:** `3`
- **Description:** Retry attempts
- **Example:** `AINATIVE_SERVICES_DESIGN_RETRY_ATTEMPTS=5`

### Strapi CMS

#### AINATIVE_SERVICES_STRAPI_ENABLED
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable Strapi service
- **Example:** `AINATIVE_SERVICES_STRAPI_ENABLED=true`

#### AINATIVE_SERVICES_STRAPI_ENDPOINT
- **Type:** String
- **Description:** Strapi endpoint
- **Example:** `AINATIVE_SERVICES_STRAPI_ENDPOINT=https://strapi.example.com`

#### AINATIVE_SERVICES_STRAPI_API_KEY
- **Type:** String
- **Description:** Strapi API key
- **Example:** `AINATIVE_SERVICES_STRAPI_API_KEY=strapi-key`

#### AINATIVE_SERVICES_STRAPI_TIMEOUT
- **Type:** Duration
- **Default:** `30s`
- **Description:** Request timeout
- **Example:** `AINATIVE_SERVICES_STRAPI_TIMEOUT=60s`

#### AINATIVE_SERVICES_STRAPI_RETRY_ATTEMPTS
- **Type:** Integer
- **Default:** `3`
- **Description:** Retry attempts
- **Example:** `AINATIVE_SERVICES_STRAPI_RETRY_ATTEMPTS=5`

### RLHF Service

#### AINATIVE_SERVICES_RLHF_ENABLED
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable RLHF service
- **Example:** `AINATIVE_SERVICES_RLHF_ENABLED=true`

#### AINATIVE_SERVICES_RLHF_ENDPOINT
- **Type:** String
- **Description:** RLHF service endpoint
- **Example:** `AINATIVE_SERVICES_RLHF_ENDPOINT=https://rlhf.ainative.studio`

#### AINATIVE_SERVICES_RLHF_API_KEY
- **Type:** String
- **Description:** RLHF service API key
- **Example:** `AINATIVE_SERVICES_RLHF_API_KEY=rlhf-key`

#### AINATIVE_SERVICES_RLHF_TIMEOUT
- **Type:** Duration
- **Default:** `60s`
- **Description:** Request timeout
- **Example:** `AINATIVE_SERVICES_RLHF_TIMEOUT=120s`

#### AINATIVE_SERVICES_RLHF_RETRY_ATTEMPTS
- **Type:** Integer
- **Default:** `3`
- **Description:** Retry attempts
- **Example:** `AINATIVE_SERVICES_RLHF_RETRY_ATTEMPTS=5`

#### AINATIVE_SERVICES_RLHF_MODEL_ID
- **Type:** String
- **Description:** RLHF model ID
- **Example:** `AINATIVE_SERVICES_RLHF_MODEL_ID=model-123`

## Tool Configuration

### Filesystem Tool

#### AINATIVE_TOOLS_FILESYSTEM_ENABLED
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable filesystem tool
- **Example:** `AINATIVE_TOOLS_FILESYSTEM_ENABLED=true`

#### AINATIVE_TOOLS_FILESYSTEM_MAX_FILE_SIZE
- **Type:** Integer
- **Default:** `104857600` (100MB)
- **Description:** Maximum file size in bytes
- **Example:** `AINATIVE_TOOLS_FILESYSTEM_MAX_FILE_SIZE=52428800`

### Terminal Tool

#### AINATIVE_TOOLS_TERMINAL_ENABLED
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable terminal tool
- **Example:** `AINATIVE_TOOLS_TERMINAL_ENABLED=true`

#### AINATIVE_TOOLS_TERMINAL_TIMEOUT
- **Type:** Duration
- **Default:** `5m`
- **Description:** Command timeout
- **Example:** `AINATIVE_TOOLS_TERMINAL_TIMEOUT=10m`

#### AINATIVE_TOOLS_TERMINAL_WORKING_DIR
- **Type:** String
- **Description:** Working directory
- **Example:** `AINATIVE_TOOLS_TERMINAL_WORKING_DIR=/home/user/projects`

### Browser Tool

#### AINATIVE_TOOLS_BROWSER_ENABLED
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable browser tool
- **Example:** `AINATIVE_TOOLS_BROWSER_ENABLED=true`

#### AINATIVE_TOOLS_BROWSER_HEADLESS
- **Type:** Boolean
- **Default:** `true`
- **Description:** Run browser in headless mode
- **Example:** `AINATIVE_TOOLS_BROWSER_HEADLESS=false`

#### AINATIVE_TOOLS_BROWSER_TIMEOUT
- **Type:** Duration
- **Default:** `30s`
- **Description:** Browser timeout
- **Example:** `AINATIVE_TOOLS_BROWSER_TIMEOUT=60s`

#### AINATIVE_TOOLS_BROWSER_USER_AGENT
- **Type:** String
- **Default:** `AINative-Code/0.1.0`
- **Description:** Browser user agent
- **Example:** `AINATIVE_TOOLS_BROWSER_USER_AGENT=CustomBot/1.0`

### Code Analysis Tool

#### AINATIVE_TOOLS_CODE_ANALYSIS_ENABLED
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable code analysis tool
- **Example:** `AINATIVE_TOOLS_CODE_ANALYSIS_ENABLED=true`

#### AINATIVE_TOOLS_CODE_ANALYSIS_MAX_FILE_SIZE
- **Type:** Integer
- **Default:** `10485760` (10MB)
- **Description:** Maximum file size in bytes
- **Example:** `AINATIVE_TOOLS_CODE_ANALYSIS_MAX_FILE_SIZE=20971520`

#### AINATIVE_TOOLS_CODE_ANALYSIS_INCLUDE_TESTS
- **Type:** Boolean
- **Default:** `true`
- **Description:** Include test files
- **Example:** `AINATIVE_TOOLS_CODE_ANALYSIS_INCLUDE_TESTS=false`

## Performance Settings

### Cache

#### AINATIVE_PERFORMANCE_CACHE_ENABLED
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable caching
- **Example:** `AINATIVE_PERFORMANCE_CACHE_ENABLED=true`

#### AINATIVE_PERFORMANCE_CACHE_TYPE
- **Type:** String
- **Default:** `memory`
- **Valid Values:** `memory`, `redis`, `memcached`
- **Description:** Cache type
- **Example:** `AINATIVE_PERFORMANCE_CACHE_TYPE=redis`

#### AINATIVE_PERFORMANCE_CACHE_TTL
- **Type:** Duration
- **Default:** `1h`
- **Description:** Cache TTL
- **Example:** `AINATIVE_PERFORMANCE_CACHE_TTL=30m`

#### AINATIVE_PERFORMANCE_CACHE_MAX_SIZE
- **Type:** Integer
- **Default:** `100` (MB)
- **Description:** Maximum cache size
- **Example:** `AINATIVE_PERFORMANCE_CACHE_MAX_SIZE=500`

#### AINATIVE_PERFORMANCE_CACHE_REDIS_URL
- **Type:** String
- **Description:** Redis URL
- **Example:** `AINATIVE_PERFORMANCE_CACHE_REDIS_URL=redis://localhost:6379/0`

#### AINATIVE_PERFORMANCE_CACHE_MEMCACHED_URL
- **Type:** String
- **Description:** Memcached URL
- **Example:** `AINATIVE_PERFORMANCE_CACHE_MEMCACHED_URL=localhost:11211`

### Rate Limiting

#### AINATIVE_PERFORMANCE_RATE_LIMIT_ENABLED
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable rate limiting
- **Example:** `AINATIVE_PERFORMANCE_RATE_LIMIT_ENABLED=true`

#### AINATIVE_PERFORMANCE_RATE_LIMIT_REQUESTS_PER_MINUTE
- **Type:** Integer
- **Default:** `60`
- **Description:** Requests per minute
- **Example:** `AINATIVE_PERFORMANCE_RATE_LIMIT_REQUESTS_PER_MINUTE=120`

#### AINATIVE_PERFORMANCE_RATE_LIMIT_BURST_SIZE
- **Type:** Integer
- **Default:** `10`
- **Description:** Burst size
- **Example:** `AINATIVE_PERFORMANCE_RATE_LIMIT_BURST_SIZE=20`

#### AINATIVE_PERFORMANCE_RATE_LIMIT_TIME_WINDOW
- **Type:** Duration
- **Default:** `1m`
- **Description:** Time window
- **Example:** `AINATIVE_PERFORMANCE_RATE_LIMIT_TIME_WINDOW=30s`

### Concurrency

#### AINATIVE_PERFORMANCE_CONCURRENCY_MAX_WORKERS
- **Type:** Integer
- **Default:** `10`
- **Description:** Maximum workers
- **Example:** `AINATIVE_PERFORMANCE_CONCURRENCY_MAX_WORKERS=20`

#### AINATIVE_PERFORMANCE_CONCURRENCY_MAX_QUEUE_SIZE
- **Type:** Integer
- **Default:** `100`
- **Description:** Maximum queue size
- **Example:** `AINATIVE_PERFORMANCE_CONCURRENCY_MAX_QUEUE_SIZE=200`

#### AINATIVE_PERFORMANCE_CONCURRENCY_WORKER_TIMEOUT
- **Type:** Duration
- **Default:** `5m`
- **Description:** Worker timeout
- **Example:** `AINATIVE_PERFORMANCE_CONCURRENCY_WORKER_TIMEOUT=10m`

### Circuit Breaker

#### AINATIVE_PERFORMANCE_CIRCUIT_BREAKER_ENABLED
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable circuit breaker
- **Example:** `AINATIVE_PERFORMANCE_CIRCUIT_BREAKER_ENABLED=true`

#### AINATIVE_PERFORMANCE_CIRCUIT_BREAKER_FAILURE_THRESHOLD
- **Type:** Integer
- **Default:** `5`
- **Description:** Failure threshold
- **Example:** `AINATIVE_PERFORMANCE_CIRCUIT_BREAKER_FAILURE_THRESHOLD=10`

#### AINATIVE_PERFORMANCE_CIRCUIT_BREAKER_SUCCESS_THRESHOLD
- **Type:** Integer
- **Default:** `2`
- **Description:** Success threshold
- **Example:** `AINATIVE_PERFORMANCE_CIRCUIT_BREAKER_SUCCESS_THRESHOLD=3`

#### AINATIVE_PERFORMANCE_CIRCUIT_BREAKER_TIMEOUT
- **Type:** Duration
- **Default:** `60s`
- **Description:** Request timeout
- **Example:** `AINATIVE_PERFORMANCE_CIRCUIT_BREAKER_TIMEOUT=120s`

#### AINATIVE_PERFORMANCE_CIRCUIT_BREAKER_RESET_TIMEOUT
- **Type:** Duration
- **Default:** `30s`
- **Description:** Reset timeout
- **Example:** `AINATIVE_PERFORMANCE_CIRCUIT_BREAKER_RESET_TIMEOUT=60s`

## Logging Configuration

### AINATIVE_LOGGING_LEVEL
- **Type:** String
- **Default:** `info`
- **Valid Values:** `debug`, `info`, `warn`, `error`
- **Description:** Log level
- **Example:** `AINATIVE_LOGGING_LEVEL=debug`

### AINATIVE_LOGGING_FORMAT
- **Type:** String
- **Default:** `json`
- **Valid Values:** `json`, `console`
- **Description:** Log format
- **Example:** `AINATIVE_LOGGING_FORMAT=console`

### AINATIVE_LOGGING_OUTPUT
- **Type:** String
- **Default:** `stdout`
- **Valid Values:** `stdout`, `file`
- **Description:** Log output
- **Example:** `AINATIVE_LOGGING_OUTPUT=file`

### AINATIVE_LOGGING_FILE_PATH
- **Type:** String
- **Description:** Log file path
- **Example:** `AINATIVE_LOGGING_FILE_PATH=/var/log/ainative-code/app.log`

### AINATIVE_LOGGING_MAX_SIZE
- **Type:** Integer
- **Default:** `100` (MB)
- **Description:** Maximum log file size
- **Example:** `AINATIVE_LOGGING_MAX_SIZE=200`

### AINATIVE_LOGGING_MAX_BACKUPS
- **Type:** Integer
- **Default:** `3`
- **Description:** Maximum backup files
- **Example:** `AINATIVE_LOGGING_MAX_BACKUPS=5`

### AINATIVE_LOGGING_MAX_AGE
- **Type:** Integer
- **Default:** `7` (days)
- **Description:** Maximum age of log files
- **Example:** `AINATIVE_LOGGING_MAX_AGE=14`

### AINATIVE_LOGGING_COMPRESS
- **Type:** Boolean
- **Default:** `true`
- **Description:** Compress rotated logs
- **Example:** `AINATIVE_LOGGING_COMPRESS=false`

## Security Settings

### AINATIVE_SECURITY_ENCRYPT_CONFIG
- **Type:** Boolean
- **Default:** `false`
- **Description:** Encrypt configuration
- **Example:** `AINATIVE_SECURITY_ENCRYPT_CONFIG=true`

### AINATIVE_SECURITY_ENCRYPTION_KEY
- **Type:** String
- **Description:** Encryption key (32+ characters)
- **Example:** `AINATIVE_SECURITY_ENCRYPTION_KEY=your-32-character-encryption-key`

### AINATIVE_SECURITY_TLS_ENABLED
- **Type:** Boolean
- **Default:** `false`
- **Description:** Enable TLS
- **Example:** `AINATIVE_SECURITY_TLS_ENABLED=true`

### AINATIVE_SECURITY_TLS_CERT_PATH
- **Type:** String
- **Description:** TLS certificate path
- **Example:** `AINATIVE_SECURITY_TLS_CERT_PATH=/etc/ssl/certs/cert.pem`

### AINATIVE_SECURITY_TLS_KEY_PATH
- **Type:** String
- **Description:** TLS key path
- **Example:** `AINATIVE_SECURITY_TLS_KEY_PATH=/etc/ssl/private/key.pem`

### AINATIVE_SECURITY_SECRET_ROTATION
- **Type:** Duration
- **Description:** Secret rotation period
- **Example:** `AINATIVE_SECURITY_SECRET_ROTATION=90d`

## Complete Variable Reference

### Quick Reference Template

```bash
# Application
export AINATIVE_APP_NAME=ainative-code
export AINATIVE_APP_VERSION=0.1.0
export AINATIVE_APP_ENVIRONMENT=production
export AINATIVE_APP_DEBUG=false

# LLM - Anthropic
export AINATIVE_LLM_DEFAULT_PROVIDER=anthropic
export AINATIVE_LLM_ANTHROPIC_API_KEY=sk-ant-...
export AINATIVE_LLM_ANTHROPIC_MODEL=claude-3-5-sonnet-20241022
export AINATIVE_LLM_ANTHROPIC_MAX_TOKENS=4096
export AINATIVE_LLM_ANTHROPIC_TEMPERATURE=0.7

# LLM - OpenAI (optional)
export AINATIVE_LLM_OPENAI_API_KEY=sk-...
export AINATIVE_LLM_OPENAI_MODEL=gpt-4-turbo-preview
export AINATIVE_LLM_OPENAI_ORGANIZATION=org-...

# LLM - Google Gemini (optional)
export AINATIVE_LLM_GOOGLE_API_KEY=AIza...
export AINATIVE_LLM_GOOGLE_MODEL=gemini-pro

# LLM - AWS Bedrock (optional)
export AINATIVE_LLM_BEDROCK_REGION=us-east-1
export AINATIVE_LLM_BEDROCK_MODEL=anthropic.claude-3-sonnet-20240229-v1:0
export AINATIVE_LLM_BEDROCK_ACCESS_KEY_ID=AKIA...
export AINATIVE_LLM_BEDROCK_SECRET_ACCESS_KEY=...

# LLM - Azure OpenAI (optional)
export AINATIVE_LLM_AZURE_API_KEY=...
export AINATIVE_LLM_AZURE_ENDPOINT=https://resource.openai.azure.com
export AINATIVE_LLM_AZURE_DEPLOYMENT_NAME=gpt-4-deployment

# LLM - Ollama (optional)
export AINATIVE_LLM_OLLAMA_BASE_URL=http://localhost:11434
export AINATIVE_LLM_OLLAMA_MODEL=llama2

# Platform Authentication
export AINATIVE_PLATFORM_AUTHENTICATION_METHOD=api_key
export AINATIVE_PLATFORM_AUTHENTICATION_API_KEY=your-api-key
export AINATIVE_PLATFORM_ORGANIZATION_ID=org-123

# ZeroDB
export AINATIVE_SERVICES_ZERODB_ENABLED=true
export AINATIVE_SERVICES_ZERODB_PROJECT_ID=proj-abc123
export AINATIVE_SERVICES_ZERODB_ENDPOINT=postgresql://localhost:5432
export AINATIVE_SERVICES_ZERODB_DATABASE=ainative_code
export AINATIVE_SERVICES_ZERODB_USERNAME=dbuser
export AINATIVE_SERVICES_ZERODB_PASSWORD=dbpass

# Design Service
export AINATIVE_SERVICES_DESIGN_ENABLED=true
export AINATIVE_SERVICES_DESIGN_ENDPOINT=https://design.ainative.studio/api
export AINATIVE_SERVICES_DESIGN_API_KEY=design-key

# Strapi CMS
export AINATIVE_SERVICES_STRAPI_ENABLED=false
export AINATIVE_SERVICES_STRAPI_ENDPOINT=https://strapi.example.com
export AINATIVE_SERVICES_STRAPI_API_KEY=strapi-key

# RLHF Service
export AINATIVE_SERVICES_RLHF_ENABLED=false
export AINATIVE_SERVICES_RLHF_ENDPOINT=https://rlhf.ainative.studio
export AINATIVE_SERVICES_RLHF_API_KEY=rlhf-key

# Logging
export AINATIVE_LOGGING_LEVEL=info
export AINATIVE_LOGGING_FORMAT=json
export AINATIVE_LOGGING_OUTPUT=stdout

# Security
export AINATIVE_SECURITY_ENCRYPT_CONFIG=false
export AINATIVE_SECURITY_TLS_ENABLED=false
```

## Best Practices

1. **Never commit secrets** - Use environment variables for all sensitive data
2. **Use .env files** - Keep environment-specific variables in `.env` files (add to `.gitignore`)
3. **Validate before deployment** - Test configuration in non-production environments first
4. **Document custom values** - Comment your `.env` file with explanations
5. **Rotate secrets regularly** - Update API keys and tokens periodically
6. **Use different keys per environment** - Never share production keys with development
7. **Monitor environment variables** - Log (without exposing values) which variables are loaded

## See Also

- [Configuration Guide](./configuration.md) - Complete configuration documentation
- [Security Best Practices](./security/best-practices.md) - Security guidelines
- [Deployment Guide](./deployment.md) - Production deployment
