from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from typing import cast
from uuid import UUID
import datetime

if TYPE_CHECKING:
    from ..models.content_emulation_plan_step_procedure import ContentEmulationPlanStepProcedure


T = TypeVar("T", bound="ContentEmulationPlanStep")


@_attrs_define
class ContentEmulationPlanStep:
    """One ordered step under a CTID emulation plan. `ordinal` is 1-based
    document order from the upstream plan YAML (dense, unique per plan).
    Upstream `procedure_step` labels (for example `2.1`) live inside
    `procedure` when present and are not used as the ordinal.

    """

    id: UUID
    source_id: UUID
    version: str
    """ Version token. CTID is rolling-head and always `current`. """
    plan_id: UUID
    """ Surrogate id of the parent plan. """
    ordinal: int
    """ 1-based position under the plan (document order). """
    external_id: str
    """ Stable id within the source. Prefer upstream step `id`; otherwise
    `{plan_external_id}/{ordinal}` (documented in `docs/content-ctid.md`).
     """
    name: str
    description: str
    technique_external_id: str
    """ ATT&CK technique external id when upstream provides one (for
    example `T1566.001`). Empty string when missing — allowed; M3 pin
    resolve treats empty as no technique.
     """
    procedure: ContentEmulationPlanStepProcedure
    """ Structured procedure-ish payload when upstream provides commands:
    platforms, executors (name/command/cleanup), input_arguments,
    tactic, procedure_group/step labels, cti_source, dependencies.
    Empty object when none. Snapshot onto scenario steps in M3-012 —
    never executed by Blacklight.
     """
    created_at: datetime.datetime
    updated_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.content_emulation_plan_step_procedure import ContentEmulationPlanStepProcedure  # noqa: PLC0415

        id = str(self.id)

        source_id = str(self.source_id)

        version = self.version

        plan_id = str(self.plan_id)

        ordinal = self.ordinal

        external_id = self.external_id

        name = self.name

        description = self.description

        technique_external_id = self.technique_external_id

        procedure = self.procedure.to_dict()

        created_at = self.created_at.isoformat()

        updated_at = self.updated_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "sourceId": source_id,
                "version": version,
                "planId": plan_id,
                "ordinal": ordinal,
                "externalId": external_id,
                "name": name,
                "description": description,
                "techniqueExternalId": technique_external_id,
                "procedure": procedure,
                "createdAt": created_at,
                "updatedAt": updated_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.content_emulation_plan_step_procedure import ContentEmulationPlanStepProcedure  # noqa: PLC0415

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        source_id = UUID(d.pop("sourceId"))

        version = d.pop("version")

        plan_id = UUID(d.pop("planId"))

        ordinal = d.pop("ordinal")

        external_id = d.pop("externalId")

        name = d.pop("name")

        description = d.pop("description")

        technique_external_id = d.pop("techniqueExternalId")

        procedure = ContentEmulationPlanStepProcedure.from_dict(d.pop("procedure"))

        created_at = datetime.datetime.fromisoformat(d.pop("createdAt"))

        updated_at = datetime.datetime.fromisoformat(d.pop("updatedAt"))

        content_emulation_plan_step = cls(
            id=id,
            source_id=source_id,
            version=version,
            plan_id=plan_id,
            ordinal=ordinal,
            external_id=external_id,
            name=name,
            description=description,
            technique_external_id=technique_external_id,
            procedure=procedure,
            created_at=created_at,
            updated_at=updated_at,
        )

        content_emulation_plan_step.additional_properties = d
        return content_emulation_plan_step

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
