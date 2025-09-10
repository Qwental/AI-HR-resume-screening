# ИИ-скрининг резюме

Интеллектуальная система для автоматизации процесса скрининга кандидатов с использованием искусственного интеллекта.


ИИ-скрининг резюме — это современная HR-платформа, которая позволяет:
- Автоматически анализировать резюме кандидатов
- Проводить голосовые интервью с ИИ-аватаром
- Генерировать детальные отчеты по результатам собеседований
- Управлять вакансиями и кандидатами
- Получать аналитику по процессам найма

# Технологический стек

### Backend
- **Go** - основной язык разработки
- **Gin** - HTTP фреймворк
- **PostgreSQL** - основная база данных
- **JWT** - аутентификация и авторизация
- **Docker** - контейнеризация
- **Nginx** - контейнеризация

### Frontend
- **Next.js** - React фреймворк с SSR/SSG
- **TailwindCSS** - utility-first CSS фреймворк
- **Zustand** - управление состоянием
- **Recharts** - графики и аналитика
- **React Dropzone** - загрузка файлов


## Архитектура проекта

```
ai-hr-system/
├── auth/                    # Auth Service 
│   ├── cmd/server/
│   ├── internal/
│   ├── Dockerfile
│   └── docker-compose.yml
├── interview/               # Interview Service 
│   ├── cmd/
│   ├── internal/
│   └── migrations/
├── frontend/                
│   ├── pages/
│   ├── components/
│   └── utils/          
└── docker-compose.yml       
```

## Быстрый старт

### Предварительные требования
- Docker и Docker Compose
- Go 1.21+ (для разработки)
- Node.js 18+ (для разработки фронтенда)

### 1. Клонирование репозитория

### 2. Настройка переменных окружения

Создайте файл `.env` в корне проекта:
```env
# База данных
DB_HOST=postgres
DB_PORT=5432
DB_NAME=ai_hr_db
DB_USER=postgres
DB_PASSWORD=postgres123

# Auth Service
AUTH_PORT=8080
JWT_SECRET=your-super-secret-jwt-key-here-min-32-chars
JWT_REFRESH_SECRET=your-refresh-secret-key-here-min-32-chars
JWT_EXPIRES_IN=15m
JWT_REFRESH_EXPIRES_IN=7d

# Interview Service  
INTERVIEW_PORT=8081
INTERVIEW_DB_DSN=postgres://postgres:postgres123@postgres:5432/ai_hr_db?sslmode=disable

# S3 Storage (для файлов резюме)
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-east-1
S3_BUCKET_NAME=ai-hr-resumes

# Frontend
NEXT_PUBLIC_AUTH_API_URL=http://localhost:8080
NEXT_PUBLIC_INTERVIEW_API_URL=http://localhost:8081


# Логирование
LOG_LEVEL=debug
LOG_FORMAT=json
```




### 3. Запуск через Docker Compose

```bash
# Запуск всех сервисов
docker-compose up -d

# Просмотр логов
docker-compose logs -f

# Остановка сервисов
docker-compose down
```

### 4. Ручной запуск для разработки

```bash
cd frontend
npm install
npm run dev
```

## 📱 Использование

После успешного запуска:
- http://localhost






![скрин](1.jpg)
![скрин](2.jpg)
![скрин](3.jpg)
![скрин](4.jpg)
![скрин](5.jpg)
![скрин](6.jpg)
![скрин](7.jpg)

