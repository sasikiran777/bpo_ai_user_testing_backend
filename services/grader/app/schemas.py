from __future__ import annotations

from typing import List, Optional

from pydantic import BaseModel, Field


class SectionAttempt(BaseModel):
    test_section_mapping_id: str = Field(..., min_length=1)
    name: str = Field(..., min_length=1)
    description: str = ""
    max_marks: int = Field(..., ge=0)
    questions: List[str] = Field(default_factory=list)
    answers: List[str] = Field(default_factory=list)
    test_notes: List[str] = Field(default_factory=list)


class GradeRequest(BaseModel):
    user_test_mapping_id: str = Field(..., min_length=1)
    test_id: str = Field(..., min_length=1)
    sections: List[SectionAttempt] = Field(default_factory=list)


class SectionResult(BaseModel):
    test_section_mapping_id: str
    marks_obtained: int
    ai_feedback: str
    transcript: Optional[str] = None


class GradeResponse(BaseModel):
    sections: List[SectionResult]
