# Multi-stage Dockerfile for Sovereign Maps Backend
# Uses pytorch image to avoid long torch builds

# Stage 1: Build with PyTorch base
FROM pytorch/pytorch:2.1.0-runtime-slim as builder

WORKDIR /build

# Install build tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    gcc \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements (torch/torchvision already installed in base image)
COPY requirements.txt .

# Install remaining dependencies (exclude torch/torchvision since they're in base)
RUN pip install --no-cache-dir --user --no-warn-script-location \
    flwr==1.7.0 \
    flwr-datasets==0.4.0 \
    opacus==1.4.0 \
    ecdsa==0.18.0 \
    cryptography==41.0.7 \
    Flask==3.0.0 \
    Flask-CORS==4.0.0 \
    Flask-SocketIO==5.3.5 \
    python-socketio==5.10.0 \
    python-engineio==4.8.0 \
    numpy==1.24.3 \
    scipy==1.11.4 \
    scikit-learn==1.3.2 \
    pandas==2.1.3 \
    aiohttp==3.9.1 \
    google-generativeai==0.3.0 \
    google-auth==2.25.2 \
    redis==5.0.1 \
    prometheus-client==0.19.0 \
    prometheus-flask-exporter==0.23.0 \
    python-dotenv==1.0.0 \
    requests==2.31.0 \
    pydantic==2.5.0 \
    pydantic-settings==2.1.0 \
    pytest==7.4.3 \
    pytest-asyncio==0.21.1 \
    pytest-cov==4.1.0 \
    pytest-timeout==2.2.0

# Stage 2: Runtime
FROM pytorch/pytorch:2.1.0-runtime-slim

WORKDIR /app

# Copy Python dependencies from builder
COPY --from=builder /root/.local /root/.local

# Set PATH and Python vars
ENV PATH=/root/.local/bin:$PATH
ENV PYTHONUNBUFFERED=1
ENV PYTHONDONTWRITEBYTECODE=1

# Install curl for healthchecks
RUN apt-get update && apt-get install -y --no-install-recommends curl && rm -rf /var/lib/apt/lists/*

# Copy application code
COPY sovereignmap_production_backend_v2.py .
COPY src/ ./src/

# Health check
HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Expose ports
EXPOSE 8000 8080

# Run backend
CMD ["python", "sovereignmap_production_backend_v2.py"]
