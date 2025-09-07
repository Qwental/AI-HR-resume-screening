# ЭТО ВРЕМЕННЫЙ ФАЙЛ



import json
import os
import time
import pika
import logging
from pika.exceptions import AMQPConnectionError

from rich.console import Console
from rich.json import JSON
from rich.panel import Panel
from rich.table import Table
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.text import Text
from rich.layout import Layout

# Настройка логирования
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

console = Console()

# --- Константы из Go-сервиса ---
RABBITMQ_HOST = os.getenv("RABBITMQ_HOST", "localhost")
RABBITMQ_URL = f"amqp://guest:guest@{RABBITMQ_HOST}:5672/"
EXCHANGE_NAME = "resume_exchange"
QUEUE_NAME = "resume_analysis_queue"
ROUTING_KEY = "resume_analysis_queue"


def print_startup_banner():
    """Красивый стартовый баннер"""
    banner = """
    ╔══════════════════════════════════════════════════════════╗
    ║                  RESUME  consumer                        ║
    ║              Система анализа резюме v1.0                 ║
    ╚══════════════════════════════════════════════════════════╝
    """
    console.print(banner, style="bold blue")


def connect_rabbitmq(retries=5, delay=5):
    """Подключение к RabbitMQ с логикой повторных попыток."""
    with console.status("[bold green]Подключение к RabbitMQ...", spinner="dots"):
        for attempt in range(1, retries + 1):
            try:
                connection = pika.BlockingConnection(pika.URLParameters(RABBITMQ_URL))
                channel = connection.channel()
                channel.exchange_declare(exchange=EXCHANGE_NAME, exchange_type='direct', durable=True)
                channel.queue_declare(queue=QUEUE_NAME, durable=True)
                channel.queue_bind(exchange=EXCHANGE_NAME, queue=QUEUE_NAME, routing_key=ROUTING_KEY)

                console.print("✅ [bold green]Успешное подключение к RabbitMQ[/bold green]")
                return connection, channel

            except AMQPConnectionError as e:
                console.print(f"❌ [bold red]Попытка подключения {attempt}/{retries} не удалась:[/bold red] {e}")
                if attempt == retries:
                    console.print("❌ [bold red]Превышено максимальное количество попыток. Выход.[/bold red]")
                    raise
                time.sleep(delay)
    return None, None


def display_resume_info(message_data):
    """Отображение основной информации о резюме в виде таблицы"""
    table = Table(title="📋 Информация о резюме", show_header=True, header_style="bold magenta")
    table.add_column("Поле", style="cyan", no_wrap=True)
    table.add_column("Значение", style="green")

    table.add_row("ID резюме", str(message_data.get('id', 'N/A')))
    table.add_row("ID вакансии", str(message_data.get('vacancy_id', 'N/A')))
    table.add_row("Кандидат", str(message_data.get('candidate_name', 'N/A')))
    table.add_row("Email", str(message_data.get('candidate_email', 'N/A')))
    table.add_row("Телефон", str(message_data.get('phone', 'N/A')))
    table.add_row("Опыт (лет)", str(message_data.get('experience_years', 'N/A')))

    console.print(table)


def display_json_data(title, data, style="blue"):
    """Красивое отображение JSON данных в панели"""
    if data:
        json_text = JSON.from_data(data)
        panel = Panel(json_text, title=title, border_style=style)
        console.print(panel)
    else:
        empty_panel = Panel(
            Text("Данные отсутствуют", style="italic red"),
            title=title,
            border_style="red"
        )
        console.print(empty_panel)


def simulate_ai_analysis(resume_id):
    """Симуляция AI анализа с прогресс-баром"""
    processing_time = 3 + hash(resume_id) % 5

    with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            console=console,
    ) as progress:
        task = progress.add_task("🧠 Выполняется AI анализ резюме...", total=processing_time)

        for i in range(processing_time):
            time.sleep(1)
            progress.update(task, advance=1)

    # Генерация результата
    match_score = 50 + hash(resume_id) % 50
    return match_score


def display_analysis_results(match_score):
    """Отображение результатов анализа"""
    if match_score >= 80:
        style = "bold green"
        emoji = "🎉"
        status = "ОТЛИЧНОЕ СООТВЕТСТВИЕ"
    elif match_score >= 60:
        style = "bold yellow"
        emoji = "👍"
        status = "ХОРОШЕЕ СООТВЕТСТВИЕ"
    else:
        style = "bold red"
        emoji = "👎"
        status = "СЛАБОЕ СООТВЕТСТВИЕ"

    result_text = f"{emoji} {status}: {match_score}%"
    result_panel = Panel(
        Text(result_text, style=style, justify="center"),
        title="📊 Результат анализа",
        border_style=style.split()[1]  # Извлекаем цвет из стиля
    )
    console.print(result_panel)


def process_resume(body):
    """Обработка резюме с красивым отображением."""
    try:
        message_data = json.loads(body)
    except json.JSONDecodeError as e:
        console.print(f"❌ [bold red]Ошибка парсинга JSON:[/bold red] {e}")
        return

    console.rule("🔍 НОВОЕ РЕЗЮМЕ ДЛЯ АНАЛИЗА", style="bold blue")

    # Отображение основной информации о резюме
    display_resume_info(message_data)

    # Красивый вывод JSON данных
    text_resume = message_data.get('text_resume_jsonb', {})
    text_vacancy = message_data.get('text_vacancy_jsonb', {})

    display_json_data("📄 Содержание резюме", text_resume, "green")
    display_json_data("📋 Требования вакансии", text_vacancy, "blue")

    # AI анализ с прогресс-баром
    resume_id = message_data.get('id', '')
    match_score = simulate_ai_analysis(resume_id)

    # Отображение результатов
    display_analysis_results(match_score)

    console.print("✅ [bold green]Анализ завершен успешно![/bold green]")
    console.rule(style="bold blue")


def main():
    """Главная функция consumer."""
    print_startup_banner()

    console.print("🚀 [bold blue]Запуск AI consumer для анализа резюме...[/bold blue]")

    while True:
        connection, channel = None, None
        try:
            connection, channel = connect_rabbitmq()
            console.print("👀 [bold cyan]Ожидание сообщений... (Нажмите Ctrl+C для остановки)[/bold cyan]")

            while True:
                method_frame, _, body = channel.basic_get(queue=QUEUE_NAME, auto_ack=False)
                if method_frame:
                    try:
                        process_resume(body)
                        channel.basic_ack(delivery_tag=method_frame.delivery_tag)
                        console.print("👍 [bold green]Сообщение успешно обработано и подтверждено.[/bold green]")
                    except Exception as e:
                        console.print(f"❌ [bold red]Ошибка обработки сообщения:[/bold red] {e}")
                        channel.basic_nack(delivery_tag=method_frame.delivery_tag, requeue=True)
                else:
                    console.print("💤 [dim]Очередь пуста, ожидание 3 секунды...[/dim]")
                    time.sleep(3)

        except AMQPConnectionError as e:
            console.print(f"🔌 [bold red]Соединение потеряно:[/bold red] {e}. [yellow]Переподключение через 10 секунд...[/yellow]")
            time.sleep(10)
        except KeyboardInterrupt:
            console.print("🛑 [bold yellow]Получен сигнал остановки.[/bold yellow]")
            break
        except Exception as e:
            console.print(f"❌ [bold red]Неожиданная ошибка:[/bold red] {e}")
            break
        finally:
            if connection and connection.is_open:
                connection.close()
                console.print("🔌 [bold blue]Соединение с RabbitMQ закрыто.[/bold blue]")


if __name__ == "__main__":
    main()
