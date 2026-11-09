from fastapi import FastAPI
from router import router

app = FastAPI(
    title="FML AI Agent Microservice",
    description="Микросервис генерации рекомендаций на базе YandexGPT",
    version="1.0.0",
    docs_url=None,
    redoc_url=None,
    openapi_url=None,
)

app.include_router(router)


@app.get("/health", tags=["System"])
def health_check():
    return {"status": "healthy"}
