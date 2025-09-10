FROM python:3.9-slim

# Установка системных зависимостей
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Создание рабочей директории
WORKDIR /app


# Установка Python зависимостей
RUN pip install --no-cache-dir -r requirements.txt

# Копирование исходного кода
COPY consumer.py .

# Создание пользователя для безопасности
RUN useradd -m -u 1000 consumer
USER consumer

# Запуск consumer
CMD ["python", "consumer.py"]