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
    from ..models.content_emulation_plan_metadata import ContentEmulationPlanMetadata


T = TypeVar("T", bound="ContentEmulationPlan")


@_attrs_define
class ContentEmulationPlan:
    """One CTID adversary emulation plan (catalog only in M2). Threat actor /
    group identity is `adversaryName`. Source metadata (attack_version,
    format_version, archive path) is in `metadata`.

    """

    id: UUID
    source_id: UUID
    version: str
    """ Version token. CTID is rolling-head and always `current`. """
    external_id: str
    """ Stable id within the source. Prefer upstream
    `emulation_plan_details.id`; otherwise the actor directory slug
    (documented in `docs/content-ctid.md`).
     """
    name: str
    """ Display name (usually the adversary name). """
    description: str
    adversary_name: str
    """ Upstream threat-actor / group label (for example `FIN6`, `APT29`).
    Group refs are text labels in M2 — not resolved ATT&CK group rows.
     """
    metadata: ContentEmulationPlanMetadata
    """ Source-side bookkeeping: `attack_version`, `format_version`,
    archive `path`, `actor_slug` when known.
     """
    created_at: datetime.datetime
    updated_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.content_emulation_plan_metadata import ContentEmulationPlanMetadata  # noqa: PLC0415

        id = str(self.id)

        source_id = str(self.source_id)

        version = self.version

        external_id = self.external_id

        name = self.name

        description = self.description

        adversary_name = self.adversary_name

        metadata = self.metadata.to_dict()

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
                "adversaryName": adversary_name,
                "metadata": metadata,
                "createdAt": created_at,
                "updatedAt": updated_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.content_emulation_plan_metadata import ContentEmulationPlanMetadata  # noqa: PLC0415

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        source_id = UUID(d.pop("sourceId"))

        version = d.pop("version")

        external_id = d.pop("externalId")

        name = d.pop("name")

        description = d.pop("description")

        adversary_name = d.pop("adversaryName")

        metadata = ContentEmulationPlanMetadata.from_dict(d.pop("metadata"))

        created_at = datetime.datetime.fromisoformat(d.pop("createdAt"))

        updated_at = datetime.datetime.fromisoformat(d.pop("updatedAt"))

        content_emulation_plan = cls(
            id=id,
            source_id=source_id,
            version=version,
            external_id=external_id,
            name=name,
            description=description,
            adversary_name=adversary_name,
            metadata=metadata,
            created_at=created_at,
            updated_at=updated_at,
        )

        content_emulation_plan.additional_properties = d
        return content_emulation_plan

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
