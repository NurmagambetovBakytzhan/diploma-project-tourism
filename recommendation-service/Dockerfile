# Use official Python 3.10 image as base
FROM python:3.10-slim

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1

# Install system dependencies for psycopg2 and other potential requirements
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    libpq-dev \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy requirements first to leverage Docker cache
COPY requirements.txt .

# Upgrade pip to latest version
RUN pip install --upgrade pip

# Install Python dependencies
# Use psycopg2-binary as fallback if psycopg2 fails
RUN pip install --no-cache-dir -r requirements.txt || \
    (sed -i 's/psycopg2/psycopg2-binary/g' requirements.txt && \
     pip install --no-cache-dir -r requirements.txt)

# Copy the rest of the application
COPY . .

# Expose the port the app runs on
EXPOSE 8050

# Command to run the application
CMD ["uvicorn", "app.main:app", "--reload", "--host", "0.0.0.0", "--port", "8050"]