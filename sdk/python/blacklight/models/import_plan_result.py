from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast

if TYPE_CHECKING:
    from ..models.import_plan_warning import ImportPlanWarning
    from ..models.scenario import Scenario
    from ..models.step import Step


T = TypeVar("T", bound="ImportPlanResult")


@_attrs_define
class ImportPlanResult:
    scenario: Scenario
    steps: list[Step]
    step_count: int
    """ Total number of steps imported. """
    warnings: list[ImportPlanWarning]
    """ Steps whose technique external id did not resolve in the engagement's
    pinned ATT&CK version. These steps were still imported, but their
    technique/tactic/subtechnique fields are empty.
     """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.import_plan_warning import ImportPlanWarning  # noqa: PLC0415
        from ..models.scenario import Scenario  # noqa: PLC0415
        from ..models.step import Step  # noqa: PLC0415

        scenario = self.scenario.to_dict()

        steps = []
        for steps_item_data in self.steps:
            steps_item = steps_item_data.to_dict()
            steps.append(steps_item)

        step_count = self.step_count

        warnings = []
        for warnings_item_data in self.warnings:
            warnings_item = warnings_item_data.to_dict()
            warnings.append(warnings_item)

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "scenario": scenario,
                "steps": steps,
                "stepCount": step_count,
                "warnings": warnings,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.import_plan_warning import ImportPlanWarning  # noqa: PLC0415
        from ..models.scenario import Scenario  # noqa: PLC0415
        from ..models.step import Step  # noqa: PLC0415

        d = dict(src_dict)
        scenario = Scenario.from_dict(d.pop("scenario"))

        steps = []
        _steps = d.pop("steps")
        for steps_item_data in _steps:
            steps_item = Step.from_dict(steps_item_data)

            steps.append(steps_item)

        step_count = d.pop("stepCount")

        warnings = []
        _warnings = d.pop("warnings")
        for warnings_item_data in _warnings:
            warnings_item = ImportPlanWarning.from_dict(warnings_item_data)

            warnings.append(warnings_item)

        import_plan_result = cls(
            scenario=scenario,
            steps=steps,
            step_count=step_count,
            warnings=warnings,
        )

        import_plan_result.additional_properties = d
        return import_plan_result

    @property
    def additional_keys(self) -> list[str]:
        return list(self.additional_properties.keys())

    def __getitem__(self, key: str) -> Any:
        return self.additional_properties[key]

    def __setitem__(self, key: str, value: Any) -> None:
        self.additional_properties[key] = value

    def __delitem__(self, key: str) -> None:
        del self.additional_properties[key]

    def __contains__(self, key: str) -> bool:
        return key in self.additional_properties
