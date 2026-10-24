from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..models.login_status import LoginStatus
from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.current_user import CurrentUser


T = TypeVar("T", bound="LoginResult")


@_attrs_define
class LoginResult:
    """The outcome of a successful `POST /auth/login`."""

    status: LoginStatus
    """ What a successful `POST /auth/login` established. A client must branch on
    this rather than assume that 200 means signed in.
     """
    user: CurrentUser | Unset = UNSET
    """ The caller, as `GET /auth/me` reports them. It carries nothing about the
    session token: the cookie is the only place that value exists outside the
    client.
     """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.current_user import CurrentUser  # noqa: PLC0415

        status = self.status.value

        user: dict[str, Any] | Unset = UNSET
        if not isinstance(self.user, Unset):
            user = self.user.to_dict()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "status": status,
            }
        )
        if user is not UNSET:
            field_dict["user"] = user

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.current_user import CurrentUser  # noqa: PLC0415

        d = dict(src_dict)
        status = LoginStatus(d.pop("status"))

        _user = d.pop("user", UNSET)
        user: CurrentUser | Unset
        if isinstance(_user, Unset):
            user = UNSET
        else:
            user = CurrentUser.from_dict(_user)

        login_result = cls(
            status=status,
            user=user,
        )

        login_result.additional_properties = d
        return login_result

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
