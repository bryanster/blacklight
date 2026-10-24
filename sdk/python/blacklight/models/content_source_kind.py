from enum import StrEnum


class ContentSourceKind(StrEnum):
    ATOMIC = "atomic"
    ATTACK = "attack"
    CTID = "ctid"
    CUSTOM = "custom"
    SIGMA = "sigma"

    def __str__(self) -> str:
        return str(self.value)
