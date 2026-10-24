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
    from ..models.report_block import ReportBlock
    from ..models.report_colours_type_0 import ReportColoursType0


T = TypeVar("T", bound="Report")


@_attrs_define
class Report:
    id: UUID
    engagement_id: UUID
    title: str
    created_by: str
    created_at: datetime.datetime
    updated_at: datetime.datetime
    block_count: int
    """ How many draft blocks the report holds. Always present, including on list endpoints, which omit `blocks`
    itself — a list row that wants to show a count must read this rather than measuring `blocks`. """
    client_name: None | str | Unset = UNSET
    logo_blob_ref: None | str | Unset = UNSET
    colours: None | ReportColoursType0 | Unset = UNSET
    """ Nullable JSON object with optional primary/secondary hex colours. """
    updated_by: None | str | Unset = UNSET
    blocks: list[ReportBlock] | Unset = UNSET
    """ Draft blocks in ordinal order. Present in GET /reports/{reportId} responses; absent from list endpoints. """

    def to_dict(self) -> dict[str, Any]:
        from ..models.report_block import ReportBlock  # noqa: PLC0415
        from ..models.report_colours_type_0 import ReportColoursType0  # noqa: PLC0415

        id = str(self.id)

        engagement_id = str(self.engagement_id)

        title = self.title

        created_by = self.created_by

        created_at = self.created_at.isoformat()

        updated_at = self.updated_at.isoformat()

        block_count = self.block_count

        client_name: None | str | Unset
        if isinstance(self.client_name, Unset):
            client_name = UNSET
        else:
            client_name = self.client_name

        logo_blob_ref: None | str | Unset
        if isinstance(self.logo_blob_ref, Unset):
            logo_blob_ref = UNSET
        else:
            logo_blob_ref = self.logo_blob_ref

        colours: dict[str, Any] | None | Unset
        if isinstance(self.colours, Unset):
            colours = UNSET
        elif isinstance(self.colours, ReportColoursType0):
            colours = self.colours.to_dict()
        else:
            colours = self.colours

        updated_by: None | str | Unset
        if isinstance(self.updated_by, Unset):
            updated_by = UNSET
        else:
            updated_by = self.updated_by

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
                "title": title,
                "createdBy": created_by,
                "createdAt": created_at,
                "updatedAt": updated_at,
                "blockCount": block_count,
            }
        )
        if client_name is not UNSET:
            field_dict["clientName"] = client_name
        if logo_blob_ref is not UNSET:
            field_dict["logoBlobRef"] = logo_blob_ref
        if colours is not UNSET:
            field_dict["colours"] = colours
        if updated_by is not UNSET:
            field_dict["updatedBy"] = updated_by
        if blocks is not UNSET:
            field_dict["blocks"] = blocks

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.report_block import ReportBlock  # noqa: PLC0415
        from ..models.report_colours_type_0 import ReportColoursType0  # noqa: PLC0415

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        engagement_id = UUID(d.pop("engagementId"))

        title = d.pop("title")

        created_by = d.pop("createdBy")

        created_at = datetime.datetime.fromisoformat(d.pop("createdAt"))

        updated_at = datetime.datetime.fromisoformat(d.pop("updatedAt"))

        block_count = d.pop("blockCount")

        def _parse_client_name(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        client_name = _parse_client_name(d.pop("clientName", UNSET))

        def _parse_logo_blob_ref(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        logo_blob_ref = _parse_logo_blob_ref(d.pop("logoBlobRef", UNSET))

        def _parse_colours(data: object) -> None | ReportColoursType0 | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            try:
                if not isinstance(data, dict):
                    raise TypeError()
                colours_type_0 = ReportColoursType0.from_dict(data)

                return colours_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(None | ReportColoursType0 | Unset, data)

        colours = _parse_colours(d.pop("colours", UNSET))

        def _parse_updated_by(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        updated_by = _parse_updated_by(d.pop("updatedBy", UNSET))

        _blocks = d.pop("blocks", UNSET)
        blocks: list[ReportBlock] | Unset = UNSET
        if _blocks is not UNSET:
            blocks = []
            for blocks_item_data in _blocks:
                blocks_item = ReportBlock.from_dict(blocks_item_data)

                blocks.append(blocks_item)

        report = cls(
            id=id,
            engagement_id=engagement_id,
            title=title,
            created_by=created_by,
            created_at=created_at,
            updated_at=updated_at,
            block_count=block_count,
            client_name=client_name,
            logo_blob_ref=logo_blob_ref,
            colours=colours,
            updated_by=updated_by,
            blocks=blocks,
        )

        return report
