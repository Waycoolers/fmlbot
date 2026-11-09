from typing import Annotated, Literal

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


PlanStep = Annotated[str, StringConstraints(strip_whitespace=True, min_length=1, max_length=160)]
PlanDetail = Annotated[str, StringConstraints(strip_whitespace=True, min_length=1, max_length=220)]


class PlanCard(BaseModel):
    title: Criterion
    duration_minutes: Annotated[int, Field(strict=True, ge=1, le=1440)]
    estimated_cost: Annotated[str, StringConstraints(strip_whitespace=True, min_length=1, max_length=60)]
    preparation: PlanDetail
    steps: Annotated[list[PlanStep], Field(min_length=2, max_length=4)]
    backup: PlanDetail


class PlanOptionsRequest(BaseModel):
    brief: Annotated[str, StringConstraints(strip_whitespace=True, min_length=1, max_length=600)]
    adjustment: Literal["", "simpler", "cheaper", "more_involved", "more_expensive"] = ""
    reference: PlanCard | None = None


class PlanOptions(BaseModel):
    cards: Annotated[list[PlanCard], Field(min_length=2, max_length=2)]
