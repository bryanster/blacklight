from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast
from uuid import UUID

if TYPE_CHECKING:
    from ..models.presence_heartbeat_focus import PresenceHeartbeatFocus


T = TypeVar("T", bound="PresenceHeartbeat")


@_attrs_define
class PresenceHeartbeat:
    presence_id: UUID
    """ Client-generated UUIDv7 for this tab/window. """
    focus: PresenceHeartbeatFocus | Unset = UNSET

    def to_dict(self) -> dict[str, Any]:
        from ..models.presence_heartbeat_focus import PresenceHeartbeatFocus  # noqa: PLC0415

        presence_id = str(self.presence_id)

        focus: dict[str, Any] | Unset = UNSET
        if not isinstance(self.focus, Unset):
            focus = self.focus.to_dict()

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "presenceId": presence_id,
            }
        )
        if focus is not UNSET:
            field_dict["focus"] = focus

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.presence_heartbeat_focus import PresenceHeartbeatFocus  # noqa: PLC0415

        d = dict(src_dict)
        presence_id = UUID(d.pop("presenceId"))

        _focus = d.pop("focus", UNSET)
        focus: PresenceHeartbeatFocus | Unset
        if isinstance(_focus, Unset):
            focus = UNSET
        else:
            focus = PresenceHeartbeatFocus.from_dict(_focus)

        presence_heartbeat = cls(
            presence_id=presence_id,
            focus=focus,
        )

        return presence_heartbeat
