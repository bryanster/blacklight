from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.compare_row import CompareRow
    from ..models.pin_mismatch import PinMismatch


T = TypeVar("T", bound="AnalyticsCompare")


@_attrs_define
class AnalyticsCompare:
    rows: list[CompareRow]
    improved: int
    regressed: int
    unchanged: int
    newly_attempted: int
    no_longer_attempted: int
    incomparable: int
    baseline_blind_filtered: bool
    current_blind_filtered: bool
    pin_mismatch: PinMismatch | Unset = UNSET

    def to_dict(self) -> dict[str, Any]:
        from ..models.compare_row import CompareRow  # noqa: PLC0415
        from ..models.pin_mismatch import PinMismatch  # noqa: PLC0415

        rows = []
        for rows_item_data in self.rows:
            rows_item = rows_item_data.to_dict()
            rows.append(rows_item)

        improved = self.improved

        regressed = self.regressed

        unchanged = self.unchanged

        newly_attempted = self.newly_attempted

        no_longer_attempted = self.no_longer_attempted

        incomparable = self.incomparable

        baseline_blind_filtered = self.baseline_blind_filtered

        current_blind_filtered = self.current_blind_filtered

        pin_mismatch: dict[str, Any] | Unset = UNSET
        if not isinstance(self.pin_mismatch, Unset):
            pin_mismatch = self.pin_mismatch.to_dict()

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "rows": rows,
                "improved": improved,
                "regressed": regressed,
                "unchanged": unchanged,
                "newlyAttempted": newly_attempted,
                "noLongerAttempted": no_longer_attempted,
                "incomparable": incomparable,
                "baselineBlindFiltered": baseline_blind_filtered,
                "currentBlindFiltered": current_blind_filtered,
            }
        )
        if pin_mismatch is not UNSET:
            field_dict["pinMismatch"] = pin_mismatch

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.compare_row import CompareRow  # noqa: PLC0415
        from ..models.pin_mismatch import PinMismatch  # noqa: PLC0415

        d = dict(src_dict)
        rows = []
        _rows = d.pop("rows")
        for rows_item_data in _rows:
            rows_item = CompareRow.from_dict(rows_item_data)

            rows.append(rows_item)

        improved = d.pop("improved")

        regressed = d.pop("regressed")

        unchanged = d.pop("unchanged")

        newly_attempted = d.pop("newlyAttempted")

        no_longer_attempted = d.pop("noLongerAttempted")

        incomparable = d.pop("incomparable")

        baseline_blind_filtered = d.pop("baselineBlindFiltered")

        current_blind_filtered = d.pop("currentBlindFiltered")

        _pin_mismatch = d.pop("pinMismatch", UNSET)
        pin_mismatch: PinMismatch | Unset
        if isinstance(_pin_mismatch, Unset):
            pin_mismatch = UNSET
        else:
            pin_mismatch = PinMismatch.from_dict(_pin_mismatch)

        analytics_compare = cls(
            rows=rows,
            improved=improved,
            regressed=regressed,
            unchanged=unchanged,
            newly_attempted=newly_attempted,
            no_longer_attempted=no_longer_attempted,
            incomparable=incomparable,
            baseline_blind_filtered=baseline_blind_filtered,
            current_blind_filtered=current_blind_filtered,
            pin_mismatch=pin_mismatch,
        )

        return analytics_compare
