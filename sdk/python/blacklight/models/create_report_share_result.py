from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, BinaryIO, TextIO, TYPE_CHECKING, Generator

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

from ..types import UNSET, Unset
from typing import cast

if TYPE_CHECKING:
    from ..models.report_share import ReportShare


T = TypeVar("T", bound="CreateReportShareResult")


@_attrs_define
class CreateReportShareResult:
    share: ReportShare
    claim_url: str
    """ Absolute claim URL for sharing with recipients. """
    token: str | Unset = UNSET
    """ The share token. Shown once; never stored plaintext. """

    def to_dict(self) -> dict[str, Any]:
        from ..models.report_share import ReportShare  # noqa: PLC0415

        share = self.share.to_dict()

        claim_url = self.claim_url

        token = self.token

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "share": share,
                "claimUrl": claim_url,
            }
        )
        if token is not UNSET:
            field_dict["token"] = token

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.report_share import ReportShare  # noqa: PLC0415

        d = dict(src_dict)
        share = ReportShare.from_dict(d.pop("share"))

        claim_url = d.pop("claimUrl")

        token = d.pop("token", UNSET)

        create_report_share_result = cls(
            share=share,
            claim_url=claim_url,
            token=token,
        )

        return create_report_share_result
