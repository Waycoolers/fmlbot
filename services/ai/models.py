from pydantic import BaseModel


class ComplimentRequest(BaseModel):
    partner_name: str
    context: str = ""


class DateIdeaRequest(BaseModel):
    event_type: str = "Годовщина"
    days_until: int = 7
    context: str = ""
