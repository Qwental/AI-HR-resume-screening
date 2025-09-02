# Zalupa


## запуск

```bash
# Убедись что Docker запущен (иначе будет ошибка)
docker-compose up --build
```
или если что-то наебнулось
```bash
docker-compose down                 
docker-compose build --no-cache
docker-compose up

```



## Что где висит

- **морда**: http://localhost:8080
- **База данных**: localhost:5433 (не 5432, потому что у кого-то уже занят)

## Страницы

- `/` - главная с кнопочками
- `/login` - вход
- `/register` - регистрация
- `/dashboard` - типа дашборд после входа

## API эндпоинты

### Публичные (без токена)
- `POST /api/auth/register` - регистрация
- `POST /api/auth/login` - авторизация
- `GET /health` - проверить что живо

### Приватные (нужен токен в хедере)
- `GET /api/profile` - профиль юзера
- `GET /api/protected` - тестовый защищенный
- `GET /api/hr/dashboard` - только для HR

## База данных

- **Хост**: localhost
- **Порт**: 5433
- **База**: ai_hr_service_db
- **Юзер/пароль**: postgres/password

Подключаться в GoLand Database на порт 5433, не 5432.

## Если что-то не работает

```bash
# Посмотреть что происходит
docker-compose logs

# Если совсем все сломалось
docker-compose down -v
docker system prune -f
docker-compose up --build

# Зайти в контейнер и покопаться
docker-compose exec app sh
docker-compose exec db psql -U postgres -d auth_demo
```

# Создать БД
### Подключитесь к работающему контейнеру PostgreSQL

```
docker-compose exec db psql -U postgres

# В psql выполните:
CREATE DATABASE ai_hr_service_db;

# Проверьте что база создалась:
\l

# Выйдите из psql:
\q

# Перезапустите приложение
docker-compose restart app

# Проверьте логи
docker-compose logs -f app
```

## Структура проекта

```
/cmd/server/main.go     - точка входа
/internal/auth/         - вся логика аутентификации  
/internal/config/       - настройки из .env
/internal/database/     - подключение к базе
/internal/middleware/   - проверка токенов и ролей
/static/               - HTML страницы
```

## Переменные окружения

В `.env` файле:
- `DB_HOST` - хост базы (в Docker это `db`)
- `DB_PORT` - порт базы
- `JWT_SECRET` - секрет для токенов (поменяй в продакшене)
- `PORT` - порт приложения

## Тестирование

```bash
# Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","surname":"user","email":"test@test.com","password":"123456"}'

# Вход  
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"123456"}'

# Сохрани токен из ответа и используй для защищенных запросов
curl -H "Authorization: Bearer ТУТ_ТОКЕН" http://localhost:8080/api/profile
```

=аботает как надо.