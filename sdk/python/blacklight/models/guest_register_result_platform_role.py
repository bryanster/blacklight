from enum import StrEnum


class GuestRegisterResultPlatformRole(StrEnum):
    MEMBER = "member"

    def __str__(self) -> str:
        return str(self.value)
