FROM python:3.13-slim
WORKDIR /app

COPY services/ai/requirements.txt .
RUN --mount=type=cache,target=/root/.cache/pip pip install -r requirements.txt

COPY services/ai/ .

RUN useradd --system --no-create-home app
USER app

EXPOSE 8083

CMD ["sh", "-c", "exec python -m uvicorn main:app --host 0.0.0.0 --port ${AI_PORT:-8083}"]
