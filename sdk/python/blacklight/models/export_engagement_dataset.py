from enum import StrEnum


class ExportEngagementDataset(StrEnum):
    COVERAGE = "coverage"
    EXECUTIONS = "executions"
    FINDINGS = "findings"

    def __str__(self) -> str:
        return str(self.value)
