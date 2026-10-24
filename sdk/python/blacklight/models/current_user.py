from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..models.platform_role import PlatformRole
from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.engagement_membership import EngagementMembership
    from ..models.mfa_state import MFAState


T = TypeVar("T", bound="CurrentUser")


@_attrs_define
class CurrentUser:
    """The caller, as `GET /auth/me` reports them. It carries nothing about the
    session token: the cookie is the only place that value exists outside the
    client.

    """

    id: str
    """ The user's identifier — a UUIDv7, as every identifier here is. """
    email: str
    """ The address as it was typed when the account was created. """
    display_name: str
    """ The name to show for this person. """
    platform_role: PlatformRole
    """ What somebody may do to this installation: `admin` manages users, content
    and every engagement; `member` takes part in the engagements they belong
    to. What they may do *inside* one is `EngagementRole`, and the two are
    deliberately not the same vocabulary.
     """
    mfa: MFAState
    """ Where this person and this session stand on multi-factor authentication.
    Whether an administrator *requires* it and whether this session has
    *satisfied* it are separate facts — conflating them is the hole M1-008
    closes.
     """
    memberships: list[EngagementMembership]
    """ Every engagement this person belongs to, and their role in it. An
    administrator's list is still only what they are a *member* of —
    their reach beyond it comes from `platformRole`, not from here.
     """
    csrf_token: str | Unset = UNSET
    """ The double-submit CSRF token for this session (M1-005) — the same
    value as the `bl_csrf` cookie, which is where a browser client
    should read it from.

    It is here for a client that has no cookie jar to read, and it is
    not a secret in the way the session token is: it authorizes nothing
    on its own. Absent for a caller authenticated by a service token,
    which is not subject to the check.
     """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.engagement_membership import EngagementMembership  # noqa: PLC0415
        from ..models.mfa_state import MFAState  # noqa: PLC0415

        id = self.id

        email = self.email

        display_name = self.display_name

        platform_role = self.platform_role.value

        mfa = self.mfa.to_dict()

        memberships = []
        for memberships_item_data in self.memberships:
            memberships_item = memberships_item_data.to_dict()
            memberships.append(memberships_item)

        csrf_token = self.csrf_token

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "email": email,
                "displayName": display_name,
                "platformRole": platform_role,
                "mfa": mfa,
                "memberships": memberships,
            }
        )
        if csrf_token is not UNSET:
            field_dict["csrfToken"] = csrf_token

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.engagement_membership import EngagementMembership  # noqa: PLC0415
        from ..models.mfa_state import MFAState  # noqa: PLC0415

        d = dict(src_dict)
        id = d.pop("id")

        email = d.pop("email")

        display_name = d.pop("displayName")

        platform_role = PlatformRole(d.pop("platformRole"))

        mfa = MFAState.from_dict(d.pop("mfa"))

        memberships = []
        _memberships = d.pop("memberships")
        for memberships_item_data in _memberships:
            memberships_item = EngagementMembership.from_dict(memberships_item_data)

            memberships.append(memberships_item)

        csrf_token = d.pop("csrfToken", UNSET)

        current_user = cls(
            id=id,
            email=email,
            display_name=display_name,
            platform_role=platform_role,
            mfa=mfa,
            memberships=memberships,
            csrf_token=csrf_token,
        )

        current_user.additional_properties = d
        return current_user

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
