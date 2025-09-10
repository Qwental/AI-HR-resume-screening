FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt ./

RUN pip install --no-cache-dir -r requirements.txt

COPY server.py ./
COPY cv_example.txt ./
COPY vacancy_example.txt ./
COPY interview_prompt.txt ./

EXPOSE 5020

CMD ["uvicorn", "server:app", "--host", "0.0.0.0", "--port", "5020"]
