from enum import StrEnum


class ContentSyncJobStatus(StrEnum):
    CANCELLED = "cancelled"
    CANCELLING = "cancelling"
    FAILED = "failed"
    INTERRUPTED = "interrupted"
    QUEUED = "queued"
    RUNNING = "running"
    SUCCEEDED = "succeeded"

    def __str__(self) -> str:
        return str(self.value)
