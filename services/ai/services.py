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
            timeout=8.0,
            max_retries=0,
            default_headers={"x-data-logging-enabled": "false"},
        )

    def call_yandex_gpt(
        self, system_prompt: str, user_prompt: str, *, json_output: bool = False
    ) -> str:
        options = {"text": {"format": {"type": "json_object"}}, "timeout": 15.0} if json_output else {}
        try:
            response = self.client.responses.create(
                model=f"gpt://{settings.YANDEX_FOLDER_ID}/yandexgpt/latest",
                store=False,
                max_output_tokens=1100 if json_output else 700,
                input=[
                    {
                        "role": "system",
                        "content": system_prompt
                    },
                    {
                        "role": "user",
                        "content": user_prompt
                    }
                ],
                **options,
            )

        except Exception as error:
            logger.warning("YandexGPT request failed (%s)", type(error).__name__)
            raise HTTPException(status_code=502, detail="Idea generation failed") from error

        output = response.output_text.strip()
        if response.status != "completed" or not output:
            raise HTTPException(status_code=502, detail="Idea generation did not complete")
        return output
