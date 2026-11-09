from typing import Annotated

from pydantic import BaseModel, Field, StringConstraints


Criterion = Annotated[str, StringConstraints(strip_whitespace=True, min_length=1, max_length=80)]
Context = Annotated[str, StringConstraints(max_length=1000)]


class ComplimentRequest(BaseModel):
    partner_name: Annotated[str, StringConstraints(strip_whitespace=True, min_length=1, max_length=255)]
    context: Context = ""


class DateIdeaRequest(BaseModel):
    event_type: Criterion = "Годовщина"
    days_until: Annotated[int, Field(ge=0, le=366)] = 7
    context: Context = ""


class LeisureIdeaRequest(BaseModel):
    location: Criterion
    activity_level: Criterion
    budget: Criterion
    extra_context: Context = ""
