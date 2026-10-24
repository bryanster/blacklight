from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..models.content_source_kind import ContentSourceKind
from ..models.content_source_status import ContentSourceStatus
from ..types import UNSET, Unset
from typing import cast
from uuid import UUID
import datetime

if TYPE_CHECKING:
    from ..models.content_sync_job_summary import ContentSyncJobSummary


T = TypeVar("T", bound="ContentSourceDetail")


@_attrs_define
class ContentSourceDetail:
    id: UUID
    kind: ContentSourceKind
    """ Closed vocabulary of content libraries. New kinds are a migration, not
    a string somebody passed to an API. There is no create-source endpoint
    in M2 — only the seeded rows.
     """
    name: str
    url: str
    """ Default HTTPS archive / bundle base URL. Empty for custom. """
    ref: str
    """ Adapter-specific ref pattern or branch/tag hint. Empty when unused. """
    enabled: bool
    """ Soft switch. Disabled sources stay on disk; browse/search/pickers
    omit their objects and new references are refused.
     """
    status: ContentSourceStatus
    """ Operational state of a source as a whole. Independent of `enabled`: a
    disabled source can still be idle, and an enabled one can be in error.
     """
    item_count: int
    """ Bookkeeping count of objects currently held for this source. """
    error: str
    """ Last error message from a failed job. Empty when none. """
    license_spdx: str
    """ SPDX license identifier. """
    license_name: str
    license_url: str
    attribution: str
    """ Attribution text shown in UI detail and export headers. """
    created_at: datetime.datetime
    updated_at: datetime.datetime
    last_synced_at: datetime.datetime | Unset = UNSET
    """ When this source last finished a successful sync. Absent if never. """
    last_job: ContentSyncJobSummary | Unset = UNSET
    """ The most recent job for a source, as detail surfaces it. Prefer
    `ContentSyncJob` when reading a job by id.
     """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.content_sync_job_summary import ContentSyncJobSummary  # noqa: PLC0415

        id = str(self.id)

        kind = self.kind.value

        name = self.name

        url = self.url

        ref = self.ref

        enabled = self.enabled

        status = self.status.value

        item_count = self.item_count

        error = self.error

        license_spdx = self.license_spdx

        license_name = self.license_name

        license_url = self.license_url

        attribution = self.attribution

        created_at = self.created_at.isoformat()

        updated_at = self.updated_at.isoformat()

        last_synced_at: str | Unset = UNSET
        if not isinstance(self.last_synced_at, Unset):
            last_synced_at = self.last_synced_at.isoformat()

        last_job: dict[str, Any] | Unset = UNSET
        if not isinstance(self.last_job, Unset):
            last_job = self.last_job.to_dict()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "kind": kind,
                "name": name,
                "url": url,
                "ref": ref,
                "enabled": enabled,
                "status": status,
                "itemCount": item_count,
                "error": error,
                "licenseSpdx": license_spdx,
                "licenseName": license_name,
                "licenseUrl": license_url,
                "attribution": attribution,
                "createdAt": created_at,
                "updatedAt": updated_at,
            }
        )
        if last_synced_at is not UNSET:
            field_dict["lastSyncedAt"] = last_synced_at
        if last_job is not UNSET:
            field_dict["lastJob"] = last_job

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.content_sync_job_summary import ContentSyncJobSummary  # noqa: PLC0415

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        kind = ContentSourceKind(d.pop("kind"))

        name = d.pop("name")

        url = d.pop("url")

        ref = d.pop("ref")

        enabled = d.pop("enabled")

        status = ContentSourceStatus(d.pop("status"))

        item_count = d.pop("itemCount")

        error = d.pop("error")

        license_spdx = d.pop("licenseSpdx")

        license_name = d.pop("licenseName")

        license_url = d.pop("licenseUrl")

        attribution = d.pop("attribution")

        created_at = datetime.datetime.fromisoformat(d.pop("createdAt"))

        updated_at = datetime.datetime.fromisoformat(d.pop("updatedAt"))

        _last_synced_at = d.pop("lastSyncedAt", UNSET)
        last_synced_at: datetime.datetime | Unset
        if isinstance(_last_synced_at, Unset):
            last_synced_at = UNSET
        else:
            last_synced_at = datetime.datetime.fromisoformat(_last_synced_at)

        _last_job = d.pop("lastJob", UNSET)
        last_job: ContentSyncJobSummary | Unset
        if isinstance(_last_job, Unset):
            last_job = UNSET
        else:
            last_job = ContentSyncJobSummary.from_dict(_last_job)

        content_source_detail = cls(
            id=id,
            kind=kind,
            name=name,
            url=url,
            ref=ref,
            enabled=enabled,
            status=status,
            item_count=item_count,
            error=error,
            license_spdx=license_spdx,
            license_name=license_name,
            license_url=license_url,
            attribution=attribution,
            created_at=created_at,
            updated_at=updated_at,
            last_synced_at=last_synced_at,
            last_job=last_job,
        )

        content_source_detail.additional_properties = d
        return content_source_detail

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
