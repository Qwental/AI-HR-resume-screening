import json
import time
import pika
import logging
from datetime import datetime

# Настройка логирования
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(message)s')
logger = logging.getLogger(__name__)

def connect_rabbitmq():
    """Подключение к RabbitMQ"""
    try:
        connection = pika.BlockingConnection(
            pika.URLParameters("amqp://guest:guest@localhost:5672/")
        )
        channel = connection.channel()

        # Объявляем exchange и queue
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

        logger.info("✅ Connected to RabbitMQ")
        return connection, channel, queue_name

    except Exception as e:
        logger.error(f"❌ Failed to connect to RabbitMQ: {e}")
        return None, None, None

def process_resume(message_data):
    """Обработка резюме - просто выводим данные"""
    print("\n" + "="*60)
    print("🔍 ПОЛУЧЕНО НОВОЕ РЕЗЮМЕ")
    print("="*60)

    # Красиво выводим JSON
    print(json.dumps(message_data, indent=2, ensure_ascii=False))

    print("-"*60)
    print(f"📋 ID резюме: {message_data.get('id', 'N/A')}")
    print(f"💼 ID вакансии: {message_data.get('vacancy_id', 'N/A')}")
    print(f"📧 Email: {message_data.get('mail', 'N/A')}")
    print(f"📅 Время создания: {message_data.get('created_at', 'N/A')}")
    print(f"💾 Storage Key: {message_data.get('storage_key', 'N/A')}")

    # Если есть текст резюме
    text_data = message_data.get('text_jsonb')
    if text_data:
        print(f"📄 Есть текст резюме: {len(str(text_data))} символов")

    print("-"*60)
    print("⏳ Имитируем обработку резюме...")

    # Имитируем работу 5 секунд
    for i in range(5, 0, -1):
        print(f"⏱️  Обработка... {i} секунд до завершения")
        time.sleep(1)

    print("✅ Резюме обработано!")
    print("="*60 + "\n")

def main():
    """Главная функция"""
    logger.info("🚀 Запуск простого consumer для резюме")

    # Подключаемся к RabbitMQ
    connection, channel, queue_name = connect_rabbitmq()
    if not connection:
        return

    try:
        logger.info("👀 Ожидание резюме из очереди...")
        logger.info("📝 Нажмите Ctrl+C для остановки")

        while True:
            try:
                # Получаем одно сообщение
                method_frame, header_frame, body = channel.basic_get(
                    queue=queue_name,
                    auto_ack=False
                )

                if method_frame:
                    # Есть сообщение
                    try:
                        message_data = json.loads(body.decode('utf-8'))
                        process_resume(message_data)

                        # Подтверждаем обработку
                        channel.basic_ack(delivery_tag=method_frame.delivery_tag)
                        logger.info("✅ Сообщение подтверждено")

                    except json.JSONDecodeError as e:
                        logger.error(f"❌ Ошибка парсинга JSON: {e}")
                        channel.basic_nack(
                            delivery_tag=method_frame.delivery_tag,
                            requeue=False
                        )
                    except Exception as e:
                        logger.error(f"❌ Ошибка обработки: {e}")
                        channel.basic_nack(
                            delivery_tag=method_frame.delivery_tag,
                            requeue=True
                        )
                else:
                    # Нет сообщений
                    print("💤 Очередь пуста, ждем 3 секунды...")
                    time.sleep(3)

            except pika.exceptions.ConnectionClosed:
                logger.warning("🔌 Соединение потеряно, переподключаемся...")
                connection, channel, queue_name = connect_rabbitmq()
                if not connection:
                    time.sleep(10)
                    continue

    except KeyboardInterrupt:
        logger.info("🛑 Получен сигнал остановки")
    except Exception as e:
        logger.error(f"❌ Неожиданная ошибка: {e}")
    finally:
        if connection and not connection.is_closed:
            connection.close()
            logger.info("🔌 Соединение закрыто")

if __name__ == "__main__":
    main()
