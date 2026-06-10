from fastapi import APIRouter
from models import ComplimentRequest, DateIdeaRequest, LeisureIdeaRequest
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
        "Ты - эксперт по романтическим событиям. "
        "Предложи ОДНУ идею чем заняться чтобы отметить событие. "
        "Отвечай коротко, без вступлений."
    )
    user_prompt = f"Событие: {req.event_type} через {req.days_until} дней. Контекст: {req.context}"

    result = ai_service.call_yandex_gpt(system_prompt, user_prompt)
    return {"idea": result.strip()}


@router.post("/leisure-idea")
def generate_leisure_idea(req: LeisureIdeaRequest):
    system_prompt = (
        "Ты - креативный эксперт по досугу для влюбленных пар. "
        "Твоя задача: предложить ОДНУ детальную, необычную и выполнимую идею для совместного времяпрепровождения "
        "на основе заданных критериев. Опиши идею вкусно, по шагам (подготовка, сам процесс, фишки). "
        "Не предлагай банальности вроде 'посмотрите фильм' или 'сходите в ресторан', придумай что-то с изюминкой. "
        "Пиши не слишком подробно. Лучше чтобы было покороче."
    )

    user_prompt = (
        f"Локация: {req.location}. "
        f"Настроение/Активность: {req.activity_level}. "
        f"Бюджет: {req.budget}. "
    )

    if req.extra_context and req.extra_context.lower() != "пропустить":
        user_prompt += f"Дополнительные пожелания: {req.extra_context}."

    result = ai_service.call_yandex_gpt(system_prompt, user_prompt)
    return {"idea": result.strip()}
