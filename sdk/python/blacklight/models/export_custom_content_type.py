from enum import StrEnum


class ExportCustomContentType(StrEnum):
    DETECTION_RULES = "detection_rules"
    NOTES = "notes"
    PROCEDURE_TEMPLATES = "procedure_templates"

    def __str__(self) -> str:
        return str(self.value)
