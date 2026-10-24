from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast

if TYPE_CHECKING:
    from ..models.distribution_result import DistributionResult


T = TypeVar("T", bound="AnalyticsDistribution")


@_attrs_define
class AnalyticsDistribution:
    category: DistributionResult
    protection: DistributionResult
    outcome: DistributionResult
    modifier: DistributionResult
    blind_filtered: bool

    def to_dict(self) -> dict[str, Any]:
        from ..models.distribution_result import DistributionResult  # noqa: PLC0415

        category = self.category.to_dict()

        protection = self.protection.to_dict()

        outcome = self.outcome.to_dict()

        modifier = self.modifier.to_dict()

        blind_filtered = self.blind_filtered

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "category": category,
                "protection": protection,
                "outcome": outcome,
                "modifier": modifier,
                "blindFiltered": blind_filtered,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.distribution_result import DistributionResult  # noqa: PLC0415

        d = dict(src_dict)
        category = DistributionResult.from_dict(d.pop("category"))

        protection = DistributionResult.from_dict(d.pop("protection"))

        outcome = DistributionResult.from_dict(d.pop("outcome"))

        modifier = DistributionResult.from_dict(d.pop("modifier"))

        blind_filtered = d.pop("blindFiltered")

        analytics_distribution = cls(
            category=category,
            protection=protection,
            outcome=outcome,
            modifier=modifier,
            blind_filtered=blind_filtered,
        )

        return analytics_distribution
