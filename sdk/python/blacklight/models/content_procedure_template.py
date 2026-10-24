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
    from ..models.content_procedure_input_arg import ContentProcedureInputArg


T = TypeVar("T", bound="ContentProcedureTemplate")


@_attrs_define
class ContentProcedureTemplate:
    """One structured procedure template (Atomic test or custom). Platforms,
    executor, command, cleanup, and input args stay distinct — never a
    single flattened actions string (PLAN.md §3).

    """

    id: UUID
    source_id: UUID
    version: str
    """ Version token. Atomic Red Team is rolling-head and always `current`.
     """
    external_id: str
    """ Stable id within the source. Prefer upstream `auto_generated_guid`;
    otherwise `{technique}/{zero-based-index}` (documented in
    `docs/content-atomic.md`).
     """
    name: str
    description: str
    platforms: list[str]
    """ Supported platforms (`windows`, `linux`, `macos`, …). """
    executor: str
    """ Executor name (`powershell`, `command_prompt`, `sh`, `bash`, `manual`, …). """
    elevation_required: bool
    command: str
    """ Primary command or manual steps body. """
    cleanup: str
    """ Cleanup command when present; empty string otherwise. """
    input_args: list[ContentProcedureInputArg]
    technique_external_ids: list[str]
    """ ATT&CK technique external ids this template maps to. """
    dependency_executor_name: str
    dependencies: str
    """ Serialized dependency list when present (JSON array text). Empty
    string when the upstream test has no dependencies.
     """
    created_at: datetime.datetime
    updated_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.content_procedure_input_arg import ContentProcedureInputArg  # noqa: PLC0415

        id = str(self.id)

        source_id = str(self.source_id)

        version = self.version

        external_id = self.external_id

        name = self.name

        description = self.description

        platforms = self.platforms

        executor = self.executor

        elevation_required = self.elevation_required

        command = self.command

        cleanup = self.cleanup

        input_args = []
        for input_args_item_data in self.input_args:
            input_args_item = input_args_item_data.to_dict()
            input_args.append(input_args_item)

        technique_external_ids = self.technique_external_ids

        dependency_executor_name = self.dependency_executor_name

        dependencies = self.dependencies

        created_at = self.created_at.isoformat()

        updated_at = self.updated_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "sourceId": source_id,
                "version": version,
                "externalId": external_id,
                "name": name,
                "description": description,
                "platforms": platforms,
                "executor": executor,
                "elevationRequired": elevation_required,
                "command": command,
                "cleanup": cleanup,
                "inputArgs": input_args,
                "techniqueExternalIds": technique_external_ids,
                "dependencyExecutorName": dependency_executor_name,
                "dependencies": dependencies,
                "createdAt": created_at,
                "updatedAt": updated_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.content_procedure_input_arg import ContentProcedureInputArg  # noqa: PLC0415

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        source_id = UUID(d.pop("sourceId"))

        version = d.pop("version")

        external_id = d.pop("externalId")

        name = d.pop("name")

        description = d.pop("description")

        platforms = cast(list[str], d.pop("platforms"))

        executor = d.pop("executor")

        elevation_required = d.pop("elevationRequired")

        command = d.pop("command")

        cleanup = d.pop("cleanup")

        input_args = []
        _input_args = d.pop("inputArgs")
        for input_args_item_data in _input_args:
            input_args_item = ContentProcedureInputArg.from_dict(input_args_item_data)

            input_args.append(input_args_item)

        technique_external_ids = cast(list[str], d.pop("techniqueExternalIds"))

        dependency_executor_name = d.pop("dependencyExecutorName")

        dependencies = d.pop("dependencies")

        created_at = datetime.datetime.fromisoformat(d.pop("createdAt"))

        updated_at = datetime.datetime.fromisoformat(d.pop("updatedAt"))

        content_procedure_template = cls(
            id=id,
            source_id=source_id,
            version=version,
            external_id=external_id,
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
            created_at=created_at,
            updated_at=updated_at,
        )

        content_procedure_template.additional_properties = d
        return content_procedure_template

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
