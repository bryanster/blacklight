from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast

if TYPE_CHECKING:
    from ..models.content_import_issue import ContentImportIssue


T = TypeVar("T", bound="ContentImportReport")


@_attrs_define
class ContentImportReport:
    """Result of a synchronous custom/v1 import or dry-run. Async jobs surface
    the same counts in the job message and activity delta.

    """

    dry_run: bool
    format_: str
    """ Resolved format after auto-detection. """
    procedures_created: int
    procedures_updated: int
    notes_created: int
    notes_updated: int
    detections_created: int
    detections_updated: int
    warnings: list[ContentImportIssue]
    errors: list[ContentImportIssue]

    def to_dict(self) -> dict[str, Any]:
        from ..models.content_import_issue import ContentImportIssue  # noqa: PLC0415

        dry_run = self.dry_run

        format_ = self.format_

        procedures_created = self.procedures_created

        procedures_updated = self.procedures_updated

        notes_created = self.notes_created

        notes_updated = self.notes_updated

        detections_created = self.detections_created

        detections_updated = self.detections_updated

        warnings = []
        for warnings_item_data in self.warnings:
            warnings_item = warnings_item_data.to_dict()
            warnings.append(warnings_item)

        errors = []
        for errors_item_data in self.errors:
            errors_item = errors_item_data.to_dict()
            errors.append(errors_item)

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "dryRun": dry_run,
                "format": format_,
                "proceduresCreated": procedures_created,
                "proceduresUpdated": procedures_updated,
                "notesCreated": notes_created,
                "notesUpdated": notes_updated,
                "detectionsCreated": detections_created,
                "detectionsUpdated": detections_updated,
                "warnings": warnings,
                "errors": errors,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.content_import_issue import ContentImportIssue  # noqa: PLC0415

        d = dict(src_dict)
        dry_run = d.pop("dryRun")

        format_ = d.pop("format")

        procedures_created = d.pop("proceduresCreated")

        procedures_updated = d.pop("proceduresUpdated")

        notes_created = d.pop("notesCreated")

        notes_updated = d.pop("notesUpdated")

        detections_created = d.pop("detectionsCreated")

        detections_updated = d.pop("detectionsUpdated")

        warnings = []
        _warnings = d.pop("warnings")
        for warnings_item_data in _warnings:
            warnings_item = ContentImportIssue.from_dict(warnings_item_data)

            warnings.append(warnings_item)

        errors = []
        _errors = d.pop("errors")
        for errors_item_data in _errors:
            errors_item = ContentImportIssue.from_dict(errors_item_data)

            errors.append(errors_item)

        content_import_report = cls(
            dry_run=dry_run,
            format_=format_,
            procedures_created=procedures_created,
            procedures_updated=procedures_updated,
            notes_created=notes_created,
            notes_updated=notes_updated,
            detections_created=detections_created,
            detections_updated=detections_updated,
            warnings=warnings,
            errors=errors,
        )

        return content_import_report
