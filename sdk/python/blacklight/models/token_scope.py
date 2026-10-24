from enum import StrEnum


class TokenScope(StrEnum):
    ADMINREAD = "admin:read"
    ADMINWRITE = "admin:write"
    CONTENTREAD = "content:read"
    CONTENTSYNC = "content:sync"
    CONTENTWRITE = "content:write"
    ENGAGEMENTSREAD = "engagements:read"
    ENGAGEMENTSWRITE = "engagements:write"
    REPORTSREAD = "reports:read"
    REPORTSWRITE = "reports:write"

    def __str__(self) -> str:
        return str(self.value)
