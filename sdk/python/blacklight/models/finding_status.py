from enum import StrEnum


class FindingStatus(StrEnum):
    ACCEPTED_RISK = "accepted_risk"
    IN_PROGRESS = "in_progress"
    OPEN = "open"
    RESOLVED = "resolved"

    def __str__(self) -> str:
        return str(self.value)
