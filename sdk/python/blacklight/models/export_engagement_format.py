from enum import StrEnum


class ExportEngagementFormat(StrEnum):
    CSV = "csv"
    JSON = "json"

    def __str__(self) -> str:
        return str(self.value)
