from enum import StrEnum


class DetectionModifier(StrEnum):
    ALERT = "alert"
    CONFIG_CHANGE = "config_change"
    CORRELATED = "correlated"
    DELAYED = "delayed"
    RESIDUAL_ARTIFACT = "residual_artifact"

    def __str__(self) -> str:
        return str(self.value)
