from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast
from uuid import UUID
import datetime

if TYPE_CHECKING:
    from ..models.content_detection_rule_logsource import ContentDetectionRuleLogsource


T = TypeVar("T", bound="ContentDetectionRule")


@_attrs_define
class ContentDetectionRule:
    """One detection rule reference (Sigma or custom). Reference only —
    Blacklight never executes, deploys, or converts rules to product
    queries (`PLAN.md` §3).

    """

    id: UUID
    source_id: UUID
    version: str
    """ Version token. SigmaHQ is rolling-head and always `current`.
     """
    external_id: str
    """ Stable id within the source. Prefer upstream rule `id` when present;
    otherwise the archive-relative path (documented in
    `docs/content-sigma.md`).
     """
    name: str
    """ Rule title. """
    description: str
    technique_external_ids: list[str]
    """ ATT&CK technique external ids extracted from Sigma tags
    (`attack.t1059` → `T1059`). Only technique-mapped rules are stored.
     """
    level: str
    """ Sigma level (`informational`, `low`, `medium`, `high`, `critical`, …). """
    status: str
    """ Upstream rule status (`stable`, `test`, `experimental`, …). """
    logsource: ContentDetectionRuleLogsource
    """ Upstream `logsource` object as JSON (category/product/service/…). """
    rule_yaml: str
    """ Full rule body as YAML text, sufficient to display or copy in the UI.
    Never executed by Blacklight.
     """
    created_at: datetime.datetime
    updated_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.content_detection_rule_logsource import ContentDetectionRuleLogsource  # noqa: PLC0415

        id = str(self.id)

        source_id = str(self.source_id)

        version = self.version

        external_id = self.external_id

        name = self.name

        description = self.description

        technique_external_ids = self.technique_external_ids

        level = self.level

        status = self.status

        logsource = self.logsource.to_dict()

        rule_yaml = self.rule_yaml

        created_at = self.created_at.isoformat()

        updated_at = self.updated_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "sourceId": source_id,
                "version": version,
                "externalId": external_id,
                "name": name,
                "description": description,
                "techniqueExternalIds": technique_external_ids,
                "level": level,
                "status": status,
                "logsource": logsource,
                "ruleYaml": rule_yaml,
                "createdAt": created_at,
                "updatedAt": updated_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.content_detection_rule_logsource import ContentDetectionRuleLogsource  # noqa: PLC0415

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        source_id = UUID(d.pop("sourceId"))

        version = d.pop("version")

        external_id = d.pop("externalId")

        name = d.pop("name")

        description = d.pop("description")

        technique_external_ids = cast(list[str], d.pop("techniqueExternalIds"))

        level = d.pop("level")

        status = d.pop("status")

        logsource = ContentDetectionRuleLogsource.from_dict(d.pop("logsource"))

        rule_yaml = d.pop("ruleYaml")

        created_at = datetime.datetime.fromisoformat(d.pop("createdAt"))

        updated_at = datetime.datetime.fromisoformat(d.pop("updatedAt"))

        content_detection_rule = cls(
            id=id,
            source_id=source_id,
            version=version,
            external_id=external_id,
            name=name,
            description=description,
            technique_external_ids=technique_external_ids,
            level=level,
            status=status,
            logsource=logsource,
            rule_yaml=rule_yaml,
            created_at=created_at,
            updated_at=updated_at,
        )

        content_detection_rule.additional_properties = d
        return content_detection_rule

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
