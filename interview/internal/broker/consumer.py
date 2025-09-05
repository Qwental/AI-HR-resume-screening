import json
import os
import time
import pika
import logging
from pika.exceptions import AMQPConnectionError

# Настройка логирования
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# --- Константы из Go-сервиса ---
RABBITMQ_HOST = os.getenv("RABBITMQ_HOST", "localhost")
RABBITMQ_URL = f"amqp://guest:guest@{RABBITMQ_HOST}:5672/"
EXCHANGE_NAME = "resume_exchange"
QUEUE_NAME = "resume_analysis_queue"
ROUTING_KEY = "resume_analysis_queue"


def connect_rabbitmq(retries=5, delay=5):
    """Подключение к RabbitMQ с логикой повторных попыток."""
    for attempt in range(1, retries + 1):
        try:
            connection = pika.BlockingConnection(pika.URLParameters(RABBITMQ_URL))
            channel = connection.channel()
            channel.exchange_declare(exchange=EXCHANGE_NAME, exchange_type='direct', durable=True)
            channel.queue_declare(queue=QUEUE_NAME, durable=True)
            channel.queue_bind(exchange=EXCHANGE_NAME, queue=QUEUE_NAME, routing_key=ROUTING_KEY)
            logger.info("✅ Успешное подключение к RabbitMQ")
            return connection, channel
        except AMQPConnectionError as e:
            logger.error(f"❌ Попытка подключения {attempt}/{retries} не удалась: {e}")
            if attempt == retries:
                logger.error("❌ Превышено максимальное количество попыток. Выход.")
                raise
            time.sleep(delay)
    return None, None

def process_resume(body):
    """Обработка резюме с новой структурой данных."""
    try:
        message_data = json.loads(body)
    except json.JSONDecodeError as e:
        logger.error(f"❌ Ошибка парсинга JSON: {e}")
        return

    logger.info("="*60)
    logger.info(f"🔍 ПОЛУЧЕНО РЕЗЮМЕ ID: {message_data.get('id', 'N/A')}")

    text_resume = message_data.get('text_resume_jsonb', {})
    text_vacancy = message_data.get('text_vacancy_jsonb', {})

    logger.info(f"📄 Текст резюме: {len(json.dumps(text_resume))} символов")
    logger.info(f"📋 Текст вакансии: {len(json.dumps(text_vacancy))} символов")

    logger.info("⏳ Начинаем AI анализ...")
    processing_time = 3 + hash(message_data.get('id', '')) % 5
    time.sleep(processing_time)

    match_score = 50 + hash(message_data.get('id', '')) % 50
    logger.info(f"📊 Результат: соответствие {match_score}%")
    logger.info("✅ Анализ завершен!")
    logger.info("="*60)


def main():
    """Главная функция consumer."""
    logger.info("🚀 Запуск AI consumer для анализа резюме...")

    while True:
        connection, channel = None, None
        try:
            connection, channel = connect_rabbitmq()
            logger.info("👀 Ожидание сообщений... (Нажмите Ctrl+C для остановки)")

            while True:
                method_frame, _, body = channel.basic_get(queue=QUEUE_NAME, auto_ack=False)
                if method_frame:
                    try:
                        process_resume(body)
                        channel.basic_ack(delivery_tag=method_frame.delivery_tag)
                        logger.info("👍 Сообщение успешно обработано и подтверждено.")
                    except Exception as e:
                        logger.error(f"❌ Ошибка обработки сообщения: {e}")
                        channel.basic_nack(delivery_tag=method_frame.delivery_tag, requeue=True)
                else:
                    logger.info("💤 Очередь пуста, ожидание 3 секунды...")
                    time.sleep(3)

        except AMQPConnectionError as e:
            logger.error(f"🔌 Соединение потеряно: {e}. Переподключение через 10 секунд...")
            time.sleep(10)
        except KeyboardInterrupt:
            logger.info("🛑 Получен сигнал остановки.")
            break
        except Exception as e:
            logger.error(f"❌ Неожиданная ошибка: {e}")
            break
        finally:
            if connection and connection.is_open:
                connection.close()
                logger.info("🔌 Соединение с RabbitMQ закрыто.")

if __name__ == "__main__":
    main()
