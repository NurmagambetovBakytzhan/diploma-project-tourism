FROM python:3.12-slim

WORKDIR /app

# Системные зависимости
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    python3-dev \
    libpq-dev \
    build-essential \
    postgresql-client \
 && apt-get clean && rm -rf /var/lib/apt/lists/*
# Обновление pip
RUN pip install --upgrade pip setuptools wheel

# Копируем heavy зависимости и устанавливаем из локальной папки
COPY requirements.txt .
COPY wheels /wheels
RUN pip install --no-index --find-links=/wheels -r requirements.txt


# Копируем приложение
COPY . .

# Запуск приложения
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8050"]
