# Database Heartbeat Monitor Dockerfile
FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
RUN pip install --no-cache-dir \
    requests \
    psycopg2-binary \
    pymongo

# Set working directory
WORKDIR /app

# Copy the heartbeat script
COPY run.py /app/run.py
RUN chmod +x /app/run.py

# Environment variables will be provided externally
ENV NEO4J_URL \
    NEO4J_USER \
    NEO4J_PASSWORD \
    POSTGRES_HOST \
    POSTGRES_PORT \
    POSTGRES_USER \
    POSTGRES_PASSWORD \
    POSTGRES_DB \
    MONGODB_HOST \
    MONGODB_PORT \
    MONGODB_USER \
    MONGODB_PASSWORD \
    MONGODB_DB \
    HEARTBEAT_INTERVAL

# Health check
HEALTHCHECK --interval=60s --timeout=10s --start-period=30s --retries=3 \
    CMD python /app/run.py --health-check || exit 1

# Run the heartbeat monitor
CMD ["python", "/app/run.py"]
