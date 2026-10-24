from enum import StrEnum


class DetectionCategory(StrEnum):
    GENERAL = "general"
    NONE = "none"
    TACTIC = "tactic"
    TECHNIQUE = "technique"
    TELEMETRY = "telemetry"

    def __str__(self) -> str:
        return str(self.value)
