from enum import StrEnum


class EngagementMode(StrEnum):
    BLIND = "blind"
    STANDARD = "standard"

    def __str__(self) -> str:
        return str(self.value)
