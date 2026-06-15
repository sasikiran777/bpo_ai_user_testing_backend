from __future__ import annotations

import logging
import os
import time
import uuid

from fastapi import Depends, FastAPI, Header, HTTPException, Request

from .grade import grade_request
from .schemas import GradeRequest, GradeResponse

try:
    from dotenv import load_dotenv

    load_dotenv()
except Exception:
    pass

_LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO").strip().upper() or "INFO"
logging.basicConfig(
    level=getattr(logging, _LOG_LEVEL, logging.INFO),
    format="%(asctime)s %(levelname)s [%(name)s] %(message)s",
)

app = FastAPI()
logger = logging.getLogger("grader")


def _auth(x_grader_token: str | None = Header(default=None)) -> None:
    expected = os.getenv("GRADER_TOKEN", "").strip()
    if not expected:
        return
    if (x_grader_token or "").strip() != expected:
        raise HTTPException(status_code=401, detail="Unauthorized")


@app.middleware("http")
async def log_requests(request: Request, call_next):
    request_id = request.headers.get("X-Request-ID") or str(uuid.uuid4())
    start = time.perf_counter()
    client_host = request.client.host if request.client else ""
    logger.info(
        "http_request_start request_id=%s method=%s path=%s client=%s",
        request_id,
        request.method,
        request.url.path,
        client_host,
    )
    try:
        response = await call_next(request)
    except Exception:
        logger.exception(
            "http_request_error request_id=%s method=%s path=%s duration_ms=%s",
            request_id,
            request.method,
            request.url.path,
            int((time.perf_counter() - start) * 1000),
        )
        raise
    logger.info(
        "http_request_finish request_id=%s method=%s path=%s status_code=%s duration_ms=%s",
        request_id,
        request.method,
        request.url.path,
        response.status_code,
        int((time.perf_counter() - start) * 1000),
    )
    response.headers["X-Request-ID"] = request_id
    return response


@app.get("/health")
def health() -> dict:
    logger.info("health_endpoint_hit")
    return {"ok": True}


@app.post("/v1/grade", response_model=GradeResponse)
def grade(req: GradeRequest, _: None = Depends(_auth)) -> GradeResponse:
    start = time.perf_counter()
    logger.info(
        "grade_endpoint_hit user_test_mapping_id=%s test_id=%s",
        req.user_test_mapping_id,
        req.test_id,
    )
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
