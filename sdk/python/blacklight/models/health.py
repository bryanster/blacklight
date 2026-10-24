from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..models.health_state import HealthState
from typing import cast

if TYPE_CHECKING:
    from ..models.health_checks import HealthChecks


T = TypeVar("T", bound="Health")


@_attrs_define
class Health:
    """Health of the server and of everything it depends on. Reported with a 200
    when `status` is `ok` and a 503 when it is `error`, so a monitor that only
    looks at the status code still gets the right answer.

    """

    status: HealthState
    """ Outcome of a single health check, or of the report as a whole. """
    checks: HealthChecks
    """ One entry per dependency. Adding a dependency means adding its check
    here — a check that exists in code but not in this schema is invisible to
    every client.
     """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.health_checks import HealthChecks  # noqa: PLC0415

        status = self.status.value

        checks = self.checks.to_dict()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "status": status,
                "checks": checks,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.health_checks import HealthChecks  # noqa: PLC0415

        d = dict(src_dict)
        status = HealthState(d.pop("status"))

        checks = HealthChecks.from_dict(d.pop("checks"))

        health = cls(
            status=status,
            checks=checks,
        )

        health.additional_properties = d
        return health

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
