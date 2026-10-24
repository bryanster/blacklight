from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.content_procedure_input_arg import ContentProcedureInputArg


T = TypeVar("T", bound="UpdateCustomProcedureTemplateRequest")


@_attrs_define
class UpdateCustomProcedureTemplateRequest:
    """Partial patch for a custom procedure template. Omitted fields stay."""

    name: str | Unset = UNSET
    description: str | Unset = UNSET
    platforms: list[str] | Unset = UNSET
    executor: str | Unset = UNSET
    elevation_required: bool | Unset = UNSET
    command: str | Unset = UNSET
    cleanup: str | Unset = UNSET
    input_args: list[ContentProcedureInputArg] | Unset = UNSET
    technique_external_ids: list[str] | Unset = UNSET
    dependency_executor_name: str | Unset = UNSET
    dependencies: str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.content_procedure_input_arg import ContentProcedureInputArg  # noqa: PLC0415

        name = self.name

        description = self.description

        platforms: list[str] | Unset = UNSET
        if not isinstance(self.platforms, Unset):
            platforms = self.platforms

        executor = self.executor

        elevation_required = self.elevation_required

        command = self.command

        cleanup = self.cleanup

        input_args: list[dict[str, Any]] | Unset = UNSET
        if not isinstance(self.input_args, Unset):
            input_args = []
            for input_args_item_data in self.input_args:
                input_args_item = input_args_item_data.to_dict()
                input_args.append(input_args_item)

        technique_external_ids: list[str] | Unset = UNSET
        if not isinstance(self.technique_external_ids, Unset):
            technique_external_ids = self.technique_external_ids

        dependency_executor_name = self.dependency_executor_name

        dependencies = self.dependencies

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if name is not UNSET:
            field_dict["name"] = name
        if description is not UNSET:
            field_dict["description"] = description
        if platforms is not UNSET:
            field_dict["platforms"] = platforms
        if executor is not UNSET:
            field_dict["executor"] = executor
        if elevation_required is not UNSET:
            field_dict["elevationRequired"] = elevation_required
        if command is not UNSET:
            field_dict["command"] = command
        if cleanup is not UNSET:
            field_dict["cleanup"] = cleanup
        if input_args is not UNSET:
            field_dict["inputArgs"] = input_args
        if technique_external_ids is not UNSET:
            field_dict["techniqueExternalIds"] = technique_external_ids
        if dependency_executor_name is not UNSET:
            field_dict["dependencyExecutorName"] = dependency_executor_name
        if dependencies is not UNSET:
            field_dict["dependencies"] = dependencies

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.content_procedure_input_arg import ContentProcedureInputArg  # noqa: PLC0415

        d = dict(src_dict)
        name = d.pop("name", UNSET)

        description = d.pop("description", UNSET)

        platforms = cast(list[str], d.pop("platforms", UNSET))

        executor = d.pop("executor", UNSET)

        elevation_required = d.pop("elevationRequired", UNSET)

        command = d.pop("command", UNSET)

        cleanup = d.pop("cleanup", UNSET)

        _input_args = d.pop("inputArgs", UNSET)
        input_args: list[ContentProcedureInputArg] | Unset = UNSET
        if _input_args is not UNSET:
            input_args = []
            for input_args_item_data in _input_args:
                input_args_item = ContentProcedureInputArg.from_dict(input_args_item_data)

                input_args.append(input_args_item)

        technique_external_ids = cast(list[str], d.pop("techniqueExternalIds", UNSET))

        dependency_executor_name = d.pop("dependencyExecutorName", UNSET)

        dependencies = d.pop("dependencies", UNSET)

        update_custom_procedure_template_request = cls(
            name=name,
            description=description,
            platforms=platforms,
            executor=executor,
            elevation_required=elevation_required,
            command=command,
            cleanup=cleanup,
            input_args=input_args,
            technique_external_ids=technique_external_ids,
            dependency_executor_name=dependency_executor_name,
            dependencies=dependencies,
        )

        update_custom_procedure_template_request.additional_properties = d
        return update_custom_procedure_template_request

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
