from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..models.problem_code import ProblemCode
from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.field_error import FieldError


T = TypeVar("T", bound="Problem")


@_attrs_define
class Problem:
    """RFC 9457 problem detail — the only error shape this API produces, served
    as `application/problem+json` (M0B-007).

    Clients switch on `code`. `title` and `detail` are prose for a human and
    may be reworded at any time.

    """

    type_: str
    """ URI reference identifying the problem type. `about:blank` when the
    status code says everything there is to say.
     """
    title: str
    """ Short human-readable summary of the problem type. Stable for a given `type`. """
    status: int
    """ The HTTP status code, repeated so a logged or forwarded body stands alone. """
    code: ProblemCode
    """ Stable machine-readable error identifier. Clients switch on this, never on
    `detail` and never on the status alone.

    Each code maps to exactly one HTTP status; the mapping lives in
    internal/httpapi/apierr and a test there asserts it (M0B-007). Adding a
    code means adding it in both places in the same change.

    The reverse is *nearly* 1:1. A code may refine another — same status,
    more specific instruction — and `mfa_enrolment_required` is the one that
    does, refining `forbidden`. A client that does not know a refinement can
    always fall back to the code it refines, and internal/httpapi/apierr
    holds the refinement table that says which that is.
     """
    detail: str | Unset = UNSET
    """ Explanation specific to this occurrence. Never carries internal
    detail: an unrecognised server-side error is reported generically and
    the real error goes to the log.
     """
    instance: str | Unset = UNSET
    """ The request ID, echoed as `X-Request-Id`. A user quotes it in a bug
    report and an operator greps for it.

    RFC 9457 types this as a URI reference; a bare request ID is a
    relative reference, and it is far more useful to a support desk than
    a URI that resolves to nothing.
     """
    errors: list[FieldError] | Unset = UNSET
    """ Field-level failures. Present when `code` is `validation_failed`. """
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.field_error import FieldError  # noqa: PLC0415

        type_ = self.type_

        title = self.title

        status = self.status

        code = self.code.value

        detail = self.detail

        instance = self.instance

        errors: list[dict[str, Any]] | Unset = UNSET
        if not isinstance(self.errors, Unset):
            errors = []
            for errors_item_data in self.errors:
                errors_item = errors_item_data.to_dict()
                errors.append(errors_item)

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "type": type_,
                "title": title,
                "status": status,
                "code": code,
            }
        )
        if detail is not UNSET:
            field_dict["detail"] = detail
        if instance is not UNSET:
            field_dict["instance"] = instance
        if errors is not UNSET:
            field_dict["errors"] = errors

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.field_error import FieldError  # noqa: PLC0415

        d = dict(src_dict)
        type_ = d.pop("type")

        title = d.pop("title")

        status = d.pop("status")

        code = ProblemCode(d.pop("code"))

        detail = d.pop("detail", UNSET)

        instance = d.pop("instance", UNSET)

        _errors = d.pop("errors", UNSET)
        errors: list[FieldError] | Unset = UNSET
        if _errors is not UNSET:
            errors = []
            for errors_item_data in _errors:
                errors_item = FieldError.from_dict(errors_item_data)

                errors.append(errors_item)

        problem = cls(
            type_=type_,
            title=title,
            status=status,
            code=code,
            detail=detail,
            instance=instance,
            errors=errors,
        )

        problem.additional_properties = d
        return problem

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
