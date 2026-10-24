from enum import StrEnum


class ContentSourceStatus(StrEnum):
    ERROR = "error"
    IDLE = "idle"
    SYNCING = "syncing"

    def __str__(self) -> str:
        return str(self.value)
