from enum import StrEnum


class ContentSoftwareType(StrEnum):
    MALWARE = "malware"
    TOOL = "tool"

    def __str__(self) -> str:
        return str(self.value)
