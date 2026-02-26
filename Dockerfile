# Multi-stage Dockerfile for Sovereign Maps Backend

# Stage 1: Builder
FROM python:3.11-slim as builder

WORKDIR /build

# Install build tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    gcc \
    curl \
    git \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY requirements.txt .

# Install dependencies
RUN pip install --no-cache-dir --user --no-warn-script-location \
    -r requirements.txt || pip install --no-cache-dir --user --no-warn-script-location \
    flwr==1.7.0 \
    torch==2.1.0 \
    torchvision==0.16.0 \
    opacus==1.4.0 \
    ecdsa==0.18.0 \
    cryptography==41.0.7 \
    Flask==3.0.0 \
    numpy==1.24.3 \
    prometheus-client==0.19.0 \
    prometheus-flask-exporter==0.23.0 \
    requests==2.31.0 \
    pydantic==2.5.0 \
    google-generativeai==0.3.0

# Stage 2: Runtime
FROM python:3.11-slim

WORKDIR /app

# Copy Python dependencies from builder
COPY --from=builder /root/.local /root/.local

# Set PATH and Python vars
ENV PATH=/root/.local/bin:$PATH
ENV PYTHONUNBUFFERED=1
ENV PYTHONDONTWRITEBYTECODE=1

# Install runtime dependencies only
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

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
