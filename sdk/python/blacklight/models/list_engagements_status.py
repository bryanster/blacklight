from enum import StrEnum


class ListEngagementsStatus(StrEnum):
    ACTIVE = "active"
    ARCHIVED = "archived"
    CLOSED = "closed"
    DRAFT = "draft"

    def __str__(self) -> str:
        return str(self.value)
