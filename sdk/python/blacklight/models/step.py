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
    from ..models.step_procedure import StepProcedure


T = TypeVar("T", bound="Step")


@_attrs_define
class Step:
    id: UUID
    """ UUIDv7. """
    scenario_id: UUID
    ordinal: int
    """ 1-based dense position; UI order. """
    name: str
    objective: str
    template_id: str
    """ Weak lineage to content procedure template. """
    target_asset: str
    attack_version: str
    """ Snapshot of the engagement's ATT&CK version at create time. """
    created_at: datetime.datetime
    updated_at: datetime.datetime
    technique_id: str | Unset = UNSET
    """ ATT&CK technique external id (e.g. "T1059"). """
    subtechnique_id: str | Unset = UNSET
    """ ATT&CK subtechnique external id (e.g. "T1059.001"). """
    tactic_id: str | Unset = UNSET
    """ ATT&CK tactic external id (e.g. "TA0002"). """
    procedure: StepProcedure | Unset = UNSET
    """ Structured procedure payload (platform, executor, command, etc). """
    tools: list[str] | Unset = UNSET
    """ Tools used in this step. """
    controls_in_scope: list[str] | Unset = UNSET
    """ Controls in scope for this step. """
    revealed_at: datetime.datetime | Unset = UNSET
    """ When this step was revealed to blue (null if unrevealed). """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.step_procedure import StepProcedure  # noqa: PLC0415

        id = str(self.id)

        scenario_id = str(self.scenario_id)

        ordinal = self.ordinal

        name = self.name

        objective = self.objective

        template_id = self.template_id

        target_asset = self.target_asset

        attack_version = self.attack_version

        created_at = self.created_at.isoformat()

        updated_at = self.updated_at.isoformat()

        technique_id = self.technique_id

        subtechnique_id = self.subtechnique_id

        tactic_id = self.tactic_id

        procedure: dict[str, Any] | Unset = UNSET
        if not isinstance(self.procedure, Unset):
            procedure = self.procedure.to_dict()

        tools: list[str] | Unset = UNSET
        if not isinstance(self.tools, Unset):
            tools = self.tools

        controls_in_scope: list[str] | Unset = UNSET
        if not isinstance(self.controls_in_scope, Unset):
            controls_in_scope = self.controls_in_scope

        revealed_at: str | Unset = UNSET
        if not isinstance(self.revealed_at, Unset):
            revealed_at = self.revealed_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "scenarioId": scenario_id,
                "ordinal": ordinal,
                "name": name,
                "objective": objective,
                "templateId": template_id,
                "targetAsset": target_asset,
                "attackVersion": attack_version,
                "createdAt": created_at,
                "updatedAt": updated_at,
            }
        )
        if technique_id is not UNSET:
            field_dict["techniqueId"] = technique_id
        if subtechnique_id is not UNSET:
            field_dict["subtechniqueId"] = subtechnique_id
        if tactic_id is not UNSET:
            field_dict["tacticId"] = tactic_id
        if procedure is not UNSET:
            field_dict["procedure"] = procedure
        if tools is not UNSET:
            field_dict["tools"] = tools
        if controls_in_scope is not UNSET:
            field_dict["controlsInScope"] = controls_in_scope
        if revealed_at is not UNSET:
            field_dict["revealedAt"] = revealed_at

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.step_procedure import StepProcedure  # noqa: PLC0415

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        scenario_id = UUID(d.pop("scenarioId"))

        ordinal = d.pop("ordinal")

        name = d.pop("name")

        objective = d.pop("objective")

        template_id = d.pop("templateId")

        target_asset = d.pop("targetAsset")

        attack_version = d.pop("attackVersion")

        created_at = datetime.datetime.fromisoformat(d.pop("createdAt"))

        updated_at = datetime.datetime.fromisoformat(d.pop("updatedAt"))

        technique_id = d.pop("techniqueId", UNSET)

        subtechnique_id = d.pop("subtechniqueId", UNSET)

        tactic_id = d.pop("tacticId", UNSET)

        _procedure = d.pop("procedure", UNSET)
        procedure: StepProcedure | Unset
        if isinstance(_procedure, Unset):
            procedure = UNSET
        else:
            procedure = StepProcedure.from_dict(_procedure)

        tools = cast(list[str], d.pop("tools", UNSET))

        controls_in_scope = cast(list[str], d.pop("controlsInScope", UNSET))

        _revealed_at = d.pop("revealedAt", UNSET)
        revealed_at: datetime.datetime | Unset
        if isinstance(_revealed_at, Unset):
            revealed_at = UNSET
        else:
            revealed_at = datetime.datetime.fromisoformat(_revealed_at)

        step = cls(
            id=id,
            scenario_id=scenario_id,
            ordinal=ordinal,
            name=name,
            objective=objective,
            template_id=template_id,
            target_asset=target_asset,
            attack_version=attack_version,
            created_at=created_at,
            updated_at=updated_at,
            technique_id=technique_id,
            subtechnique_id=subtechnique_id,
            tactic_id=tactic_id,
            procedure=procedure,
            tools=tools,
            controls_in_scope=controls_in_scope,
            revealed_at=revealed_at,
        )

        step.additional_properties = d
        return step

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
