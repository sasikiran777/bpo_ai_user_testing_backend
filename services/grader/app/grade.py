from __future__ import annotations

import json
import os
from typing import Any, Dict, List, Optional, Tuple

from openai import OpenAI
from pydantic import BaseModel, ConfigDict, Field

from .schemas import GradeRequest, GradeResponse, SectionResult
from .transcribe import transcribe_audio

_CLIENT: Optional[OpenAI] = None


def _get_client() -> OpenAI:
    global _CLIENT
    if _CLIENT is not None:
        return _CLIENT
    base_url = os.getenv("CHUTES_AI_URL", "").strip()
    api_key = os.getenv("CHUTES_AI_API_KEY", "").strip()
    _CLIENT = OpenAI(base_url=base_url, api_key=api_key)
    return _CLIENT


def _get_model() -> str:
    return os.getenv("DEEPSEEK_MODEL", "").strip()


def _is_speaking_section(name: str, description: str) -> bool:
    n = (name or "").strip().lower()
    if n in {"speak", "speaking"}:
        return True
    d = (description or "").strip().lower()
    return "speak" in d


def _clamp_marks(v: int, max_marks: int) -> int:
    if v < 0:
        return 0
    if v > max_marks:
        return max_marks
    return v


def _extract_json_array(raw: str) -> List[Dict[str, Any]]:
    raw = (raw or "").strip()
    if not raw:
        return []
    try:
        data = json.loads(raw)
        if isinstance(data, list):
            return [x for x in data if isinstance(x, dict)]
    except Exception:
        pass
    start = raw.find("[")
    end = raw.rfind("]")
    if start == -1 or end == -1 or end <= start:
        return []
    try:
        data = json.loads(raw[start : end + 1])
        if isinstance(data, list):
            return [x for x in data if isinstance(x, dict)]
    except Exception:
        return []
    return []


class _AISectionGrade(BaseModel):
    model_config = ConfigDict(extra="ignore")

    test_section_mapping_id: str = Field(..., min_length=1)
    marks_obtained: int = 0
    ai_feedback: str = ""


def _parse_ai_grades(raw: str) -> Dict[str, _AISectionGrade]:
    items = _extract_json_array(raw)
    if not items:
        return {}
    parsed: Dict[str, _AISectionGrade] = {}
    for item in items:
        try:
            g = _AISectionGrade.model_validate(item)
        except Exception:
            continue
        parsed[g.test_section_mapping_id] = g
    return parsed


def _build_prompt(
    req: GradeRequest,
    section_inputs: List[Tuple[str, str, int, List[str], List[str], List[str]]],
) -> str:
    header = f"""You are an examiner. Grade each section independently.
Return ONLY valid JSON array. No markdown.
marks_obtained must be between 0 and max_marks.
IMPORTANT: Do NOT change test_section_mapping_id values and do NOT change the order of the JSON array.
Return the same JSON array structure provided in OUTPUT_TEMPLATE with only marks_obtained and ai_feedback filled.

user_test_mapping_id: {req.user_test_mapping_id}
test_id: {req.test_id}

"""
    parts: List[str] = header.splitlines()

    for idx, (section_id, name, max_marks, questions, answers, test_notes) in enumerate(section_inputs, start=1):
        parts.extend(
            f"""section_{idx}:
test_section_mapping_id: {section_id}
name: {name}
max_marks: {max_marks}""".splitlines()
        )
        if test_notes:
            parts.append("test_notes:")
            for n in test_notes:
                n = (n or "").strip()
                if n:
                    parts.append(f"- {n}")
        parts.append("answers:")
        if questions and answers and len(questions) == len(answers):
            for q, a in zip(questions, answers):
                parts.append(f"Q: {q}")
                parts.append(f"A: {a}")
        elif answers:
            for a in answers:
                parts.append(f"A: {a}")
        parts.append("")

    parts.append("OUTPUT_TEMPLATE:")
    template: List[Dict[str, Any]] = [
        {
            "test_section_mapping_id": section_id,
            "section_name": name,
            "max_marks": max_marks,
            "marks_obtained": 0,
            "ai_feedback": "",
        }
        for section_id, name, max_marks, *_ in section_inputs
    ]
    parts.append(json.dumps(template, ensure_ascii=False, indent=2))
    parts.append("")
    parts.append("Return the JSON now.")
    parts.append("")
    return "\n".join(parts)


def grade_request(req: GradeRequest) -> GradeResponse:
    section_inputs: List[Tuple[str, str, int, List[str], List[str], List[str]]] = []
    transcripts: Dict[str, str] = {}

    for s in req.sections:
        if _is_speaking_section(s.name, s.description):
            transcript = ""
            if s.test_notes:
                transcript = (s.test_notes[0] or "").strip()
            if not transcript and s.answers:
                transcript = transcribe_audio(s.answers[0])
            transcript = (transcript or "").strip()
            transcripts[s.test_section_mapping_id] = transcript
            section_inputs.append(
                (
                    s.test_section_mapping_id,
                    s.name,
                    s.max_marks,
                    s.questions,
                    [transcript] if transcript else [],
                    [transcript] if transcript else [],
                )
            )
            continue

        section_inputs.append(
            (
                s.test_section_mapping_id,
                s.name,
                s.max_marks,
                s.questions,
                s.answers,
                s.test_notes,
            )
        )

    prompt = _build_prompt(req, section_inputs)

    client = _get_client()
    model = _get_model()
    resp = client.chat.completions.create(
        model=model,
        messages=[{"role": "user", "content": prompt}],
        temperature=0,
    )
    raw = ""
    if resp.choices and resp.choices[0].message and resp.choices[0].message.content:
        raw = resp.choices[0].message.content

    parsed = _parse_ai_grades(raw)

    results: List[SectionResult] = []
    for s in req.sections:
        marks = 0
        feedback = ""
        g = parsed.get(s.test_section_mapping_id)
        if g:
            marks = g.marks_obtained
            feedback = g.ai_feedback

        marks = _clamp_marks(marks, s.max_marks)
        transcript = transcripts.get(s.test_section_mapping_id)
        results.append(
            SectionResult(
                test_section_mapping_id=s.test_section_mapping_id,
                marks_obtained=marks,
                ai_feedback=feedback,
                transcript=transcript if transcript else None,
            )
        )

    return GradeResponse(sections=results)
