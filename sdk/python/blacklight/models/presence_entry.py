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
    from ..models.presence_entry_focus import PresenceEntryFocus


T = TypeVar("T", bound="PresenceEntry")


@_attrs_define
class PresenceEntry:
    user_id: UUID
    """ The user's platform id. """
    display_name: str
    """ The user's display name. """
    last_seen_at: datetime.datetime
    """ When this user's most recent heartbeat arrived. """
    tab_count: int
    """ Number of active tabs/windows for this user in this engagement. """
    focus: PresenceEntryFocus | Unset = UNSET

    def to_dict(self) -> dict[str, Any]:
        from ..models.presence_entry_focus import PresenceEntryFocus  # noqa: PLC0415

        user_id = str(self.user_id)

        display_name = self.display_name

        last_seen_at = self.last_seen_at.isoformat()

        tab_count = self.tab_count

        focus: dict[str, Any] | Unset = UNSET
        if not isinstance(self.focus, Unset):
            focus = self.focus.to_dict()

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "userId": user_id,
                "displayName": display_name,
                "lastSeenAt": last_seen_at,
                "tabCount": tab_count,
            }
        )
        if focus is not UNSET:
            field_dict["focus"] = focus

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.presence_entry_focus import PresenceEntryFocus  # noqa: PLC0415

        d = dict(src_dict)
        user_id = UUID(d.pop("userId"))

        display_name = d.pop("displayName")

        last_seen_at = datetime.datetime.fromisoformat(d.pop("lastSeenAt"))

        tab_count = d.pop("tabCount")

        _focus = d.pop("focus", UNSET)
        focus: PresenceEntryFocus | Unset
        if isinstance(_focus, Unset):
            focus = UNSET
        else:
            focus = PresenceEntryFocus.from_dict(_focus)

        presence_entry = cls(
            user_id=user_id,
            display_name=display_name,
            last_seen_at=last_seen_at,
            tab_count=tab_count,
            focus=focus,
        )

        return presence_entry
