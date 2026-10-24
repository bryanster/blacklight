from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.content_detection_logsource import ContentDetectionLogsource


T = TypeVar("T", bound="UpdateCustomDetectionRuleRequest")


@_attrs_define
class UpdateCustomDetectionRuleRequest:
    """Partial patch for a custom detection rule reference."""

    name: str | Unset = UNSET
    description: str | Unset = UNSET
    technique_external_ids: list[str] | Unset = UNSET
    level: str | Unset = UNSET
    status: str | Unset = UNSET
    logsource: ContentDetectionLogsource | Unset = UNSET
    """ Structured logsource fields for a custom detection rule. Known keys
    only — request bodies must not accept free-form properties
    (PLAN.md §4). Upstream Sigma rules still store arbitrary logsource
    JSON on the read model.
     """
    rule_yaml: str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.content_detection_logsource import ContentDetectionLogsource  # noqa: PLC0415

        name = self.name

        description = self.description

        technique_external_ids: list[str] | Unset = UNSET
        if not isinstance(self.technique_external_ids, Unset):
            technique_external_ids = self.technique_external_ids

        level = self.level

        status = self.status

        logsource: dict[str, Any] | Unset = UNSET
        if not isinstance(self.logsource, Unset):
            logsource = self.logsource.to_dict()

        rule_yaml = self.rule_yaml

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if name is not UNSET:
            field_dict["name"] = name
        if description is not UNSET:
            field_dict["description"] = description
        if technique_external_ids is not UNSET:
            field_dict["techniqueExternalIds"] = technique_external_ids
        if level is not UNSET:
            field_dict["level"] = level
        if status is not UNSET:
            field_dict["status"] = status
        if logsource is not UNSET:
            field_dict["logsource"] = logsource
        if rule_yaml is not UNSET:
            field_dict["ruleYaml"] = rule_yaml

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.content_detection_logsource import ContentDetectionLogsource  # noqa: PLC0415

        d = dict(src_dict)
        name = d.pop("name", UNSET)

        description = d.pop("description", UNSET)

        technique_external_ids = cast(list[str], d.pop("techniqueExternalIds", UNSET))

        level = d.pop("level", UNSET)

        status = d.pop("status", UNSET)

        _logsource = d.pop("logsource", UNSET)
        logsource: ContentDetectionLogsource | Unset
        if isinstance(_logsource, Unset):
            logsource = UNSET
        else:
            logsource = ContentDetectionLogsource.from_dict(_logsource)

        rule_yaml = d.pop("ruleYaml", UNSET)

        update_custom_detection_rule_request = cls(
            name=name,
            description=description,
            technique_external_ids=technique_external_ids,
            level=level,
            status=status,
            logsource=logsource,
            rule_yaml=rule_yaml,
        )

        update_custom_detection_rule_request.additional_properties = d
        return update_custom_detection_rule_request

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
