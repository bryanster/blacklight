from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.navigator_metadata import NavigatorMetadata


T = TypeVar("T", bound="NavigatorTechnique")


@_attrs_define
class NavigatorTechnique:
    technique_id: str
    score: int
    color: str
    enabled: bool
    comment: str | Unset = UNSET
    metadata: list[NavigatorMetadata] | Unset = UNSET
    show_subtechniques: bool | Unset = UNSET

    def to_dict(self) -> dict[str, Any]:
        from ..models.navigator_metadata import NavigatorMetadata  # noqa: PLC0415

        technique_id = self.technique_id

        score = self.score

        color = self.color

        enabled = self.enabled

        comment = self.comment

        metadata: list[dict[str, Any]] | Unset = UNSET
        if not isinstance(self.metadata, Unset):
            metadata = []
            for metadata_item_data in self.metadata:
                metadata_item = metadata_item_data.to_dict()
                metadata.append(metadata_item)

        show_subtechniques = self.show_subtechniques

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "techniqueID": technique_id,
                "score": score,
                "color": color,
                "enabled": enabled,
            }
        )
        if comment is not UNSET:
            field_dict["comment"] = comment
        if metadata is not UNSET:
            field_dict["metadata"] = metadata
        if show_subtechniques is not UNSET:
            field_dict["showSubtechniques"] = show_subtechniques

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.navigator_metadata import NavigatorMetadata  # noqa: PLC0415

        d = dict(src_dict)
        technique_id = d.pop("techniqueID")

        score = d.pop("score")

        color = d.pop("color")

        enabled = d.pop("enabled")

        comment = d.pop("comment", UNSET)

        _metadata = d.pop("metadata", UNSET)
        metadata: list[NavigatorMetadata] | Unset = UNSET
        if _metadata is not UNSET:
            metadata = []
            for metadata_item_data in _metadata:
                metadata_item = NavigatorMetadata.from_dict(metadata_item_data)

                metadata.append(metadata_item)

        show_subtechniques = d.pop("showSubtechniques", UNSET)

        navigator_technique = cls(
            technique_id=technique_id,
            score=score,
            color=color,
            enabled=enabled,
            comment=comment,
            metadata=metadata,
            show_subtechniques=show_subtechniques,
        )

        return navigator_technique
