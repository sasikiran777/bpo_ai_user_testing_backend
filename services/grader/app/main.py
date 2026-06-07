from __future__ import annotations

import os

from fastapi import Depends, FastAPI, Header, HTTPException

from .grade import grade_request
from .schemas import GradeRequest, GradeResponse

try:
    from dotenv import load_dotenv

    load_dotenv()
except Exception:
    pass

app = FastAPI()


def _auth(x_grader_token: str | None = Header(default=None)) -> None:
    expected = os.getenv("GRADER_TOKEN", "").strip()
    if not expected:
        return
    if (x_grader_token or "").strip() != expected:
        raise HTTPException(status_code=401, detail="Unauthorized")


@app.get("/health")
def health() -> dict:
    return {"ok": True}


@app.post("/v1/grade", response_model=GradeResponse)
def grade(req: GradeRequest, _: None = Depends(_auth)) -> GradeResponse:
    return grade_request(req)
