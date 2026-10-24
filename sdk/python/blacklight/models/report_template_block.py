from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast

if TYPE_CHECKING:
    from ..models.report_template_block_params import ReportTemplateBlockParams


T = TypeVar("T", bound="ReportTemplateBlock")


@_attrs_define
class ReportTemplateBlock:
    ordinal: int
    block_id: str
    """ Stable block id from the catalogue (e.g. "cover", "rich_text"). """
    params: ReportTemplateBlockParams
    """ Block parameters, validated against the registry's ParamSchema. """

    def to_dict(self) -> dict[str, Any]:
        from ..models.report_template_block_params import ReportTemplateBlockParams  # noqa: PLC0415

        ordinal = self.ordinal

        block_id = self.block_id

        params = self.params.to_dict()

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "ordinal": ordinal,
                "blockId": block_id,
                "params": params,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.report_template_block_params import ReportTemplateBlockParams  # noqa: PLC0415

        d = dict(src_dict)
        ordinal = d.pop("ordinal")

        block_id = d.pop("blockId")

        params = ReportTemplateBlockParams.from_dict(d.pop("params"))

        report_template_block = cls(
            ordinal=ordinal,
            block_id=block_id,
            params=params,
        )

        return report_template_block
