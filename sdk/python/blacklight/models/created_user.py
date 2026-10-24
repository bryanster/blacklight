from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast

if TYPE_CHECKING:
    from ..models.user import User


T = TypeVar("T", bound="CreatedUser")


@_attrs_define
class CreatedUser:
    """A newly created account and where to send the person it belongs to."""

    user: User
    """ One account, as user administration reports it. Compare `CurrentUser`,
    which is the caller describing themselves and carries their memberships
    and their MFA state; this is an administrator describing somebody else.

    What is *not* here is the point of it: there is no password hash, no
    authenticator secret, no recovery code and no session token, so no
    response built from this schema can carry one.
     """
    invite_url: str
    """ Where this installation is signed in to. There is no mail transport
    in this deployment, so nothing was sent: an administrator passes
    this on themselves, with the password they chose or with a note to
    use the single sign-on button.

    It carries no credential and grants nothing — it is a link to the
    front door, and anybody may already know it.
     """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.user import User  # noqa: PLC0415

        user = self.user.to_dict()

        invite_url = self.invite_url

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "user": user,
                "inviteUrl": invite_url,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.user import User  # noqa: PLC0415

        d = dict(src_dict)
        user = User.from_dict(d.pop("user"))

        invite_url = d.pop("inviteUrl")

        created_user = cls(
            user=user,
            invite_url=invite_url,
        )

        created_user.additional_properties = d
        return created_user

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
