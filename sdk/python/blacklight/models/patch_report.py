from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.patch_report_colours_type_0 import PatchReportColoursType0


T = TypeVar("T", bound="PatchReport")


@_attrs_define
class PatchReport:
    title: str | Unset = UNSET
    client_name: None | str | Unset = UNSET
    logo_blob_ref: None | str | Unset = UNSET
    colours: None | PatchReportColoursType0 | Unset = UNSET

    def to_dict(self) -> dict[str, Any]:
        from ..models.patch_report_colours_type_0 import PatchReportColoursType0  # noqa: PLC0415

        title = self.title

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
        elif isinstance(self.colours, PatchReportColoursType0):
            colours = self.colours.to_dict()
        else:
            colours = self.colours

        field_dict: dict[str, Any] = {}

        field_dict.update({})
        if title is not UNSET:
            field_dict["title"] = title
        if client_name is not UNSET:
            field_dict["clientName"] = client_name
        if logo_blob_ref is not UNSET:
            field_dict["logoBlobRef"] = logo_blob_ref
        if colours is not UNSET:
            field_dict["colours"] = colours

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.patch_report_colours_type_0 import PatchReportColoursType0  # noqa: PLC0415

        d = dict(src_dict)
        title = d.pop("title", UNSET)

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

        def _parse_colours(data: object) -> None | PatchReportColoursType0 | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            try:
                if not isinstance(data, dict):
                    raise TypeError()
                colours_type_0 = PatchReportColoursType0.from_dict(data)

                return colours_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(None | PatchReportColoursType0 | Unset, data)

        colours = _parse_colours(d.pop("colours", UNSET))

        patch_report = cls(
            title=title,
            client_name=client_name,
            logo_blob_ref=logo_blob_ref,
            colours=colours,
        )

        return patch_report
