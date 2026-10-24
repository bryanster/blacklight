from enum import StrEnum


class ExecutionOutcome(StrEnum):
    DETECTED = "detected"
    NOT_APPLICABLE = "not_applicable"
    NOT_DETECTED = "not_detected"
    PREVENTED = "prevented"

    def __str__(self) -> str:
        return str(self.value)
