from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast

if TYPE_CHECKING:
    from ..models.sso_provider import SSOProvider


T = TypeVar("T", bound="AuthProviders")


@_attrs_define
class AuthProviders:
    """What the login page may offer. It is deliberately a list rather than a
    set of booleans: SAML sits beside OIDC in it (M1-010), and a page that
    renders this array needed no change when it arrived.

    """

    password: bool
    """ Whether local email-and-password sign-in is offered. Always true
    today; it is stated rather than assumed so that a deployment which
    later turns it off has somewhere to say so.
     """
    sso: list[SSOProvider]
    """ Every single sign-on provider that is configured **and** reachable
    right now. A configured provider that cannot be discovered is absent
    rather than listed-and-broken.
     """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.sso_provider import SSOProvider  # noqa: PLC0415

        password = self.password

        sso = []
        for sso_item_data in self.sso:
            sso_item = sso_item_data.to_dict()
            sso.append(sso_item)

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "password": password,
                "sso": sso,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.sso_provider import SSOProvider  # noqa: PLC0415

        d = dict(src_dict)
        password = d.pop("password")

        sso = []
        _sso = d.pop("sso")
        for sso_item_data in _sso:
            sso_item = SSOProvider.from_dict(sso_item_data)

            sso.append(sso_item)

        auth_providers = cls(
            password=password,
            sso=sso,
        )

        auth_providers.additional_properties = d
        return auth_providers

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
