from fastapi import APIRouter
from services import AIService

router = APIRouter(prefix="/api/v1/ai", tags=["AI Agent"])
ai_service = AIService()


@router.post("/compliment")
def generate_compliment(req: ComplimentRequest):
    system_prompt = (
        "Ты - эмпатичный ИИ-помощник в мобильном приложении для влюбленных пар. "
        "Помоги пользователю написать искренний и романтичный комплимент. "
        "Отвечай коротко, без вступлений. Выдай ровно ОДИН готовый комплимент."
    )
    user_prompt = f"Сгенерируй комплимент для партнера по имени {req.partner_name}."
    if req.context:
        user_prompt += f" Контекст: {req.context}."

    result = ai_service.call_yandex_gpt(system_prompt, user_prompt)
    return {"compliment": result.strip()}


@router.post("/date-idea")
def generate_date_idea(req: DateIdeaRequest):
    system_prompt = (
    )
    user_prompt = f"Событие: {req.event_type} через {req.days_until} дней. Контекст: {req.context}"

    result = ai_service.call_yandex_gpt(system_prompt, user_prompt)