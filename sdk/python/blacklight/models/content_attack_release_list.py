from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.content_attack_release import ContentAttackRelease


T = TypeVar("T", bound="ContentAttackReleaseList")


@_attrs_define
class ContentAttackReleaseList:
    items: list[ContentAttackRelease]
    reachable: bool
    """ Whether the upstream index answered. `false` is a normal outcome on
    an air-gapped installation, not a failed request.
     """
    source_enabled: bool
    """ Whether the ATT&CK source is enabled. A sync started against a
    disabled source is refused, so a picker should say so before
    offering one.
     """
    unreachable: None | str | Unset = UNSET
    """ Why the index could not be read, when `reachable` is false — the
    transport or parse error, for an administrator reading it beside the
    source URL they configured. Null otherwise.
     """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.content_attack_release import ContentAttackRelease  # noqa: PLC0415

        items = []
        for items_item_data in self.items:
            items_item = items_item_data.to_dict()
            items.append(items_item)

        reachable = self.reachable

        source_enabled = self.source_enabled

        unreachable: None | str | Unset
        if isinstance(self.unreachable, Unset):
            unreachable = UNSET
        else:
            unreachable = self.unreachable

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "items": items,
                "reachable": reachable,
                "sourceEnabled": source_enabled,
            }
        )
        if unreachable is not UNSET:
            field_dict["unreachable"] = unreachable

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.content_attack_release import ContentAttackRelease  # noqa: PLC0415

        d = dict(src_dict)
        items = []
        _items = d.pop("items")
        for items_item_data in _items:
            items_item = ContentAttackRelease.from_dict(items_item_data)

            items.append(items_item)

        reachable = d.pop("reachable")

        source_enabled = d.pop("sourceEnabled")

        def _parse_unreachable(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        unreachable = _parse_unreachable(d.pop("unreachable", UNSET))

        content_attack_release_list = cls(
            items=items,
            reachable=reachable,
            source_enabled=source_enabled,
            unreachable=unreachable,
        )

        content_attack_release_list.additional_properties = d
        return content_attack_release_list

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
