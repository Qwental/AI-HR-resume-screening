#!/bin/bash

# Скрипт для тестирования работы токенов
BASE_URL="http://localhost:8080/api/v1"

echo "🧪 Тестирование системы токенов"
echo "================================"

# 1. Регистрация нового пользователя
echo "1️⃣ Регистрация пользователя..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "surname": "Test",
    "email": "test@example.com",
    "password": "123456"
  }')

echo "Ответ регистрации:"
echo "$REGISTER_RESPONSE" | jq .

# Извлекаем токены
ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.access_token')
REFRESH_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.refresh_token')

echo ""
echo "Access Token: ${ACCESS_TOKEN:0:50}..."
echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."

# 2. Проверяем защищенный endpoint
echo ""
echo "2️⃣ Проверка защищенного endpoint..."
PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Профиль пользователя:"
echo "$PROFILE_RESPONSE" | jq .

# 3. Обновляем токены
echo ""
echo "3️⃣ Обновление токенов..."
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}")

echo "Новые токены:"
echo "$REFRESH_RESPONSE" | jq .

# Извлекаем новые токены
NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.data.access_token')
NEW_REFRESH_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.data.refresh_token')

# 4. Тестируем новый токен
echo ""
echo "4️⃣ Проверка нового токена..."
NEW_PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/profile" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN")

echo "Профиль с новым токеном:"
echo "$NEW_PROFILE_RESPONSE" | jq .

# 5. Logout
echo ""
echo "5️⃣ Выход из системы..."
LOGOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/logout" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$NEW_REFRESH_TOKEN\"}")

echo "Logout результат:"
echo "$LOGOUT_RESPONSE" | jq .

# 6. Пытаемся использовать отозванный refresh токен
echo ""
echo "6️⃣ Попытка использовать отозванный токен..."
INVALID_REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$NEW_REFRESH_TOKEN\"}")

echo "Результат с отозванным токеном:"
echo "$INVALID_REFRESH_RESPONSE" | jq .

echo ""
echo "✅ Тестирование завершено!"
echo ""
echo "💡 Теперь проверьте БД:"
echo "   docker exec -it ai-hr-service-db psql -U postgres -d ai_hr_service_db -c 'SELECT * FROM tokens;'"