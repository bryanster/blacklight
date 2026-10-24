from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast

if TYPE_CHECKING:
    from ..models.content_custom_export_meta import ContentCustomExportMeta
    from ..models.content_detection_rule import ContentDetectionRule
    from ..models.content_note import ContentNote
    from ..models.content_procedure_template import ContentProcedureTemplate


T = TypeVar("T", bound="ContentCustomExport")


@_attrs_define
class ContentCustomExport:
    """Export document for custom content. Shape is accepted by the v1/custom
    import path (M2-012) or documented as the delta if the importer needs
    a thin adapter. Empty arrays mean that family was omitted or empty.

    """

    meta: ContentCustomExportMeta
    """ License/attribution header for a custom content export. """
    procedure_templates: list[ContentProcedureTemplate]
    detection_rules: list[ContentDetectionRule]
    notes: list[ContentNote]
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.content_custom_export_meta import ContentCustomExportMeta  # noqa: PLC0415
        from ..models.content_detection_rule import ContentDetectionRule  # noqa: PLC0415
        from ..models.content_note import ContentNote  # noqa: PLC0415
        from ..models.content_procedure_template import ContentProcedureTemplate  # noqa: PLC0415

        meta = self.meta.to_dict()

        procedure_templates = []
        for procedure_templates_item_data in self.procedure_templates:
            procedure_templates_item = procedure_templates_item_data.to_dict()
            procedure_templates.append(procedure_templates_item)

        detection_rules = []
        for detection_rules_item_data in self.detection_rules:
            detection_rules_item = detection_rules_item_data.to_dict()
            detection_rules.append(detection_rules_item)

        notes = []
        for notes_item_data in self.notes:
            notes_item = notes_item_data.to_dict()
            notes.append(notes_item)

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "meta": meta,
                "procedureTemplates": procedure_templates,
                "detectionRules": detection_rules,
                "notes": notes,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.content_custom_export_meta import ContentCustomExportMeta  # noqa: PLC0415
        from ..models.content_detection_rule import ContentDetectionRule  # noqa: PLC0415
        from ..models.content_note import ContentNote  # noqa: PLC0415
        from ..models.content_procedure_template import ContentProcedureTemplate  # noqa: PLC0415

        d = dict(src_dict)
        meta = ContentCustomExportMeta.from_dict(d.pop("meta"))

        procedure_templates = []
        _procedure_templates = d.pop("procedureTemplates")
        for procedure_templates_item_data in _procedure_templates:
            procedure_templates_item = ContentProcedureTemplate.from_dict(procedure_templates_item_data)

            procedure_templates.append(procedure_templates_item)

        detection_rules = []
        _detection_rules = d.pop("detectionRules")
        for detection_rules_item_data in _detection_rules:
            detection_rules_item = ContentDetectionRule.from_dict(detection_rules_item_data)

            detection_rules.append(detection_rules_item)

        notes = []
        _notes = d.pop("notes")
        for notes_item_data in _notes:
            notes_item = ContentNote.from_dict(notes_item_data)

            notes.append(notes_item)

        content_custom_export = cls(
            meta=meta,
            procedure_templates=procedure_templates,
            detection_rules=detection_rules,
            notes=notes,
        )

        content_custom_export.additional_properties = d
        return content_custom_export

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
