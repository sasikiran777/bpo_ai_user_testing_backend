from __future__ import annotations

import os
from pathlib import Path
from typing import Optional

from faster_whisper import WhisperModel

_MODEL: Optional[WhisperModel] = None


def _get_model() -> WhisperModel:
    global _MODEL
    if _MODEL is not None:
        return _MODEL
    model_name = os.getenv("WHISPER_MODEL", "small")
    device = os.getenv("WHISPER_DEVICE", "cpu")
    compute_type = os.getenv("WHISPER_COMPUTE_TYPE", "int8")
    _MODEL = WhisperModel(model_name, device=device, compute_type=compute_type)
    return _MODEL


def _resolve_storage_path(p: str) -> Path:
    p = p.strip().replace("\\", "/")
    if p.startswith("/"):
        raise ValueError("absolute paths are not allowed")
    if not p.startswith("storage/audio/"):
        raise ValueError("not an audio storage path")
    base = Path(os.getenv("STORAGE_ROOT", ".")).resolve()
    full = (base / p).resolve()
    if base not in full.parents and full != base:
        raise ValueError("invalid path")
    if not full.exists() or not full.is_file():
        raise FileNotFoundError(str(full))
    return full


def transcribe_audio(audio_path: str) -> str:
    full = _resolve_storage_path(audio_path)
    model = _get_model()
    segments, _ = model.transcribe(str(full), vad_filter=True)
    text_parts = []
    for s in segments:
        t = (s.text or "").strip()
        if t:
            text_parts.append(t)
    return " ".join(text_parts).strip()
