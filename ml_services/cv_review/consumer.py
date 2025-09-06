# ml_services/cv_review/consumer.py
import json
import os
import time
import pika
import requests
import logging
from pika.exceptions import AMQPConnectionError
from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich.json import JSON
from rich.progress import track
from datetime import datetime

# Настройки подключения
# RABBITMQ_HOST = os.getenv("RABBITMQ_HOST", "localhost")
# ml_services/cv_review/consumer.py

# 🔧 Используем переменную окружения или fallback на guest
RABBITMQ_HOST = os.getenv("RABBITMQ_HOST", "localhost")
RABBITMQ_URL = f"amqp://guest:guest@{RABBITMQ_HOST}:5672/"  # 🎯 GUEST, не admin!
CV_REVIEW_SERVICE = os.getenv("CV_REVIEW_SERVICE", "http://localhost:5030")

EXCHANGE_NAME = "resume_exchange"
QUEUE_NAME = "resume_analysis_queue"
ROUTING_KEY = "resume_analysis_queue"

console = Console()
logger = logging.getLogger(__name__)

def log_message_received(message_data):
    """Логирование полученного сообщения"""
    console.print("\n" + "="*60, style="bold blue")
    console.print("📥 ПОЛУЧЕНО НОВОЕ СООБЩЕНИЕ", style="bold green", justify="center")
    console.print("="*60, style="bold blue")

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

def log_ai_request(resume_id, skillvals):
    """Логирование запроса к ИИ"""
    console.print(f"\n🚀 [bold yellow]ОТПРАВКА НА АНАЛИЗ ИИ[/bold yellow]")
    console.print(f"📋 Резюме ID: {resume_id}")
    console.print(f"⚖️ Веса навыков: {skillvals}")
    console.print(f"🤖 Сервис ИИ: {CV_REVIEW_SERVICE}")

def log_ai_response(resume_id, response, analysis_result=None):
    """Логирование ответа от ИИ"""
    if response.status_code == 200:
        console.print(f"\n✅ [bold green]УСПЕШНЫЙ АНАЛИЗ[/bold green]")
        console.print(f"📋 Резюме ID: {resume_id}")
        console.print(f"📊 Статус ответа: {response.status_code}")

        # Пытаемся извлечь основную информацию из результата
        if analysis_result and isinstance(analysis_result, str):
            # Показываем первые 200 символов результата
            preview = analysis_result[:200] + "..." if len(analysis_result) > 200 else analysis_result

            result_panel = Panel(
                preview,
                title="🎯 Результат анализа (превью)",
                border_style="green"
            )
            console.print(result_panel)

            # Пытаемся найти оценку в тексте
            if "%" in analysis_result or "балл" in analysis_result.lower() or "оценка" in analysis_result.lower():
                console.print("📈 [bold green]Найдена оценка в результате![/bold green]")

        console.print(f"📝 Размер ответа: {len(str(analysis_result))} символов")

    else:
        console.print(f"\n❌ [bold red]ОШИБКА АНАЛИЗА[/bold red]")
        console.print(f"📋 Резюме ID: {resume_id}")
        console.print(f"💥 HTTP статус: {response.status_code}")
        console.print(f"💬 Ошибка: {response.text}")

def process_resume_message(body):
    """Обработка сообщения из RabbitMQ с подробным логированием"""
    try:
        message_data = json.loads(body)

        # 📥 Логируем получение сообщения
        resume_id, cv_text, vacancy_data, weight_soft, weight_hard, weight_case = log_message_received(message_data)

        # 🎯 Формируем skillvals динамически
        skillvals = f"hard_skills:{weight_hard},soft_skills:{weight_soft},experience:{weight_case}"

        # 🚀 Логируем запрос к ИИ
        log_ai_request(resume_id, skillvals)

        # Формируем запрос к CV Review сервису
        review_request = {
            "vacancy": json.dumps(vacancy_data),
            "cv": cv_text,
            "skillvals": skillvals
        }

        # Отправляем запрос к FastAPI сервису
        start_time = time.time()

        response = requests.get(
            f"{CV_REVIEW_SERVICE}/get_review",
            params=review_request,
            timeout=120  # Увеличиваем timeout для ИИ
        )

        end_time = time.time()
        processing_time = round(end_time - start_time, 2)

        # 📊 Логируем результат
        if response.status_code == 200:
            analysis_result = response.json()
            log_ai_response(resume_id, response, analysis_result)

            console.print(f"⏱️ [bold blue]Время обработки: {processing_time}с[/bold blue]")

            # 🎉 Финальный вердикт
            console.print("\n" + "🎉 " * 20)
            console.print(f"[bold green]✅ РЕЗЮМЕ {resume_id} УСПЕШНО ОБРАБОТАНО![/bold green]", justify="center")
            console.print("🎉 " * 20)

            # TODO: Здесь можно отправить результат обратно в RabbitMQ
            # send_result_to_queue(resume_id, analysis_result)

        else:
            log_ai_response(resume_id, response)

            console.print("\n" + "💥 " * 20)
            console.print(f"[bold red]❌ ОШИБКА ОБРАБОТКИ РЕЗЮМЕ {resume_id}![/bold red]", justify="center")
            console.print("💥 " * 20)

    except json.JSONDecodeError as e:
        console.print(f"💥 [bold red]Ошибка парсинга JSON:[/bold red] {e}")
    except requests.exceptions.Timeout:
        console.print(f"⏰ [bold red]Таймаут запроса к ИИ сервису![/bold red]")
    except requests.exceptions.ConnectionError:
        console.print(f"🔌 [bold red]Ошибка подключения к ИИ сервису![/bold red]")
    except Exception as e:
        console.print(f"💥 [bold red]Неожиданная ошибка:[/bold red] {e}")
        import traceback
        console.print(traceback.format_exc())

def print_startup_banner():
    """Красивый стартовый баннер"""
    banner = """
╔══════════════════════════════════════════════════════════╗
║                 🤖 AI RESUME ANALYZER                    ║
║                    Consumer v2.0                         ║
╚══════════════════════════════════════════════════════════╝
    """
    console.print(banner, style="bold blue")

def main():
    """Основная функция consumer"""
    print_startup_banner()
    console.print("🚀 [bold blue]Запуск AI Consumer для анализа резюме...[/bold blue]")

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
            channel.queue_bind(exchange='resume_exchange', queue='resume_analysis_queue', routing_key='resume_analysis_queue')

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
