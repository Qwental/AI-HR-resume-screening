#!/bin/bash

# Правильный скрипт для тестирования сервиса интервью
# Учитывает реальные эндпоинты и форматы данных

set -e

# Настройки
BASE_URL="http://localhost:8081"
API_URL="$BASE_URL/api"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}🚀 Тестирование сервиса интервью (ИСПРАВЛЕННАЯ ВЕРСИЯ)${NC}"
echo "=================================================="

# Проверка доступности сервиса
echo -e "${YELLOW}🔍 Проверка доступности сервиса...${NC}"
if ! curl -s --connect-timeout 5 "$BASE_URL" > /dev/null; then
    echo -e "${RED}❌ Сервис недоступен на $BASE_URL${NC}"
    echo "Убедитесь, что сервис запущен: docker-compose up -d"
    exit 1
fi
echo -e "${GREEN}✅ Сервис доступен${NC}"

# Создание тестовых файлов
echo -e "${YELLOW}📄 Создание тестовых файлов...${NC}"

# Файл вакансии
cat > /tmp/test_vacancy.txt << 'EOF'
Статус: Открыта
Название: Senior Go Developer
Регион: Москва
Город: Москва
Адрес: Красная площадь, 1
Тип трудового: Постоянно
Тип занятости: Полная занятость
Текст график работы: Гибридный формат
Доход (руб/мес): 200000-300000
Оклад макс. (руб/мес): 300000
Оклад мин. (руб/мес): 200000
Годовая премия (%): 20
Тип премирования. Описание: По результатам года
Обязанности (для публикации): Разработка на Go, проектирование архитектуры, работа с PostgreSQL и Redis
Требования (для публикации): Опыт Go 3+ лет, PostgreSQL, Docker, Kubernetes
Уровень образования: Высшее
Требуемый опыт работы: 3-5 лет
Знание специальных программ: Go, PostgreSQL, Docker
Навыки работы на компьютере: Эксперт
Знание иностранных языков: Английский
Уровень владения языка: Intermediate
Наличие командировок: Нет
Дополнительная информация: Отличная команда
EOF

# Файл резюме
cat > /tmp/test_resume.txt << 'EOF'
РЕЗЮМЕ

ФИО: Петров Иван Александрович
Email: ivan.petrov@email.com
Телефон: +7-999-123-4567

ОПЫТ РАБОТЫ:
2022-2024: Go Developer в ООО "ТехКорп"
- Разработка REST API на Go
- Работа с PostgreSQL, Redis
- Docker контейнеризация
- Unit тестирование

ОБРАЗОВАНИЕ:
2018-2022: МГТУ им. Баумана
Программная инженерия

НАВЫКИ:
- Go (Gin, GORM)
- PostgreSQL, Redis
- Docker, Git
- REST API

ПРОЕКТЫ:
- API для мобильного приложения
- Система управления пользователями
- Микросервис платежей

Готов к обучению и работе в команде.
EOF

echo -e "${GREEN}✅ Тестовые файлы созданы${NC}"

# 1. Создание вакансии (FORM-DATA с файлом)
echo -e "\n${BLUE}1️⃣  Создание вакансии${NC}"
echo "=========================================="

VACANCY_RESPONSE=$(curl -s -X POST "http://localhost:8081/vacancies" \
  -F "users_id=1" \
  -F "title=Senior Go Developer" \
  -F "description=Разработчик Go для стартапа" \
  -F "weight_soft=25" \
  -F "weight_hard=50" \
  -F "weight_case=25" \
  -F "file=@test_vacancy.txt")

echo -e "${YELLOW}Ответ сервера:${NC}"
echo "$VACANCY_RESPONSE" | jq '.' 2>/dev/null || echo "$VACANCY_RESPONSE"

# Извлекаем ID вакансии
VACANCY_ID=$(echo "$VACANCY_RESPONSE" | jq -r '.id' 2>/dev/null)

if [ "$VACANCY_ID" = "null" ] || [ -z "$VACANCY_ID" ]; then
    echo -e "${RED}❌ Не удалось создать вакансию${NC}"
    echo "Ответ: $VACANCY_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✅ Вакансия создана с ID: $VACANCY_ID${NC}"

# 2. Получение вакансии по ID (GET с ID в URL)
echo -e "\n${BLUE}2️⃣  Получение вакансии по ID${NC}"
echo "=========================================="

VACANCY_DATA=$(curl -s -X GET "$API_URL/vacancies/$VACANCY_ID")
echo -e "${YELLOW}Данные вакансии:${NC}"
echo "$VACANCY_DATA" | jq '.' 2>/dev/null || echo "$VACANCY_DATA"

# 3. Получение всех вакансий
echo -e "\n${BLUE}3️⃣  Получение всех вакансий${NC}"
echo "=========================================="

ALL_VACANCIES=$(curl -s -X GET "$API_URL/vacancies")
echo -e "${YELLOW}Все вакансии:${NC}"
echo "$ALL_VACANCIES" | jq '.' 2>/dev/null || echo "$ALL_VACANCIES"

# 4. Создание резюме (FORM-DATA с файлом)
echo -e "\n${BLUE}4️⃣  Создание резюме${NC}"
echo "=========================================="

RESUME_RESPONSE=$(curl -s -X POST "$API_URL/resumes" \
  -F "vacancy_id=$VACANCY_ID" \
  -F "file=@/tmp/test_resume.txt")

echo -e "${YELLOW}Ответ сервера:${NC}"
echo "$RESUME_RESPONSE" | jq '.' 2>/dev/null || echo "$RESUME_RESPONSE"

# Извлекаем ID резюме
RESUME_ID=$(echo "$RESUME_RESPONSE" | jq -r '.id' 2>/dev/null)

if [ "$RESUME_ID" = "null" ] || [ -z "$RESUME_ID" ]; then
    echo -e "${RED}❌ Не удалось создать резюме${NC}"
    echo "Ответ: $RESUME_RESPONSE"
else
    echo -e "${GREEN}✅ Резюме создано с ID: $RESUME_ID${NC}"
fi

# 5. Получение резюме по ID (GET с ID в URL)
if [ "$RESUME_ID" != "null" ] && [ -n "$RESUME_ID" ]; then
    echo -e "\n${BLUE}5️⃣  Получение резюме по ID${NC}"
    echo "=========================================="

    RESUME_DATA=$(curl -s -X GET "$API_URL/resumes/$RESUME_ID")
    echo -e "${YELLOW}Данные резюме:${NC}"
    echo "$RESUME_DATA" | jq '.' 2>/dev/null || echo "$RESUME_DATA"
fi

# 6. Получение резюме для вакансии (исправлен URL)
echo -e "\n${BLUE}6️⃣  Получение резюме для вакансии${NC}"
echo "=========================================="

# Правильный URL для получения резюме по вакансии
VACANCY_RESUMES=$(curl -s -X GET "$API_URL/vacancies/$VACANCY_ID/resumes")
echo -e "${YELLOW}Резюме для вакансии:${NC}"
echo "$VACANCY_RESUMES" | jq '.' 2>/dev/null || echo "$VACANCY_RESUMES"

# 7. Создание интервью (JSON без файла)
echo -e "\n${BLUE}7️⃣  Создание интервью${NC}"
echo "=========================================="

# Формируем JSON
INTERVIEW_JSON="{\"vacancy_id\": \"$VACANCY_ID\""
if [ "$RESUME_ID" != "null" ] && [ -n "$RESUME_ID" ]; then
    INTERVIEW_JSON="$INTERVIEW_JSON, \"resume_id\": \"$RESUME_ID\""
fi
INTERVIEW_JSON="$INTERVIEW_JSON}"

INTERVIEW_RESPONSE=$(curl -s -X POST "$API_URL/admin/interviews" \
  -H "Content-Type: application/json" \
  -d "$INTERVIEW_JSON")

echo -e "${YELLOW}Ответ сервера:${NC}"
echo "$INTERVIEW_RESPONSE" | jq '.' 2>/dev/null || echo "$INTERVIEW_RESPONSE"

# Извлекаем данные интервью
INTERVIEW_ID=$(echo "$INTERVIEW_RESPONSE" | jq -r '.id' 2>/dev/null)
INTERVIEW_URL=$(echo "$INTERVIEW_RESPONSE" | jq -r '.interview_url' 2>/dev/null)
INTERVIEW_TOKEN=""

if [ "$INTERVIEW_URL" != "null" ] && [ -n "$INTERVIEW_URL" ]; then
    INTERVIEW_TOKEN=$(echo "$INTERVIEW_URL" | sed 's|.*/interview/||')
    echo -e "${GREEN}✅ Интервью создано с ID: $INTERVIEW_ID${NC}"
    echo -e "${GREEN}✅ Токен интервью: $INTERVIEW_TOKEN${NC}"
    echo -e "${GREEN}✅ URL интервью: $INTERVIEW_URL${NC}"

    # 8. Тестирование интервью
    echo -e "\n${BLUE}8️⃣  Тестирование интервью${NC}"
    echo "=========================================="

    # Получение статуса интервью (GET с токеном в URL)
    echo -e "${YELLOW}Получение статуса интервью:${NC}"
    INTERVIEW_STATUS=$(curl -s -X GET "$BASE_URL/interview/$INTERVIEW_TOKEN")
    echo "$INTERVIEW_STATUS" | jq '.' 2>/dev/null || echo "$INTERVIEW_STATUS"

    # Запуск интервью
    echo -e "\n${YELLOW}Запуск интервью:${NC}"
    START_RESPONSE=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/interview/$INTERVIEW_TOKEN/start")
    if [[ "$START_RESPONSE" == *"200"* ]]; then
        echo -e "${GREEN}✅ Интервью запущено${NC}"
    else
        echo -e "${YELLOW}Ответ запуска: $START_RESPONSE${NC}"
    fi

    # Статус после запуска
    echo -e "\n${YELLOW}Статус после запуска:${NC}"
    UPDATED_STATUS=$(curl -s -X GET "$BASE_URL/interview/$INTERVIEW_TOKEN")
    echo "$UPDATED_STATUS" | jq '.' 2>/dev/null || echo "$UPDATED_STATUS"

    # Завершение интервью
    echo -e "\n${YELLOW}Завершение интервью:${NC}"
    FINISH_RESPONSE=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/interview/$INTERVIEW_TOKEN/finish")
    if [[ "$FINISH_RESPONSE" == *"200"* ]]; then
        echo -e "${GREEN}✅ Интервью завершено${NC}"
    else
        echo -e "${YELLOW}Ответ завершения: $FINISH_RESPONSE${NC}"
    fi
else
    echo -e "${RED}❌ Не удалось создать интервью${NC}"
fi

# Очистка
echo -e "\n${BLUE}🧹 Очистка${NC}"
echo "=========================================="
rm -f /tmp/test_vacancy.txt /tmp/test_resume.txt
echo -e "${GREEN}✅ Временные файлы удалены${NC}"

# Итоги
echo -e "\n${BLUE}📊 ИТОГИ ТЕСТИРОВАНИЯ${NC}"
echo "=================================================="
echo -e "${GREEN}✅ Вакансия ID: $VACANCY_ID${NC}"
if [ "$RESUME_ID" != "null" ] && [ -n "$RESUME_ID" ]; then
    echo -e "${GREEN}✅ Резюме ID: $RESUME_ID${NC}"
fi
if [ "$INTERVIEW_ID" != "null" ] && [ -n "$INTERVIEW_ID" ]; then
    echo -e "${GREEN}✅ Интервью ID: $INTERVIEW_ID${NC}"
    echo -e "${GREEN}✅ Интервью токен: $INTERVIEW_TOKEN${NC}"
    echo -e "${BLUE}🔗 Ссылка на интервью: $INTERVIEW_URL${NC}"
fi

echo -e "\n${YELLOW}💡 Полезные команды для дальнейшего тестирования:${NC}"
echo "# Получить все вакансии:"
echo "curl -X GET $API_URL/vacancies | jq ."
echo ""
echo "# Получить конкретную вакансию:"
echo "curl -X GET $API_URL/vacancies/$VACANCY_ID | jq ."
echo ""
if [ "$RESUME_ID" != "null" ] && [ -n "$RESUME_ID" ]; then
    echo "# Получить резюме:"
    echo "curl -X GET $API_URL/resumes/$RESUME_ID | jq ."
    echo ""
    echo "# Обновить статус резюме:"
    echo "curl -X PUT $API_URL/resumes/$RESUME_ID/status -H 'Content-Type: application/json' -d '{\"status\": \"approved\"}'"
    echo ""
fi

echo -e "\n${GREEN}🎉 Тестирование завершено!${NC}"
