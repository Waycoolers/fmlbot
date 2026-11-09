import logging

from openai import OpenAI
from fastapi import HTTPException
from config import settings

logger = logging.getLogger(__name__)


class AIService:
    def __init__(self):
        self.client = OpenAI(
            api_key=settings.YANDEX_API_KEY,
            base_url="https://ai.api.cloud.yandex.net/v1",
            project=settings.YANDEX_FOLDER_ID,
            timeout=30.0,
            max_retries=1
        )

    def call_yandex_gpt(self, system_prompt: str, user_prompt: str) -> str:
        try:
            response = self.client.responses.create(
                model=f"gpt://{settings.YANDEX_FOLDER_ID}/yandexgpt/latest",
                input=[
                    {
                        "role": "system",
                        "content": system_prompt
                    },
                    {
                        "role": "user",
                        "content": user_prompt
                    }
                ]
            )

            return response.output_text

        except Exception as e:
            logger.error("YandexGPT integration error: %s", e)
            raise HTTPException(
                status_code=500,
                detail="YandexGPT integration error"
            )