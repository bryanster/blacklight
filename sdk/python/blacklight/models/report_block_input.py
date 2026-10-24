from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.report_block_input_params import ReportBlockInputParams


T = TypeVar("T", bound="ReportBlockInput")


@_attrs_define
class ReportBlockInput:
    block_id: str
    params: ReportBlockInputParams | Unset = UNSET
    """ Block parameters, validated against the registry's ParamSchema.
    Defaults are used from the registry when absent.
    HTML content within params (e.g. rich_text `html`) is limited to
    100 KiB raw and sanitized server-side on write (M6-005).

    Parameter names come from the block's own ParamSchema and are not
    interchangeable: a key the schema does not declare is rejected, so
    a client that misspells one fails the whole save.
     """

    def to_dict(self) -> dict[str, Any]:
        from ..models.report_block_input_params import ReportBlockInputParams  # noqa: PLC0415

        block_id = self.block_id

        params: dict[str, Any] | Unset = UNSET
        if not isinstance(self.params, Unset):
            params = self.params.to_dict()

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "blockId": block_id,
            }
        )
        if params is not UNSET:
            field_dict["params"] = params

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.report_block_input_params import ReportBlockInputParams  # noqa: PLC0415

        d = dict(src_dict)
        block_id = d.pop("blockId")

        _params = d.pop("params", UNSET)
        params: ReportBlockInputParams | Unset
        if isinstance(_params, Unset):
            params = UNSET
        else:
            params = ReportBlockInputParams.from_dict(_params)

        report_block_input = cls(
            block_id=block_id,
            params=params,
        )

        return report_block_input
