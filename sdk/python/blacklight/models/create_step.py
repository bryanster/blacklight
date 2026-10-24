from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.create_step_procedure import CreateStepProcedure


T = TypeVar("T", bound="CreateStep")


@_attrs_define
class CreateStep:
    name: str
    objective: str | Unset = ""
    technique_id: str | Unset = ""
    subtechnique_id: str | Unset = ""
    tactic_id: str | Unset = ""
    technique_external_id: str | Unset = UNSET
    """ When set, resolve the technique against the engagement's pinned ATT&CK version and snapshot display fields.
    Mutually exclusive with technique_id/subtechnique_id. """
    procedure: CreateStepProcedure | Unset = UNSET
    """ Structured procedure JSON. """
    template_id: str | Unset = ""
    target_asset: str | Unset = ""
    tools: list[str] | Unset = UNSET
    controls_in_scope: list[str] | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.create_step_procedure import CreateStepProcedure  # noqa: PLC0415

        name = self.name

        objective = self.objective

        technique_id = self.technique_id

        subtechnique_id = self.subtechnique_id

        tactic_id = self.tactic_id

        technique_external_id = self.technique_external_id

        procedure: dict[str, Any] | Unset = UNSET
        if not isinstance(self.procedure, Unset):
            procedure = self.procedure.to_dict()

        template_id = self.template_id

        target_asset = self.target_asset

        tools: list[str] | Unset = UNSET
        if not isinstance(self.tools, Unset):
            tools = self.tools

        controls_in_scope: list[str] | Unset = UNSET
        if not isinstance(self.controls_in_scope, Unset):
            controls_in_scope = self.controls_in_scope

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "name": name,
            }
        )
        if objective is not UNSET:
            field_dict["objective"] = objective
        if technique_id is not UNSET:
            field_dict["techniqueId"] = technique_id
        if subtechnique_id is not UNSET:
            field_dict["subtechniqueId"] = subtechnique_id
        if tactic_id is not UNSET:
            field_dict["tacticId"] = tactic_id
        if technique_external_id is not UNSET:
            field_dict["techniqueExternalId"] = technique_external_id
        if procedure is not UNSET:
            field_dict["procedure"] = procedure
        if template_id is not UNSET:
            field_dict["templateId"] = template_id
        if target_asset is not UNSET:
            field_dict["targetAsset"] = target_asset
        if tools is not UNSET:
            field_dict["tools"] = tools
        if controls_in_scope is not UNSET:
            field_dict["controlsInScope"] = controls_in_scope

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.create_step_procedure import CreateStepProcedure  # noqa: PLC0415

        d = dict(src_dict)
        name = d.pop("name")

        objective = d.pop("objective", UNSET)

        technique_id = d.pop("techniqueId", UNSET)

        subtechnique_id = d.pop("subtechniqueId", UNSET)

        tactic_id = d.pop("tacticId", UNSET)

        technique_external_id = d.pop("techniqueExternalId", UNSET)

        _procedure = d.pop("procedure", UNSET)
        procedure: CreateStepProcedure | Unset
        if isinstance(_procedure, Unset):
            procedure = UNSET
        else:
            procedure = CreateStepProcedure.from_dict(_procedure)

        template_id = d.pop("templateId", UNSET)

        target_asset = d.pop("targetAsset", UNSET)

        tools = cast(list[str], d.pop("tools", UNSET))

        controls_in_scope = cast(list[str], d.pop("controlsInScope", UNSET))

        create_step = cls(
            name=name,
            objective=objective,
            technique_id=technique_id,
            subtechnique_id=subtechnique_id,
            tactic_id=tactic_id,
            technique_external_id=technique_external_id,
            procedure=procedure,
            template_id=template_id,
            target_asset=target_asset,
            tools=tools,
            controls_in_scope=controls_in_scope,
        )

        create_step.additional_properties = d
        return create_step

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
