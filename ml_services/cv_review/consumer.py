# ml_services/cv_review/consumer.py
import json
import logging
import os
import time
from datetime import datetime

import pika
import psycopg2
import psycopg2.extras
import requests
from rich.console import Console
from rich.json import JSON
from rich.panel import Panel
from rich.table import Table

# Настройки подключения
RABBITMQ_HOST = os.getenv("RABBITMQ_HOST", "localhost")
RABBITMQ_URL = f"amqp://guest:guest@{RABBITMQ_HOST}:5672/"
CV_REVIEW_SERVICE = os.getenv("CV_REVIEW_SERVICE", "http://localhost:5030")

# Настройки БД
DB_HOST = os.getenv("DB_HOST", "localhost")
DB_USER = os.getenv("DB_USER", "postgres")
DB_PASSWORD = os.getenv("DB_PASSWORD", "postgres")
DB_NAME = os.getenv("DB_NAME", "aihrservicedb")
DB_PORT = os.getenv("DB_PORT", "5432")

EXCHANGE_NAME = "resume_exchange"
QUEUE_NAME = "resume_analysis_queue"
ROUTING_KEY = "resume_analysis_queue"

console = Console()
logger = logging.getLogger(__name__)


def get_db_connection():
    """Получение подключения к PostgreSQL"""
    try:
        connection = psycopg2.connect(
            host=DB_HOST,
            port=DB_PORT,
            database=DB_NAME,
            user=DB_USER,
            password=DB_PASSWORD
        )
        return connection
    except Exception as e:
        console.print(f"💥 [bold red]Ошибка подключения к БД:[/bold red] {e}")
        return None


def extract_email_from_ai_response(ai_response):
    """
    Извлечение email из ответа ИИ
    :param ai_response: Dict с ответом от ИИ сервиса
    :return: Email или 'user_has_no_mail'
    """
    try:
        console.print(f"🔍 [bold cyan]Извлечение email из ответа ИИ...[/bold cyan]")

        # Проверяем тип входных данных
        if not isinstance(ai_response, dict):
            console.print(f"⚠️ [bold yellow]Ожидался словарь, получен: {type(ai_response)}[/bold yellow]")
            return 'user_has_no_mail'

        # Получаем поле email
        email_field = ai_response.get('email', [])
        console.print(f"📧 Поле email: {email_field} (тип: {type(email_field)})")

        # Обрабатываем разные случаи
        if isinstance(email_field, list):
            if len(email_field) > 0:
                # Есть email в списке
                first_email = email_field[0]
                if isinstance(first_email, str) and first_email.strip():
                    extracted_email = first_email.strip()
                    console.print(f"✅ [bold green]Найден email: '{extracted_email}'[/bold green]")
                    return extracted_email
                else:
                    console.print(f"⚠️ [bold yellow]Первый элемент списка пустой[/bold yellow]")
            else:
                console.print(f"⚠️ [bold yellow]Список email пуст[/bold yellow]")
        elif isinstance(email_field, str):
            if email_field.strip():
                extracted_email = email_field.strip()
                console.print(f"✅ [bold green]Найден email (строка): '{extracted_email}'[/bold green]")
                return extracted_email
            else:
                console.print(f"⚠️ [bold yellow]Email строка пустая[/bold yellow]")
        else:
            console.print(f"⚠️ [bold yellow]Email поле неожиданного типа: {type(email_field)}[/bold yellow]")

        console.print(f"❌ [bold red]Email не найден, используем 'user_has_no_mail'[/bold red]")
        return 'user_has_no_mail'

    except Exception as e:
        console.print(f"💥 [bold red]Ошибка извлечения email:[/bold red] {e}")
        return 'user_has_no_mail'

def save_analysis_to_database(resume_id, ai_response, extracted_email):
    """
    Сохранение результатов анализа в БД
    :param resume_id: ID резюме (UUID)
    :param ai_response: Dict с полным ответом от ИИ
    :param extracted_email: Найденный email или 'user_has_no_mail'
    :return: True если успешно, False если ошибка
    """
    connection = get_db_connection()
    if not connection:
        console.print(f"❌ [bold red]Не удалось подключиться к БД[/bold red]")
        return False

    try:
        cursor = connection.cursor()

        # Преобразуем весь ответ ИИ в JSON для сохранения
        analysis_json = json.dumps(ai_response, ensure_ascii=False)

        # Определяем значение email для БД
        email_for_db = None if extracted_email == 'user_has_no_mail' else extracted_email

        console.print(f"💾 [bold cyan]Сохранение в БД:[/bold cyan]")
        console.print(f"📋 Resume ID: {resume_id}")
        console.print(f"📧 Email для БД: {email_for_db}")
        console.print(f"📊 Размер анализа: {len(analysis_json)} символов")

        # SQL запрос для обновления резюме
        update_query = """
            UPDATE resumes 
            SET 
                resume_analysis_jsonb = %s::jsonb,
                mail = %s,
                status = 'analyzed'
            WHERE id = %s
        """

        cursor.execute(update_query, (analysis_json, email_for_db, resume_id))
        connection.commit()

        if cursor.rowcount > 0:
            console.print(f"✅ [bold green]Резюме {resume_id} успешно обновлено в БД[/bold green]")

            # Проверяем что реально сохранилось
            cursor.execute(
                "SELECT mail, status FROM resumes WHERE id = %s",
                (resume_id,)
            )
            result = cursor.fetchone()
            if result:
                saved_mail, saved_status = result
                console.print(f"🔍 [bold cyan]Проверка: mail='{saved_mail}', status='{saved_status}'[/bold cyan]")

            return True
        else:
            console.print(f"⚠️ [bold yellow]Резюме {resume_id} не найдено в БД для обновления[/bold yellow]")
            return False

    except Exception as e:
        console.print(f"💥 [bold red]Ошибка сохранения в БД:[/bold red] {e}")
        import traceback
        console.print(traceback.format_exc())
        connection.rollback()
        return False
    finally:
        cursor.close()
        connection.close()


def log_message_received(message_data):
    """Логирование полученного сообщения"""
    console.print("\n" + "=" * 60, style="bold blue")
    console.print("📥 ПОЛУЧЕНО НОВОЕ СООБЩЕНИЕ", style="bold green", justify="center")
    console.print("=" * 60, style="bold blue")

    # Основная информация
    info_table = Table(title="📋 Информация о задаче", show_header=True)
    info_table.add_column("Параметр", style="cyan", no_wrap=True)
    info_table.add_column("Значение", style="white")

    resume_id = message_data.get('id', 'N/A')
    vacancy_id = message_data.get('vacancy_id', 'N/A')

    info_table.add_row("🆔 ID резюме", resume_id)
    info_table.add_row("📋 ID вакансии", vacancy_id)
    info_table.add_row("⏰ Время получения", datetime.now().strftime("%H:%M:%S"))

    # Веса навыков
    weight_soft = message_data.get('weight_soft', 33)
    weight_hard = message_data.get('weight_hard', 33)
    weight_case = message_data.get('weight_case', 34)
    total_weight = weight_soft + weight_hard + weight_case

    info_table.add_row("💪 Hard Skills", f"{weight_hard}%")
    info_table.add_row("🤝 Soft Skills", f"{weight_soft}%")
    info_table.add_row("💼 Опыт/Кейсы", f"{weight_case}%")

    # Проверка корректности весов
    if total_weight == 100:
        info_table.add_row("✅ Сумма весов", f"{total_weight}% (корректно)")
    else:
        info_table.add_row("⚠️ Сумма весов", f"{total_weight}% (должно быть 100%)")

    console.print(info_table)

    # Размеры данных
    resume_text = message_data.get('text_resume_jsonb', {})
    vacancy_text = message_data.get('text_vacancy_jsonb', {})

    cv_text = resume_text.get('text', '') if resume_text else ''
    vacancy_data = vacancy_text.get('structured_data', {}) if vacancy_text else {}

    size_table = Table(title="📊 Размер данных")
    size_table.add_column("Тип данных", style="cyan")
    size_table.add_column("Размер", style="green")

    size_table.add_row("📄 Текст резюме", f"{len(cv_text)} символов")
    size_table.add_row("📋 Данные вакансии", f"{len(str(vacancy_data))} символов")

    console.print(size_table)

    return resume_id, cv_text, vacancy_data, weight_soft, weight_hard, weight_case


def log_ai_request_details(resume_id, review_request, skillvals):
    """Подробное логирование запроса к ИИ"""
    console.print(f"\n🚀 [bold yellow]ДЕТАЛИ ЗАПРОСА К ИИ СЕРВИСУ[/bold yellow]")
    console.print("=" * 80, style="bold yellow")

    # Основная информация о запросе
    console.print(f"📋 Резюме ID: [bold cyan]{resume_id}[/bold cyan]")
    console.print(f"🤖 Сервис ИИ: [bold cyan]{CV_REVIEW_SERVICE}[/bold cyan]")
    console.print(f"⚖️ Веса навыков: [bold cyan]{skillvals}[/bold cyan]")

    # Детали запроса
    console.print("\n📦 [bold green]СОДЕРЖИМОЕ ЗАПРОСА:[/bold green]")

    # Данные вакансии
    vacancy_json = review_request.get('vacancy', '')
    try:
        vacancy_parsed = json.loads(vacancy_json) if vacancy_json else {}
        vacancy_panel = Panel(
            JSON(vacancy_parsed, indent=2),  # ИСПРАВЛЕНО
            title="📋 Данные вакансии (vacancy)",
            border_style="blue",
            expand=False
        )
        console.print(vacancy_panel)
    except Exception as e:
        console.print(f"❌ Ошибка парсинга vacancy: {e}")
        console.print(f"📋 Данные вакансии (raw): {vacancy_json[:200]}...")

    # Текст резюме
    cv_text = review_request.get('cv', '')
    cv_preview = cv_text[:300] + "..." if len(cv_text) > 300 else cv_text
    cv_panel = Panel(
        cv_preview,
        title=f"📄 Текст резюме ({len(cv_text)} символов)",
        border_style="green",
        expand=False
    )
    console.print(cv_panel)

    # Параметры запроса
    params_table = Table(title="🔧 Параметры запроса")
    params_table.add_column("Параметр", style="cyan")
    params_table.add_column("Значение", style="white")

    for key, value in review_request.items():
        if key == 'cv':
            params_table.add_row(key, f"{len(str(value))} символов")
        elif key == 'vacancy':
            params_table.add_row(key, f"{len(str(value))} символов JSON")
        else:
            params_table.add_row(key, str(value))

    console.print(params_table)

def log_ai_response_details(resume_id, response, processing_time):
    """Подробное логирование ответа от ИИ"""
    console.print(f"\n📨 [bold green]ОТВЕТ ОТ ИИ СЕРВИСА[/bold green]")
    console.print("=" * 80, style="bold green")

    # Информация о ответе
    response_table = Table(title="📊 Информация об ответе")
    response_table.add_column("Параметр", style="cyan")
    response_table.add_column("Значение", style="white")

    response_table.add_row("📋 Резюме ID", str(resume_id))
    response_table.add_row("🔢 HTTP статус", str(response.status_code))
    response_table.add_row("⏱️ Время обработки", f"{processing_time}с")
    response_table.add_row("📏 Размер ответа", f"{len(response.text)} символов")

    console.print(response_table)

    if response.status_code == 200:
        console.print("\n✅ [bold green]УСПЕШНЫЙ ОТВЕТ:[/bold green]")

        try:
            # Пытаемся распарсить JSON ответ
            analysis_result = response.json()

            # Показываем JSON в красивом формате
            result_panel = Panel(
                JSON(analysis_result, indent=2),
                title="🎯 Результат анализа (JSON)",
                border_style="green",
                expand=True
            )
            console.print(result_panel)

            # ВАЖНО: Убедимся, что возвращаем словарь, а не строку
            return analysis_result  # Это должен быть словарь

        except json.JSONDecodeError:
            # Если не JSON, показываем как текст
            text_result = response.text
            result_panel = Panel(
                text_result,
                title="🎯 Результат анализа (Текст)",
                border_style="green",
                expand=True
            )
            console.print(result_panel)

            # Возвращаем как строку, но это не то, что нам нужно
            console.print(f"⚠️ [bold yellow]ВНИМАНИЕ: Ответ не в JSON формате![/bold yellow]")
            return text_result
    else:
        console.print(f"\n❌ [bold red]ОШИБКА ОТВЕТА:[/bold red]")

        error_panel = Panel(
            response.text,
            title=f"💥 HTTP {response.status_code} - Текст ошибки",
            border_style="red",
            expand=True
        )
        console.print(error_panel)

        return None

def process_resume_message(body):
    """Обработка сообщения из RabbitMQ с подробным логированием и сохранением в БД"""
    try:
        message_data = json.loads(body)

        # 📥 Логируем получение сообщения
        resume_id, cv_text, vacancy_data, weight_soft, weight_hard, weight_case = log_message_received(message_data)

        # 🎯 Формируем skillvals динамически
        skillvals = f"hard_skills:{weight_hard},soft_skills:{weight_soft},experience:{weight_case}"

        # Формируем запрос к CV Review сервису
        review_request = {
            "vacancy": json.dumps(vacancy_data),
            "cv": cv_text,
            "skillvals": skillvals
        }

        # 🚀 Логируем детали запроса к ИИ
        log_ai_request_details(resume_id, review_request, skillvals)

        # Отправляем запрос к FastAPI сервису
        console.print(f"\n🌐 [bold yellow]Отправка HTTP GET запроса...[/bold yellow]")
        start_time = time.time()

        response = requests.get(
            f"{CV_REVIEW_SERVICE}/get_review",
            params=review_request,
            timeout=120  # Увеличиваем timeout для ИИ
        )

        end_time = time.time()
        processing_time = round(end_time - start_time, 2)

        # 📊 Логируем детали ответа
        analysis_result = log_ai_response_details(resume_id, response, processing_time)

        if response.status_code == 200 and analysis_result is not None:
            # ДОБАВИМ ПРОВЕРКУ ТИПА
            if isinstance(analysis_result, str):
                console.print(f"⚠️ [bold yellow]ВНИМАНИЕ: Ответ ИИ в строковом формате, пытаемся распарсить...[/bold yellow]")
                try:
                    analysis_result = json.loads(analysis_result)
                    console.print(f"✅ [bold green]Строка успешно преобразована в JSON[/bold green]")
                except json.JSONDecodeError:
                    console.print(f"❌ [bold red]Не удалось преобразовать строку в JSON[/bold red]")
                    # Создаем минимальный результат с ошибкой
                    analysis_result = {
                        "error": "invalid_json_format",
                        "raw_response": analysis_result[:500] + "..." if len(analysis_result) > 500 else analysis_result
                    }

            # 📧 Извлекаем email из ответа ИИ (теперь analysis_result точно словарь)
            extracted_email = extract_email_from_ai_response(analysis_result)
            console.print(f"🎯 [bold blue]Финальный извлеченный email: '{extracted_email}'[/bold blue]")

            # 💾 Сохраняем результаты в БД
            db_success = save_analysis_to_database(resume_id, analysis_result, extracted_email)

            if db_success:
                # 🎉 Полный успех
                console.print("\n" + "🎉 " * 30)
                console.print(f"[bold green]✅ РЕЗЮМЕ {resume_id} ПОЛНОСТЬЮ ОБРАБОТАНО И СОХРАНЕНО![/bold green]",
                              justify="center")
                console.print(f"[bold green]📧 EMAIL: {extracted_email}[/bold green]", justify="center")
                console.print(f"[bold green]💾 АНАЛИЗ СОХРАНЕН В БД[/bold green]", justify="center")
                console.print("🎉 " * 30)
            else:
                # ⚠️ Анализ прошел, но БД не обновилась
                console.print("\n" + "⚠️ " * 30)
                console.print(
                    f"[bold yellow]⚠️ РЕЗЮМЕ {resume_id} ПРОАНАЛИЗИРОВАНО, НО НЕ СОХРАНЕНО В БД[/bold yellow]",
                    justify="center")
                console.print("⚠️ " * 30)

        else:
            console.print("\n" + "💥 " * 30)
            console.print(f"[bold red]❌ ОШИБКА ОБРАБОТКИ РЕЗЮМЕ {resume_id}![/bold red]", justify="center")
            console.print("💥 " * 30)

    except json.JSONDecodeError as e:
        console.print(f"💥 [bold red]Ошибка парсинга JSON из RabbitMQ:[/bold red] {e}")
    except requests.exceptions.Timeout:
        console.print(f"⏰ [bold red]Таймаут запроса к ИИ сервису![/bold red]")
    except requests.exceptions.ConnectionError:
        console.print(f"🔌 [bold red]Ошибка подключения к ИИ сервису![/bold red]")
    except Exception as e:
        console.print(f"💥 [bold red]Неожиданная ошибка в process_resume_message:[/bold red] {e}")
        import traceback
        console.print(traceback.format_exc())


def print_startup_banner():
    """Красивый стартовый баннер"""
    banner = """
╔══════════════════════════════════════════════════════════╗
║                 🤖 AI RESUME ANALYZER                    ║
║                    Consumer v3.0                         ║
║          💾 WITH DATABASE INTEGRATION 📧                 ║
╚══════════════════════════════════════════════════════════╝
    """
    console.print(banner, style="bold blue")


def test_db_connection():
    """Тестирование подключения к БД при запуске"""
    console.print("🔌 [bold cyan]Тестирование подключения к PostgreSQL...[/bold cyan]")
    connection = get_db_connection()
    if connection:
        try:
            cursor = connection.cursor()
            cursor.execute("SELECT COUNT(*) FROM resumes;")
            count = cursor.fetchone()[0]
            console.print(f"✅ [bold green]Подключение к БД успешно! Найдено {count} резюме[/bold green]")
            cursor.close()
            connection.close()
            return True
        except Exception as e:
            console.print(f"❌ [bold red]Ошибка проверки БД:[/bold red] {e}")
            connection.close()
            return False
    return False


def main():
    """Основная функция consumer"""
    print_startup_banner()
    console.print("🚀 [bold blue]Запуск AI Consumer для анализа резюме с сохранением в БД...[/bold blue]")

    # Тестируем подключение к БД
    if not test_db_connection():
        console.print("💥 [bold red]Не удалось подключиться к БД. Завершение работы.[/bold red]")
        return

    processed_count = 0

    while True:
        try:
            # Подключение к RabbitMQ
            console.print("\n🔌 [bold cyan]Подключение к RabbitMQ...[/bold cyan]")
            connection = pika.BlockingConnection(pika.URLParameters(RABBITMQ_URL))
            channel = connection.channel()

            # Настройка очереди
            channel.exchange_declare(exchange='resume_exchange', exchange_type='direct', durable=True)
            channel.queue_declare(queue='resume_analysis_queue', durable=True)
            channel.queue_bind(exchange='resume_exchange', queue='resume_analysis_queue',
                               routing_key='resume_analysis_queue')

            console.print("✅ [bold green]Подключение к RabbitMQ установлено[/bold green]")
            console.print(f"👀 [bold cyan]Ожидание сообщений... (Обработано: {processed_count})[/bold cyan]")

            # Обработка сообщений
            while True:
                method_frame, _, body = channel.basic_get(queue='resume_analysis_queue', auto_ack=False)
                if method_frame:
                    process_resume_message(body)
                    channel.basic_ack(delivery_tag=method_frame.delivery_tag)
                    processed_count += 1

                    console.print(f"\n📊 [bold blue]Статистика: Обработано {processed_count} резюме[/bold blue]")
                else:
                    console.print("💤 [dim]Очередь пуста, ожидание 3 секунды...[/dim]")
                    time.sleep(3)

        except Exception as e:
            console.print(f"❌ [bold red]Ошибка соединения:[/bold red] {e}")
            console.print("🔄 [yellow]Переподключение через 10 секунд...[/yellow]")
            time.sleep(10)


if __name__ == "__main__":
    main()
