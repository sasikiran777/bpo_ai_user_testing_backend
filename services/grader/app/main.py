from __future__ import annotations

import logging
import os
import time

from fastapi import Depends, FastAPI, Header, HTTPException

from .grade import grade_request
from .schemas import GradeRequest, GradeResponse

try:
    from dotenv import load_dotenv

    load_dotenv()
except Exception:
    pass

app = FastAPI()
logger = logging.getLogger("grader")


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
    start = time.perf_counter()
    logger.info(
        "grade_request_start user_test_mapping_id=%s test_id=%s sections=%s",
        req.user_test_mapping_id,
        req.test_id,
        len(req.sections),
    )
    try:
        out = grade_request(req)
        logger.info(
            "grade_request_finish user_test_mapping_id=%s duration_ms=%s",
            req.user_test_mapping_id,
            int((time.perf_counter() - start) * 1000),
        )
        return out
    except Exception:
        logger.exception(
            "grade_request_error user_test_mapping_id=%s duration_ms=%s",
            req.user_test_mapping_id,
            int((time.perf_counter() - start) * 1000),
        )
        raise
