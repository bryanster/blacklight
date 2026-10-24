from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast

if TYPE_CHECKING:
    from ..models.service_token import ServiceToken


T = TypeVar("T", bound="CreatedServiceToken")


@_attrs_define
class CreatedServiceToken:
    """A token that has just been created: the row, plus the one and only copy
    of its secret. Nothing else in this API ever carries `token`.

    """

    service_token: ServiceToken
    """ One service token, as anybody but its creator ever sees it: everything
    about the credential except the credential.
     """
    token: str
    """ The credential, spelled `bl_<prefix>_<secret>`. Send it as
    `Authorization: Bearer <token>`.

    Show it to the person immediately and offer a way to copy it. There
    is no endpoint that reads it back, because the server keeps only a
    hash of it — a lost token is replaced, never recovered.
     """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.service_token import ServiceToken  # noqa: PLC0415

        service_token = self.service_token.to_dict()

        token = self.token

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "serviceToken": service_token,
                "token": token,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.service_token import ServiceToken  # noqa: PLC0415

        d = dict(src_dict)
        service_token = ServiceToken.from_dict(d.pop("serviceToken"))

        token = d.pop("token")

        created_service_token = cls(
            service_token=service_token,
            token=token,
        )

        created_service_token.additional_properties = d
        return created_service_token

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
