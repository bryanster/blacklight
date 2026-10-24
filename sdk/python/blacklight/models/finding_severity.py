from enum import StrEnum


class FindingSeverity(StrEnum):
    CRITICAL = "critical"
    HIGH = "high"
    INFO = "info"
    LOW = "low"
    MEDIUM = "medium"

    def __str__(self) -> str:
        return str(self.value)
