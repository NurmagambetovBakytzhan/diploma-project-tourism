# Use official Python 3.10 image as base
FROM python:3.10-slim as builder

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1
ENV PIP_NO_CACHE_DIR 1

# Install only essential system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    libpq-dev \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy only requirements first to leverage Docker cache
COPY requirements.txt .

# Install Python dependencies in one RUN layer
RUN pip install --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt || \
    (sed -i 's/psycopg2/psycopg2-binary/g' requirements.txt && \
     pip install --no-cache-dir -r requirements.txt)

# ---------------------------
# Production stage
FROM python:3.10-slim

# Copy only necessary files from builder
COPY --from=builder /usr/local/lib/python3.10/site-packages /usr/local/lib/python3.10/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# Install runtime dependencies only
RUN apt-get update && apt-get install -y --no-install-recommends \
    libpq5 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy application code (with .dockerignore to exclude unnecessary files)
COPY . .

# Expose the port the app runs on
EXPOSE 8050

# Command to run the application (remove --reload for production)
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8050"]