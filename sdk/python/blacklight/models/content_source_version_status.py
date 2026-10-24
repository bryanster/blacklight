from enum import StrEnum


class ContentSourceVersionStatus(StrEnum):
    ERROR = "error"
    PENDING = "pending"
    READY = "ready"

    def __str__(self) -> str:
        return str(self.value)
