import os


class Settings:
    YANDEX_API_KEY: str = os.getenv("YANDEX_API_KEY", "")
    YANDEX_FOLDER_ID: str = os.getenv("YANDEX_FOLDER_ID", "")
    PORT: int = int(os.getenv("AI_PORT", 8083))


settings = Settings()