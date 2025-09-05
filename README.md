# 🚀 **AI-HR System: Инструкция по развертыванию**## 📝 **Обзор****AI-HR System** — это микросервисный backend для автоматизации HR-процессов. Он включает в себя сервисы для аутентификации, управления вакансиями, резюме и интервью, используя PostgreSQL, RabbitMQ и MinIO.

## 🛠️ **Компоненты системы**- **`auth` (Auth Service):** Управляет аутентификацией пользователей и JWT-токенами.
- **`interview` (Interview Service):** Обрабатывает вакансии, резюме и интервью.
- **`db` (PostgreSQL):** Основное хранилище данных.
- **`rabbitmq` (RabbitMQ):** Брокер сообщений для асинхронной обработки задач.
- **`minio` (MinIO):** S3-совместимое хранилище для файлов.

***

## ⚙️ **Шаг 1: Настройка окружения**Создайте файл `.env` в корне проекта и заполните его по аналогии с `file.env`, предоставленным ранее:

```bash
# .env

# --- JWT ---
JWT_SECRET=your-very-long-and-secure-jwt-secret-key-at-least-32-characters-long
JWT_ACCESS_TTL=30m
JWT_REFRESH_TTL=7d

# --- Режим запуска ---
GIN_MODE=debug

# --- База данных (PostgreSQL) ---
DB_USER=ai_hr_user
DB_PASSWORD=ai_hr_password
DB_NAME=ai_hr_db

# --- MinIO S3 ---
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_BUCKET=interview-files

# --- URL для сервисов (для локальной разработки) ---
AUTH_SERVICE_URL=http://localhost:8080
INTERVIEW_SERVICE_URL=http://localhost:8081
RABBITMQ_URL=amqp://guest:guest@localhost```72/
```

## 🚀 **Шаг 2: Запуск системы**1.  **Сборка и запуск контейнеров:**
    Выполните команду в корневой директории проекта:
    ```bash
    docker compose up --build -d
    ```

2.  **Проверка статуса сервисов:**
    ```bash
    docker compose ps
    ```
    Все сервисы должны иметь статус `running (healthy)`.

***

## 🔗 **Доступы и порты**| Сервис | Локальный адрес | Пользователь | Пароль |
| :--- | :--- | :--- | :--- |
| **Auth Service API** | `http://localhost:8080` | - | - |
| **Interview Service API** | `http://localhost:8081` | - | - |
| **PostgreSQL** | `localhost:5432` | `ai_hr_user` | `ai_hr_password` |
| **RabbitMQ Management** | `http://localhost:15672` | `guest` | `guest` |
| **MinIO Console** | `http://localhost:9001` | `minioadmin` | `minioadmin123` |

***

## 🧪 **Шаг 3: Тестирование API**### **Аутентификация**Сначала получите JWT-токен через эндпоинты `auth` сервиса:
- `POST /auth/register` - для регистрации нового пользователя.
- `POST /auth/login` - для получения `access_token`.

### **Создание вакансии**Подставьте ваш `access_token` в заголовок `Authorization`.

```bash

curl -X POST http://localhost:8081/api/vacancies \
  -H "Authorization: Bearer "your_jwt_access_token_here"" \
  -F "title=Senior Backend Developer" \
  -F "description=Опытный Go-разработчик для микросервисов" \
  -F "users_id=1" \
  -F "weight_soft=30" \
  -F "weight_hard=50" \
  -F "weight_case=20" \
  -F "file=@/путь/к/вашему/файлу/test-vacancy.txt"
```
*Сохраните `id` вакансии из ответа.*

### **Загрузка резюме**```bash
VACANCY_ID="id_вакансии_из_предыдущего_шага"

curl -X POST http://localhost:8081/api/resumes \
-H "Authorization: Bearer $ACCESS_TOKEN" \
-F "vacancy_id=$VACANCY_ID" \
-F "file=@/путь/к/вашему/файлу/test-resume.txt"
```

***

## 📦 **Работа с брокером (RabbitMQ)**- **Отправка сообщений:** Сервис `interview` автоматически отправляет сообщение в очередь `resume_analysis_queue` при загрузке нового резюме.
- **Получение сообщений:** Можно запустить Python-консьюмер для их обработки.

1.  **Создайте и активируйте виртуальное окружение:**
    ```bash
    cd interview/internal/broker
    python3 -m venv venv
    source venv/bin/activate
    pip install pika
    ```

2.  **Запустите консьюмер:**
    ```bash
    python3 consumer.py
    ```
    Он будет ожидать и обрабатывать новые сообщения из очереди.

## 🗑️ **Остановка системы**Для остановки всех сервисов выполните:
```bash
docker compose down
```
Для удаления всех данных (включая volumes):
```bash
docker compose down -v
```

[1](https://ppl-ai-file-upload.s3.amazonaws.com/web/direct-files/attachments/101623268/5484ba4f-b4eb-4c25-b986-af7786885463/file.env)