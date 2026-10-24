from enum import StrEnum


class ContentSyncJobKind(StrEnum):
    BUNDLE_IMPORT = "bundle_import"
    REPROCESS = "reprocess"
    SYNC = "sync"
    V1_IMPORT = "v1_import"

    def __str__(self) -> str:
        return str(self.value)
