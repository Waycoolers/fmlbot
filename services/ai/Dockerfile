FROM python:3.13-slim
WORKDIR /app

COPY . .

WORKDIR /app/services/ai

RUN pip install --no-cache-dir -r requirements.txt

EXPOSE 8083

CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8083"]