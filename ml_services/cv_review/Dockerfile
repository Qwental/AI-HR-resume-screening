FROM python:3.11-slim

# Установка системных зависимостей включая curl для healthcheck
RUN apt-get update && apt-get install -y \
    curl \
    gcc \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

WORKDIR /app

# Копируем и устанавливаем зависимости
COPY requirements.txt .
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

# Копируем код
COPY . .

# Создаем пользователя
RUN adduser --disabled-password --gecos '' appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 5030

# Улучшенная команда запуска
CMD ["uvicorn", "server:app", "--host", "0.0.0.0", "--port", "5030", "--workers", "1", "--access-log"]
