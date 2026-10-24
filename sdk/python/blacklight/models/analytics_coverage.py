from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast

if TYPE_CHECKING:
    from ..models.tactic_coverage import TacticCoverage
    from ..models.technique_coverage import TechniqueCoverage


T = TypeVar("T", bound="AnalyticsCoverage")


@_attrs_define
class AnalyticsCoverage:
    techniques: TechniqueCoverage
    tactics: TacticCoverage
    blind_filtered: bool

    def to_dict(self) -> dict[str, Any]:
        from ..models.tactic_coverage import TacticCoverage  # noqa: PLC0415
        from ..models.technique_coverage import TechniqueCoverage  # noqa: PLC0415

        techniques = self.techniques.to_dict()

        tactics = self.tactics.to_dict()

        blind_filtered = self.blind_filtered

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "techniques": techniques,
                "tactics": tactics,
                "blindFiltered": blind_filtered,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.tactic_coverage import TacticCoverage  # noqa: PLC0415
        from ..models.technique_coverage import TechniqueCoverage  # noqa: PLC0415

        d = dict(src_dict)
        techniques = TechniqueCoverage.from_dict(d.pop("techniques"))

        tactics = TacticCoverage.from_dict(d.pop("tactics"))

        blind_filtered = d.pop("blindFiltered")

        analytics_coverage = cls(
            techniques=techniques,
            tactics=tactics,
            blind_filtered=blind_filtered,
        )

        return analytics_coverage
