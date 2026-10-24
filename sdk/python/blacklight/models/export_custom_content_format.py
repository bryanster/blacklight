from enum import StrEnum


class ExportCustomContentFormat(StrEnum):
    JSON = "json"
    YAML = "yaml"

    def __str__(self) -> str:
        return str(self.value)
