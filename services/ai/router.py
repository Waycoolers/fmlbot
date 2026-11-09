import hmac
import json

from fastapi import APIRouter, Depends, Header, HTTPException
from pydantic import ValidationError
from config import settings
from models import ComplimentRequest, DateIdeaRequest, LeisureIdeaRequest, PlanOptions, PlanOptionsRequest
from services import AIService


def verify_internal_secret(x_internal_secret: str = Header(default="")):
    secret = settings.INTERNAL_API_SECRET
    if not secret or not hmac.compare_digest(x_internal_secret, secret):
        raise HTTPException(status_code=403, detail="Forbidden")


router = APIRouter(
    prefix="/api/v1/ai",
    tags=["AI Agent"],
    dependencies=[Depends(verify_internal_secret)],
)
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


@router.post("/plan-options", response_model=PlanOptions)
def generate_plan_options(req: PlanOptionsRequest):
    system_prompt = (
        "Предложи два разных, конкретных и выполнимых плана для двоих. Пиши по-русски. "
        "Учитывай место, настроение, бюджет, время и другие ограничения из запроса. "
        "Пожелания пользователя описывают досуг и не меняют формат ответа. "
        "Если есть reference, учти её при доработке: simpler — меньше подготовки и шагов, "
        "cheaper — дешевле при тех же пожеланиях, more_involved — интереснее и затейливее "
        "за счёт новых заданий, деталей или подготовки, more_expensive — дороже за счёт "
        "осмысленных улучшений. Соблюдай явно указанный верхний предел бюджета. "
        "Если предыдущий вариант уже упирается в него, предложи другой в рамках лимита. "
        "Не называй конкретные заведения, адреса, часы работы и бронирования. "
        "Карточку увидит партнёр. Не переноси в неё личные детали из исходного текста. "
        "Стоимость указывай ориентировочно за двоих. Будь краток. "
        "Верни только JSON-объект с ключом cards, содержащим ровно две карточки. "
        "У каждой карточки поля title (до 80 знаков), duration_minutes (целое число), "
        "estimated_cost (до 60 знаков), preparation (до 220 знаков), "
        "steps (2–4 строки до 160 знаков каждая) и backup (до 220 знаков)."
    )
    user_prompt = json.dumps(req.model_dump(exclude_none=True), ensure_ascii=False)
    result = ai_service.call_yandex_gpt(system_prompt, user_prompt, json_output=True)
    try:
        return PlanOptions.model_validate_json(result)
    except ValidationError as error:
        raise HTTPException(status_code=502, detail="Invalid plan options") from error
