from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast

if TYPE_CHECKING:
    from ..models.category_bucket import CategoryBucket


T = TypeVar("T", bound="TacticCoverageRow")


@_attrs_define
class TacticCoverageRow:
    tactic_id: str
    tactic_name: str
    attempted_techniques: int
    matrix_techniques: int
    categories: list[CategoryBucket]

    def to_dict(self) -> dict[str, Any]:
        from ..models.category_bucket import CategoryBucket  # noqa: PLC0415

        tactic_id = self.tactic_id

        tactic_name = self.tactic_name

        attempted_techniques = self.attempted_techniques

        matrix_techniques = self.matrix_techniques

        categories = []
        for categories_item_data in self.categories:
            categories_item = categories_item_data.to_dict()
            categories.append(categories_item)

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "tacticId": tactic_id,
                "tacticName": tactic_name,
                "attemptedTechniques": attempted_techniques,
                "matrixTechniques": matrix_techniques,
                "categories": categories,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.category_bucket import CategoryBucket  # noqa: PLC0415

        d = dict(src_dict)
        tactic_id = d.pop("tacticId")

        tactic_name = d.pop("tacticName")

        attempted_techniques = d.pop("attemptedTechniques")

        matrix_techniques = d.pop("matrixTechniques")

        categories = []
        _categories = d.pop("categories")
        for categories_item_data in _categories:
            categories_item = CategoryBucket.from_dict(categories_item_data)

            categories.append(categories_item)

        tactic_coverage_row = cls(
            tactic_id=tactic_id,
            tactic_name=tactic_name,
            attempted_techniques=attempted_techniques,
            matrix_techniques=matrix_techniques,
            categories=categories,
        )

        return tactic_coverage_row
