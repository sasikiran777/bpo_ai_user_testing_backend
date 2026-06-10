from __future__ import annotations

import os
import tempfile
from pathlib import Path
from typing import Optional
from urllib.parse import urlparse

import boto3
from faster_whisper import WhisperModel

_MODEL: Optional[WhisperModel] = None
_S3_CLIENT = None


def _get_model() -> WhisperModel:
    global _MODEL
    if _MODEL is not None:
        return _MODEL
    model_name = os.getenv("WHISPER_MODEL", "small")
    device = os.getenv("WHISPER_DEVICE", "cpu")
    compute_type = os.getenv("WHISPER_COMPUTE_TYPE", "int8")
    _MODEL = WhisperModel(model_name, device=device, compute_type=compute_type)
    return _MODEL


def _get_s3_client():
    global _S3_CLIENT
    if _S3_CLIENT is not None:
        return _S3_CLIENT

    kwargs = {}
    region = os.getenv("AWS_S3_REGION", "").strip()
    if region:
        kwargs["region_name"] = region

    access_key_id = os.getenv("AWS_S3_ACCESS_KEY_ID", "").strip()
    secret_access_key = os.getenv("AWS_S3_SECRET_ACCESS_KEY", "").strip()
    if access_key_id and secret_access_key:
        kwargs["aws_access_key_id"] = access_key_id
        kwargs["aws_secret_access_key"] = secret_access_key

    _S3_CLIENT = boto3.client("s3", **kwargs)
    return _S3_CLIENT


def _download_s3_audio(p: str) -> Path:
    parsed = urlparse(p.strip())
    if parsed.scheme != "s3" or not parsed.netloc or not parsed.path:
        raise ValueError("invalid s3 audio path")

    bucket = parsed.netloc
    key = parsed.path.lstrip("/")
    suffix = Path(key).suffix or ".bin"
    fd, tmp_path = tempfile.mkstemp(prefix="grader-audio-", suffix=suffix)
    os.close(fd)

    try:
        _get_s3_client().download_file(bucket, key, tmp_path)
    except Exception:
        try:
            os.remove(tmp_path)
        except OSError:
            pass
        raise

    return Path(tmp_path)


def transcribe_audio(audio_path: str) -> str:
    if not (audio_path or "").strip().startswith("s3://"):
        raise ValueError("audio reference must use s3")

    full = _download_s3_audio(audio_path)
    model = _get_model()
    try:
        segments, _ = model.transcribe(str(full), vad_filter=True)
        text_parts = []
        for s in segments:
            t = (s.text or "").strip()
            if t:
                text_parts.append(t)
        return " ".join(text_parts).strip()
    finally:
        try:
            full.unlink(missing_ok=True)
        except OSError:
            pass
