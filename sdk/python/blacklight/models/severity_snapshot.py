from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast

if TYPE_CHECKING:
    from ..models.severity_bucket import SeverityBucket


T = TypeVar("T", bound="SeveritySnapshot")


@_attrs_define
class SeveritySnapshot:
    buckets: list[SeverityBucket]

    def to_dict(self) -> dict[str, Any]:
        from ..models.severity_bucket import SeverityBucket  # noqa: PLC0415

        buckets = []
        for buckets_item_data in self.buckets:
            buckets_item = buckets_item_data.to_dict()
            buckets.append(buckets_item)

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "buckets": buckets,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.severity_bucket import SeverityBucket  # noqa: PLC0415

        d = dict(src_dict)
        buckets = []
        _buckets = d.pop("buckets")
        for buckets_item_data in _buckets:
            buckets_item = SeverityBucket.from_dict(buckets_item_data)

            buckets.append(buckets_item)

        severity_snapshot = cls(
            buckets=buckets,
        )

        return severity_snapshot
