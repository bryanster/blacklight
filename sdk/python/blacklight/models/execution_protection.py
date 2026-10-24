from enum import StrEnum


class ExecutionProtection(StrEnum):
    BLOCKED = "blocked"
    NA = "n/a"
    NOT_BLOCKED = "not_blocked"
    PARTIAL = "partial"

    def __str__(self) -> str:
        return str(self.value)
