from enum import StrEnum


class CreateUserRequestStatus(StrEnum):
    ACTIVE = "active"
    INVITED = "invited"

    def __str__(self) -> str:
        return str(self.value)
