from __future__ import annotations

import os


def verbose_logging_enabled() -> bool:
    return os.getenv("GRADER_VERBOSE_LOGS", "").strip().lower() in {"1", "true", "yes", "on"}

