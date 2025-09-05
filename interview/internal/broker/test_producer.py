import json
import os
import pika
import uuid
import logging
from datetime import datetime

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# --- Константы из Go-сервиса ---
RABBITMQ_HOST = os.getenv("RABBITMQ_HOST", "localhost")
RABBITMQ_URL = f"amqp://guest:guest@{RABBITMQ_HOST}:5672/"
EXCHANGE_NAME = "resume_exchange"
QUEUE_NAME = "resume_analysis_queue"
ROUTING_KEY = "resume_analysis_queue"


def send_test_resume():
    """Отправляет тестовое резюме в брокер."""
    try:
        connection = pika.BlockingConnection(pika.URLParameters(RABBITMQ_URL))
        channel = connection.channel()

        channel.exchange_declare(exchange=EXCHANGE_NAME, exchange_type='direct', durable=True)
        channel.queue_declare(queue=QUEUE_NAME, durable=True)
        channel.queue_bind(exchange=EXCHANGE_NAME, queue=QUEUE_NAME, routing_key=ROUTING_KEY)

        # Тестовые данные в формате, как от Go-сервиса
        resume_id = str(uuid.uuid4())
        vacancy_id = str(uuid.uuid4())

        message_data = {
            "id": resume_id,
            "vacancy_id": vacancy_id,
            "text_resume_jsonb": {
                "text": "Опытный Go разработчик с 5-летним стажем. Работал с микросервисами, PostgreSQL, Docker и Kubernetes.",
                "extracted_at": datetime.now().isoformat()
            },
            "text_vacancy_jsonb": {
                "structured_data": {
                    "Название": "Senior Go Developer",
                    "Требования": "Требуется опыт работы с Go, gRPC, PostgreSQL. Уверенное знание микросервисной архитектуры.",
                    "Обязанности": "Разработка высоконагруженных сервисов."
                },
                "extracted_at": datetime.now().isoformat()
            }
        }

        message_body = json.dumps(message_data, ensure_ascii=False)

        channel.basic_publish(
            exchange=EXCHANGE_NAME,
            routing_key=ROUTING_KEY,
            body=message_body,
            properties=pika.BasicProperties(
                content_type='application/json',
                delivery_mode=2,  # Persistent
                message_id=resume_id
            )
        )
        logger.info(f"✅ Отправлено тестовое резюме. ID: {resume_id}")
        connection.close()

    except Exception as e:
        logger.error(f"❌ Ошибка отправки: {e}")

if __name__ == "__main__":
    logger.info("🚀 Отправка тестового сообщения в RabbitMQ...")
    send_test_resume()
