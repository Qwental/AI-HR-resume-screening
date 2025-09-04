import json
import pika
import uuid
from datetime import datetime
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def send_test_resume():
    """Отправляем тестовое резюме в брокер"""
    try:
        # Подключаемся к RabbitMQ
        connection = pika.BlockingConnection(
            pika.URLParameters("amqp://guest:guest@localhost:5672/")
        )
        channel = connection.channel()

        # Настраиваем exchange и queue
        exchange_name = "resume_processing"
        queue_name = "resume_queue"

        channel.exchange_declare(
            exchange=exchange_name,
            exchange_type='direct',
            durable=True
        )
        channel.queue_declare(queue=queue_name, durable=True)
        channel.queue_bind(
            exchange=exchange_name,
            queue=queue_name,
            routing_key=queue_name
        )

        # Тестовые данные резюме
        test_resumes = [
            {
                "id": str(uuid.uuid4()),
                "vacancy_id": str(uuid.uuid4()),
                "created_at": datetime.now().isoformat(),
                "storage_key": "resumes/test_resume_1.pdf",
                "mail": "john.doe@example.com",
                "text_jsonb": {
                    "content": "Senior Python Developer с 5 годами опыта. Знание Django, FastAPI, PostgreSQL.",
                    "skills": ["Python", "Django", "PostgreSQL", "Redis", "Docker"],
                    "experience": "5 лет",
                    "education": "ВУЗ: МГУ, Факультет ВМК"
                }
            },
            {
                "id": str(uuid.uuid4()),
                "vacancy_id": str(uuid.uuid4()),
                "created_at": datetime.now().isoformat(),
                "storage_key": "resumes/test_resume_2.pdf",
                "mail": "jane.smith@example.com",
                "text_jsonb": {
                    "content": "Full-stack разработчик. React, Node.js, Go, Kubernetes.",
                    "skills": ["React", "Node.js", "Go", "Kubernetes", "AWS"],
                    "experience": "3 года",
                    "education": "МГТУ им. Баумана"
                }
            },
            {
                "id": str(uuid.uuid4()),
                "vacancy_id": str(uuid.uuid4()),
                "created_at": datetime.now().isoformat(),
                "storage_key": "resumes/test_resume_3.pdf",
                "mail": "alex.petrov@example.com",
                "text_jsonb": {
                    "content": "DevOps Engineer. Опыт с микросервисами, CI/CD, мониторингом.",
                    "skills": ["Docker", "Kubernetes", "Jenkins", "Prometheus", "Grafana"],
                    "experience": "4 года",
                    "education": "СПбГУ"
                }
            }
        ]

        # Отправляем каждое резюме
        for i, resume_data in enumerate(test_resumes, 1):
            message = json.dumps(resume_data, ensure_ascii=False)

            channel.basic_publish(
                exchange=exchange_name,
                routing_key=queue_name,
                body=message,
                properties=pika.BasicProperties(
                    content_type='application/json',
                    delivery_mode=2,  # Persistent
                    message_id=resume_data["id"]
                )
            )

            logger.info(f"✅ Отправлено тестовое резюме #{i}: {resume_data['mail']}")

        connection.close()
        logger.info(f"🎉 Отправлено {len(test_resumes)} тестовых резюме!")

    except Exception as e:
        logger.error(f"❌ Ошибка отправки: {e}")

if __name__ == "__main__":
    print("🚀 Отправка тестовых резюме в RabbitMQ...")
    send_test_resume()
