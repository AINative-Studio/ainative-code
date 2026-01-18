"""Main FastAPI application module.

This module initializes the FastAPI application with middleware,
routing, and configuration.
"""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.core.config import settings
from app.api.v1.endpoints.auth import router as auth_router, chat_router

# Initialize FastAPI application
app = FastAPI(
    title="AINative Code Backend",
    version=settings.API_VERSION,
    description="Authentication and Chat Completion Backend for AINative Code",
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
)

# Configure CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.ALLOWED_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include API routers
app.include_router(auth_router, prefix="/v1")
app.include_router(chat_router, prefix="/v1")


@app.get("/health")
async def health_check():
    """Health check endpoint.

    Returns:
        dict: Health status and API version
    """
    return {"status": "healthy", "version": settings.API_VERSION}


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        app,
        host="0.0.0.0",
        port=8000,
        reload=settings.DEBUG,
    )
