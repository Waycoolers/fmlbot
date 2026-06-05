from fastapi import FastAPI
from router import router

app = FastAPI(
    title="FML AI Agent Microservice",
    description="Микросервис генерации рекомендаций на базе YandexGPT",
    version="1.0.0"
)

app.include_router(router)


@app.get("/health", tags=["System"])
def health_check():
    return {"status": "healthy"}
