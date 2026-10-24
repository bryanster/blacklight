from enum import StrEnum


class ScenarioSource(StrEnum):
    CTID = "ctid"
    IMPORTED = "imported"
    MANUAL = "manual"

    def __str__(self) -> str:
        return str(self.value)
