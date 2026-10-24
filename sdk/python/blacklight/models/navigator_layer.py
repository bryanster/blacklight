from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.navigator_layer_filters import NavigatorLayerFilters
    from ..models.navigator_layer_gradient import NavigatorLayerGradient
    from ..models.navigator_layer_versions import NavigatorLayerVersions
    from ..models.navigator_legend_item import NavigatorLegendItem
    from ..models.navigator_technique import NavigatorTechnique


T = TypeVar("T", bound="NavigatorLayer")


@_attrs_define
class NavigatorLayer:
    name: str
    description: str
    domain: str
    versions: NavigatorLayerVersions
    filters: NavigatorLayerFilters
    gradient: NavigatorLayerGradient
    legend_items: list[NavigatorLegendItem]
    techniques: list[NavigatorTechnique]
    show_tactic_row_background: bool | Unset = UNSET
    tactic_row_background: str | Unset = UNSET
    select_techniques_across_tactics: bool | Unset = UNSET
    select_subtechniques_with_parent: bool | Unset = UNSET

    def to_dict(self) -> dict[str, Any]:
        from ..models.navigator_layer_filters import NavigatorLayerFilters  # noqa: PLC0415
        from ..models.navigator_layer_gradient import NavigatorLayerGradient  # noqa: PLC0415
        from ..models.navigator_layer_versions import NavigatorLayerVersions  # noqa: PLC0415
        from ..models.navigator_legend_item import NavigatorLegendItem  # noqa: PLC0415
        from ..models.navigator_technique import NavigatorTechnique  # noqa: PLC0415

        name = self.name

        description = self.description

        domain = self.domain

        versions = self.versions.to_dict()

        filters = self.filters.to_dict()

        gradient = self.gradient.to_dict()

        legend_items = []
        for legend_items_item_data in self.legend_items:
            legend_items_item = legend_items_item_data.to_dict()
            legend_items.append(legend_items_item)

        techniques = []
        for techniques_item_data in self.techniques:
            techniques_item = techniques_item_data.to_dict()
            techniques.append(techniques_item)

        show_tactic_row_background = self.show_tactic_row_background

        tactic_row_background = self.tactic_row_background

        select_techniques_across_tactics = self.select_techniques_across_tactics

        select_subtechniques_with_parent = self.select_subtechniques_with_parent

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "name": name,
                "description": description,
                "domain": domain,
                "versions": versions,
                "filters": filters,
                "gradient": gradient,
                "legendItems": legend_items,
                "techniques": techniques,
            }
        )
        if show_tactic_row_background is not UNSET:
            field_dict["showTacticRowBackground"] = show_tactic_row_background
        if tactic_row_background is not UNSET:
            field_dict["tacticRowBackground"] = tactic_row_background
        if select_techniques_across_tactics is not UNSET:
            field_dict["selectTechniquesAcrossTactics"] = select_techniques_across_tactics
        if select_subtechniques_with_parent is not UNSET:
            field_dict["selectSubtechniquesWithParent"] = select_subtechniques_with_parent

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.navigator_layer_filters import NavigatorLayerFilters  # noqa: PLC0415
        from ..models.navigator_layer_gradient import NavigatorLayerGradient  # noqa: PLC0415
        from ..models.navigator_layer_versions import NavigatorLayerVersions  # noqa: PLC0415
        from ..models.navigator_legend_item import NavigatorLegendItem  # noqa: PLC0415
        from ..models.navigator_technique import NavigatorTechnique  # noqa: PLC0415

        d = dict(src_dict)
        name = d.pop("name")

        description = d.pop("description")

        domain = d.pop("domain")

        versions = NavigatorLayerVersions.from_dict(d.pop("versions"))

        filters = NavigatorLayerFilters.from_dict(d.pop("filters"))

        gradient = NavigatorLayerGradient.from_dict(d.pop("gradient"))

        legend_items = []
        _legend_items = d.pop("legendItems")
        for legend_items_item_data in _legend_items:
            legend_items_item = NavigatorLegendItem.from_dict(legend_items_item_data)

            legend_items.append(legend_items_item)

        techniques = []
        _techniques = d.pop("techniques")
        for techniques_item_data in _techniques:
            techniques_item = NavigatorTechnique.from_dict(techniques_item_data)

            techniques.append(techniques_item)

        show_tactic_row_background = d.pop("showTacticRowBackground", UNSET)

        tactic_row_background = d.pop("tacticRowBackground", UNSET)

        select_techniques_across_tactics = d.pop("selectTechniquesAcrossTactics", UNSET)

        select_subtechniques_with_parent = d.pop("selectSubtechniquesWithParent", UNSET)

        navigator_layer = cls(
            name=name,
            description=description,
            domain=domain,
            versions=versions,
            filters=filters,
            gradient=gradient,
            legend_items=legend_items,
            techniques=techniques,
            show_tactic_row_background=show_tactic_row_background,
            tactic_row_background=tactic_row_background,
            select_techniques_across_tactics=select_techniques_across_tactics,
            select_subtechniques_with_parent=select_subtechniques_with_parent,
        )

        return navigator_layer
