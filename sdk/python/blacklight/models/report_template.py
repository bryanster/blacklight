from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast
from uuid import UUID
import datetime

if TYPE_CHECKING:
    from ..models.report_template_block import ReportTemplateBlock


T = TypeVar("T", bound="ReportTemplate")


@_attrs_define
class ReportTemplate:
    id: UUID
    engagement_id: UUID
    name: str
    created_by: str
    created_at: datetime.datetime
    updated_at: datetime.datetime
    blocks: list[ReportTemplateBlock] | Unset = UNSET
    """ Template blocks in ordinal order. Present in GET /report-templates/{templateId} responses; absent from list
    endpoints. """

    def to_dict(self) -> dict[str, Any]:
        from ..models.report_template_block import ReportTemplateBlock  # noqa: PLC0415

        id = str(self.id)

        engagement_id = str(self.engagement_id)

        name = self.name

        created_by = self.created_by

        created_at = self.created_at.isoformat()

        updated_at = self.updated_at.isoformat()

        blocks: list[dict[str, Any]] | Unset = UNSET
        if not isinstance(self.blocks, Unset):
            blocks = []
            for blocks_item_data in self.blocks:
                blocks_item = blocks_item_data.to_dict()
                blocks.append(blocks_item)

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "id": id,
                "engagementId": engagement_id,
                "name": name,
                "createdBy": created_by,
                "createdAt": created_at,
                "updatedAt": updated_at,
            }
        )
        if blocks is not UNSET:
            field_dict["blocks"] = blocks

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.report_template_block import ReportTemplateBlock  # noqa: PLC0415

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        engagement_id = UUID(d.pop("engagementId"))

        name = d.pop("name")

        created_by = d.pop("createdBy")

        created_at = datetime.datetime.fromisoformat(d.pop("createdAt"))

        updated_at = datetime.datetime.fromisoformat(d.pop("updatedAt"))

        _blocks = d.pop("blocks", UNSET)
        blocks: list[ReportTemplateBlock] | Unset = UNSET
        if _blocks is not UNSET:
            blocks = []
            for blocks_item_data in _blocks:
                blocks_item = ReportTemplateBlock.from_dict(blocks_item_data)

                blocks.append(blocks_item)

        report_template = cls(
            id=id,
            engagement_id=engagement_id,
            name=name,
            created_by=created_by,
            created_at=created_at,
            updated_at=updated_at,
            blocks=blocks,
        )

        return report_template
