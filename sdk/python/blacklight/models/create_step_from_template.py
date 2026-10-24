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
    from ..models.create_step_from_template_arg_values import CreateStepFromTemplateArgValues


T = TypeVar("T", bound="CreateStepFromTemplate")


@_attrs_define
class CreateStepFromTemplate:
    template_id: UUID
    """ Content catalog surrogate id of the procedure template. """
    name: str | Unset = UNSET
    """ Override the step name (defaults to the template name). """
    objective: str | Unset = UNSET
    """ Override the step objective (defaults to template description). """
    target_asset: str | Unset = UNSET
    """ Target asset for this step. """
    arg_values: CreateStepFromTemplateArgValues | Unset = UNSET
    """ Map of template input arg name to value. `#{key}` placeholders in
    command and cleanup are replaced with the corresponding value.
    Keys without a provided value are left as-is in the snapshot.
     """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.create_step_from_template_arg_values import CreateStepFromTemplateArgValues  # noqa: PLC0415

        template_id = str(self.template_id)

        name = self.name

        objective = self.objective

        target_asset = self.target_asset

        arg_values: dict[str, Any] | Unset = UNSET
        if not isinstance(self.arg_values, Unset):
            arg_values = self.arg_values.to_dict()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "templateId": template_id,
            }
        )
        if name is not UNSET:
            field_dict["name"] = name
        if objective is not UNSET:
            field_dict["objective"] = objective
        if target_asset is not UNSET:
            field_dict["targetAsset"] = target_asset
        if arg_values is not UNSET:
            field_dict["argValues"] = arg_values

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.create_step_from_template_arg_values import CreateStepFromTemplateArgValues  # noqa: PLC0415

        d = dict(src_dict)
        template_id = UUID(d.pop("templateId"))

        name = d.pop("name", UNSET)

        objective = d.pop("objective", UNSET)

        target_asset = d.pop("targetAsset", UNSET)

        _arg_values = d.pop("argValues", UNSET)
        arg_values: CreateStepFromTemplateArgValues | Unset
        if isinstance(_arg_values, Unset):
            arg_values = UNSET
        else:
            arg_values = CreateStepFromTemplateArgValues.from_dict(_arg_values)

        create_step_from_template = cls(
            template_id=template_id,
            name=name,
            objective=objective,
            target_asset=target_asset,
            arg_values=arg_values,
        )

        create_step_from_template.additional_properties = d
        return create_step_from_template

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
