from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast

if TYPE_CHECKING:
    from ..models.report_block_input import ReportBlockInput


T = TypeVar("T", bound="PutReportBlocks")


@_attrs_define
class PutReportBlocks:
    blocks: list[ReportBlockInput]

    def to_dict(self) -> dict[str, Any]:
        from ..models.report_block_input import ReportBlockInput  # noqa: PLC0415

        blocks = []
        for blocks_item_data in self.blocks:
            blocks_item = blocks_item_data.to_dict()
            blocks.append(blocks_item)

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "blocks": blocks,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.report_block_input import ReportBlockInput  # noqa: PLC0415

        d = dict(src_dict)
        blocks = []
        _blocks = d.pop("blocks")
        for blocks_item_data in _blocks:
            blocks_item = ReportBlockInput.from_dict(blocks_item_data)

            blocks.append(blocks_item)

        put_report_blocks = cls(
            blocks=blocks,
        )

        return put_report_blocks
