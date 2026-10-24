from enum import StrEnum


class EvidenceSide(StrEnum):
    BLUE = "blue"
    RED = "red"

    def __str__(self) -> str:
        return str(self.value)
