from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..models.burndown_interval import BurndownInterval
from typing import cast

if TYPE_CHECKING:
    from ..models.burndown_point import BurndownPoint
    from ..models.severity_snapshot import SeveritySnapshot


T = TypeVar("T", bound="AnalyticsBurndown")


@_attrs_define
class AnalyticsBurndown:
    interval: BurndownInterval
    """ Bucket granularity for the burndown chart. """
    points: list[BurndownPoint]
    severity: SeveritySnapshot
    blind_filtered: bool

    def to_dict(self) -> dict[str, Any]:
        from ..models.burndown_point import BurndownPoint  # noqa: PLC0415
        from ..models.severity_snapshot import SeveritySnapshot  # noqa: PLC0415

        interval = self.interval.value

        points = []
        for points_item_data in self.points:
            points_item = points_item_data.to_dict()
            points.append(points_item)

        severity = self.severity.to_dict()

        blind_filtered = self.blind_filtered

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "interval": interval,
                "points": points,
                "severity": severity,
                "blindFiltered": blind_filtered,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.burndown_point import BurndownPoint  # noqa: PLC0415
        from ..models.severity_snapshot import SeveritySnapshot  # noqa: PLC0415

        d = dict(src_dict)
        interval = BurndownInterval(d.pop("interval"))

        points = []
        _points = d.pop("points")
        for points_item_data in _points:
            points_item = BurndownPoint.from_dict(points_item_data)

            points.append(points_item)

        severity = SeveritySnapshot.from_dict(d.pop("severity"))

        blind_filtered = d.pop("blindFiltered")

        analytics_burndown = cls(
            interval=interval,
            points=points,
            severity=severity,
            blind_filtered=blind_filtered,
        )

        return analytics_burndown
