from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..models.burndown_interval import BurndownInterval
from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.burndown_point import BurndownPoint
    from ..models.severity_snapshot import SeveritySnapshot


T = TypeVar("T", bound="BurndownAnalysis")


@_attrs_define
class BurndownAnalysis:
    severity: SeveritySnapshot
    blind_filtered: bool
    interval: BurndownInterval | Unset = UNSET
    """ Bucket granularity for the burndown chart. """
    points: list[BurndownPoint] | Unset = UNSET

    def to_dict(self) -> dict[str, Any]:
        from ..models.burndown_point import BurndownPoint  # noqa: PLC0415
        from ..models.severity_snapshot import SeveritySnapshot  # noqa: PLC0415

        severity = self.severity.to_dict()

        blind_filtered = self.blind_filtered

        interval: str | Unset = UNSET
        if not isinstance(self.interval, Unset):
            interval = self.interval.value

        points: list[dict[str, Any]] | Unset = UNSET
        if not isinstance(self.points, Unset):
            points = []
            for points_item_data in self.points:
                points_item = points_item_data.to_dict()
                points.append(points_item)

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "severity": severity,
                "blindFiltered": blind_filtered,
            }
        )
        if interval is not UNSET:
            field_dict["interval"] = interval
        if points is not UNSET:
            field_dict["points"] = points

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.burndown_point import BurndownPoint  # noqa: PLC0415
        from ..models.severity_snapshot import SeveritySnapshot  # noqa: PLC0415

        d = dict(src_dict)
        severity = SeveritySnapshot.from_dict(d.pop("severity"))

        blind_filtered = d.pop("blindFiltered")

        _interval = d.pop("interval", UNSET)
        interval: BurndownInterval | Unset
        if isinstance(_interval, Unset):
            interval = UNSET
        else:
            interval = BurndownInterval(_interval)

        _points = d.pop("points", UNSET)
        points: list[BurndownPoint] | Unset = UNSET
        if _points is not UNSET:
            points = []
            for points_item_data in _points:
                points_item = BurndownPoint.from_dict(points_item_data)

                points.append(points_item)

        burndown_analysis = cls(
            severity=severity,
            blind_filtered=blind_filtered,
            interval=interval,
            points=points,
        )

        return burndown_analysis
