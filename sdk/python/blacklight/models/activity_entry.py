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
    from ..models.activity_entry_delta import ActivityEntryDelta


T = TypeVar("T", bound="ActivityEntry")


@_attrs_define
class ActivityEntry:
    """One row of the append-only activity log (M1-015). Drives the SSE feed
    (M4) and the report timeline (M6). There is no update or delete.

    """

    id: UUID
    """ UUIDv7. Sortable by creation; the stable tiebreaker when two rows share a timestamp. """
    verb: str
    """ What happened, spelled `object.past_tense_verb`. Examples:
    `session.login`, `session.login_failed`, `token.created`,
    `mfa.enrolled`.
     """
    object_type: str
    """ The kind of thing acted on (`user`, `session`, `service_token`, …). """
    object_id: str
    """ The identifier of the thing acted on. """
    at: datetime.datetime
    """ When it happened, UTC. """
    engagement_id: UUID | Unset = UNSET
    """ The engagement this entry belongs to. Absent on platform events
    (logins, token lifecycle, role changes).
     """
    actor_id: UUID | Unset = UNSET
    """ Who did it. Absent when the actor is unknown — a failed login that
    named no account.
     """
    delta: ActivityEntryDelta | Unset = UNSET
    """ Before/after for changed fields, already redacted. Never a password
    hash, token secret, TOTP secret, session token or recovery code.
    Absent when there is nothing useful to say beyond the verb.
     """
    revealed: bool | Unset = UNSET
    """ Whether the step this event is about has been revealed to blue.
    Present only for step-scoped objects (step, execution, evidence,
    comment) in a blind engagement. A blue caller's activity list
    omits rows where `revealed` is false (M4-008).
     """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.activity_entry_delta import ActivityEntryDelta  # noqa: PLC0415

        id = str(self.id)

        verb = self.verb

        object_type = self.object_type

        object_id = self.object_id

        at = self.at.isoformat()

        engagement_id: str | Unset = UNSET
        if not isinstance(self.engagement_id, Unset):
            engagement_id = str(self.engagement_id)

        actor_id: str | Unset = UNSET
        if not isinstance(self.actor_id, Unset):
            actor_id = str(self.actor_id)

        delta: dict[str, Any] | Unset = UNSET
        if not isinstance(self.delta, Unset):
            delta = self.delta.to_dict()

        revealed = self.revealed

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "verb": verb,
                "objectType": object_type,
                "objectId": object_id,
                "at": at,
            }
        )
        if engagement_id is not UNSET:
            field_dict["engagementId"] = engagement_id
        if actor_id is not UNSET:
            field_dict["actorId"] = actor_id
        if delta is not UNSET:
            field_dict["delta"] = delta
        if revealed is not UNSET:
            field_dict["revealed"] = revealed

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.activity_entry_delta import ActivityEntryDelta  # noqa: PLC0415

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        verb = d.pop("verb")

        object_type = d.pop("objectType")

        object_id = d.pop("objectId")

        at = datetime.datetime.fromisoformat(d.pop("at"))

        _engagement_id = d.pop("engagementId", UNSET)
        engagement_id: UUID | Unset
        if isinstance(_engagement_id, Unset):
            engagement_id = UNSET
        else:
            engagement_id = UUID(_engagement_id)

        _actor_id = d.pop("actorId", UNSET)
        actor_id: UUID | Unset
        if isinstance(_actor_id, Unset):
            actor_id = UNSET
        else:
            actor_id = UUID(_actor_id)

        _delta = d.pop("delta", UNSET)
        delta: ActivityEntryDelta | Unset
        if isinstance(_delta, Unset):
            delta = UNSET
        else:
            delta = ActivityEntryDelta.from_dict(_delta)

        revealed = d.pop("revealed", UNSET)

        activity_entry = cls(
            id=id,
            verb=verb,
            object_type=object_type,
            object_id=object_id,
            at=at,
            engagement_id=engagement_id,
            actor_id=actor_id,
            delta=delta,
            revealed=revealed,
        )

        activity_entry.additional_properties = d
        return activity_entry

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
