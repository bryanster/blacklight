from enum import StrEnum


class EngagementRole(StrEnum):
    BLUE = "blue"
    LEAD = "lead"
    OBSERVER = "observer"
    RED = "red"

    def __str__(self) -> str:
        return str(self.value)
