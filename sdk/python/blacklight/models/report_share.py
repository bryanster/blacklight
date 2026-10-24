from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast
from uuid import UUID
import datetime

if TYPE_CHECKING:
    from ..models.report_share_grant import ReportShareGrant


T = TypeVar("T", bound="ReportShare")


@_attrs_define
class ReportShare:
    id: UUID
    version_id: UUID
    created_by: str
    created_at: datetime.datetime
    password_protected: bool | Unset = UNSET
    """ Whether this share requires a password. """
    expires_at: datetime.datetime | None | Unset = UNSET
    revoked_at: datetime.datetime | None | Unset = UNSET
    label: None | str | Unset = UNSET
    max_grants: int | None | Unset = UNSET
    """ Maximum number of grants, or null for unlimited. """
    grant_count: int | Unset = UNSET
    """ Current number of non-revoked grants. """
    grants: list[ReportShareGrant] | Unset = UNSET
    """ Grants for this share. Present in list endpoints. """

    def to_dict(self) -> dict[str, Any]:
        from ..models.report_share_grant import ReportShareGrant  # noqa: PLC0415

        id = str(self.id)

        version_id = str(self.version_id)

        created_by = self.created_by

        created_at = self.created_at.isoformat()

        password_protected = self.password_protected

        expires_at: None | str | Unset
        if isinstance(self.expires_at, Unset):
            expires_at = UNSET
        elif isinstance(self.expires_at, datetime.datetime):
            expires_at = self.expires_at.isoformat()
        else:
            expires_at = self.expires_at

        revoked_at: None | str | Unset
        if isinstance(self.revoked_at, Unset):
            revoked_at = UNSET
        elif isinstance(self.revoked_at, datetime.datetime):
            revoked_at = self.revoked_at.isoformat()
        else:
            revoked_at = self.revoked_at

        label: None | str | Unset
        if isinstance(self.label, Unset):
            label = UNSET
        else:
            label = self.label

        max_grants: int | None | Unset
        if isinstance(self.max_grants, Unset):
            max_grants = UNSET
        else:
            max_grants = self.max_grants

        grant_count = self.grant_count

        grants: list[dict[str, Any]] | Unset = UNSET
        if not isinstance(self.grants, Unset):
            grants = []
            for grants_item_data in self.grants:
                grants_item = grants_item_data.to_dict()
                grants.append(grants_item)

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "id": id,
                "versionId": version_id,
                "createdBy": created_by,
                "createdAt": created_at,
            }
        )
        if password_protected is not UNSET:
            field_dict["passwordProtected"] = password_protected
        if expires_at is not UNSET:
            field_dict["expiresAt"] = expires_at
        if revoked_at is not UNSET:
            field_dict["revokedAt"] = revoked_at
        if label is not UNSET:
            field_dict["label"] = label
        if max_grants is not UNSET:
            field_dict["maxGrants"] = max_grants
        if grant_count is not UNSET:
            field_dict["grantCount"] = grant_count
        if grants is not UNSET:
            field_dict["grants"] = grants

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.report_share_grant import ReportShareGrant  # noqa: PLC0415

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        version_id = UUID(d.pop("versionId"))

        created_by = d.pop("createdBy")

        created_at = datetime.datetime.fromisoformat(d.pop("createdAt"))

        password_protected = d.pop("passwordProtected", UNSET)

        def _parse_expires_at(data: object) -> datetime.datetime | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            try:
                if not isinstance(data, str):
                    raise TypeError()
                expires_at_type_0 = datetime.datetime.fromisoformat(data)

                return expires_at_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(datetime.datetime | None | Unset, data)

        expires_at = _parse_expires_at(d.pop("expiresAt", UNSET))

        def _parse_revoked_at(data: object) -> datetime.datetime | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            try:
                if not isinstance(data, str):
                    raise TypeError()
                revoked_at_type_0 = datetime.datetime.fromisoformat(data)

                return revoked_at_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(datetime.datetime | None | Unset, data)

        revoked_at = _parse_revoked_at(d.pop("revokedAt", UNSET))

        def _parse_label(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        label = _parse_label(d.pop("label", UNSET))

        def _parse_max_grants(data: object) -> int | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(int | None | Unset, data)

        max_grants = _parse_max_grants(d.pop("maxGrants", UNSET))

        grant_count = d.pop("grantCount", UNSET)

        _grants = d.pop("grants", UNSET)
        grants: list[ReportShareGrant] | Unset = UNSET
        if _grants is not UNSET:
            grants = []
            for grants_item_data in _grants:
                grants_item = ReportShareGrant.from_dict(grants_item_data)

                grants.append(grants_item)

        report_share = cls(
            id=id,
            version_id=version_id,
            created_by=created_by,
            created_at=created_at,
            password_protected=password_protected,
            expires_at=expires_at,
            revoked_at=revoked_at,
            label=label,
            max_grants=max_grants,
            grant_count=grant_count,
            grants=grants,
        )

        return report_share
