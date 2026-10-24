from enum import StrEnum


class ExecutionStatus(StrEnum):
    BLOCKED = "blocked"
    COMPLETE = "complete"
    PENDING = "pending"
    RUNNING = "running"
    SKIPPED = "skipped"

    def __str__(self) -> str:
        return str(self.value)
