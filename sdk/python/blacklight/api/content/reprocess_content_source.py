from http import HTTPStatus
from typing import Any, cast
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...types import Response, UNSET
from ... import errors

from ...models.content_sync_job import ContentSyncJob
from ...models.problem import Problem
from ...models.reprocess_content_source_request import ReprocessContentSourceRequest
from ...types import UNSET, Unset
from typing import cast
from uuid import UUID


def _get_kwargs(
    source_id: UUID,
    *,
    body: ReprocessContentSourceRequest | Unset = UNSET,
    x_csrf_token: str | Unset = UNSET,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}
    if not isinstance(x_csrf_token, Unset):
        headers["X-CSRF-Token"] = x_csrf_token

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/content/sources/{source_id}/reprocess".format(
            source_id=quote(str(source_id), safe=""),
        ),
    }

    if not isinstance(body, Unset):
        _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> ContentSyncJob | Problem | None:
    if response.status_code == 202:
        response_202 = ContentSyncJob.from_dict(response.json())

        return response_202

    if response.status_code == 400:
        response_400 = Problem.from_dict(response.json())

        return response_400

    if response.status_code == 401:
        response_401 = Problem.from_dict(response.json())

        return response_401

    if response.status_code == 403:
        response_403 = Problem.from_dict(response.json())

        return response_403

    if response.status_code == 404:
        response_404 = Problem.from_dict(response.json())

        return response_404

    if response.status_code == 409:
        response_409 = Problem.from_dict(response.json())

        return response_409

    if response.status_code == 500:
        response_500 = Problem.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Response[ContentSyncJob | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    source_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: ReprocessContentSourceRequest | Unset = UNSET,
    x_csrf_token: str | Unset = UNSET,
) -> Response[ContentSyncJob | Problem]:
    """Reprocess a source from its last raw snapshot.

     Administrators only (`content.sync`). Creates a `reprocess` job that
    opens the last successful raw snapshot for the named version and runs
    Parse → Normalize → Apply with no network Fetch. Use this after an
    adapter bugfix to repair the catalog without another download.

    The optional body pin `{ "version": "15.1" }` selects which snapshot.
    Rolling sources (Atomic, Sigma, CTID) default to `current` when the
    pin is omitted. ATT&CK requires an explicit version. Answers `409`
    when no raw snapshot exists for that source/version, and when another
    content job already holds the global slot.

    Args:
        source_id (UUID):
        x_csrf_token (str | Unset):
        body (ReprocessContentSourceRequest | Unset): Optional pin for which raw snapshot to
            reprocess. Additional properties
            are rejected so a mistyped field cannot silently no-op.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ContentSyncJob | Problem]
    """

    kwargs = _get_kwargs(
        source_id=source_id,
        body=body,
        x_csrf_token=x_csrf_token,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    source_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: ReprocessContentSourceRequest | Unset = UNSET,
    x_csrf_token: str | Unset = UNSET,
) -> ContentSyncJob | Problem | None:
    """Reprocess a source from its last raw snapshot.

     Administrators only (`content.sync`). Creates a `reprocess` job that
    opens the last successful raw snapshot for the named version and runs
    Parse → Normalize → Apply with no network Fetch. Use this after an
    adapter bugfix to repair the catalog without another download.

    The optional body pin `{ "version": "15.1" }` selects which snapshot.
    Rolling sources (Atomic, Sigma, CTID) default to `current` when the
    pin is omitted. ATT&CK requires an explicit version. Answers `409`
    when no raw snapshot exists for that source/version, and when another
    content job already holds the global slot.

    Args:
        source_id (UUID):
        x_csrf_token (str | Unset):
        body (ReprocessContentSourceRequest | Unset): Optional pin for which raw snapshot to
            reprocess. Additional properties
            are rejected so a mistyped field cannot silently no-op.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ContentSyncJob | Problem
    """

    return sync_detailed(
        source_id=source_id,
        client=client,
        body=body,
        x_csrf_token=x_csrf_token,
    ).parsed


async def asyncio_detailed(
    source_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: ReprocessContentSourceRequest | Unset = UNSET,
    x_csrf_token: str | Unset = UNSET,
) -> Response[ContentSyncJob | Problem]:
    """Reprocess a source from its last raw snapshot.

     Administrators only (`content.sync`). Creates a `reprocess` job that
    opens the last successful raw snapshot for the named version and runs
    Parse → Normalize → Apply with no network Fetch. Use this after an
    adapter bugfix to repair the catalog without another download.

    The optional body pin `{ "version": "15.1" }` selects which snapshot.
    Rolling sources (Atomic, Sigma, CTID) default to `current` when the
    pin is omitted. ATT&CK requires an explicit version. Answers `409`
    when no raw snapshot exists for that source/version, and when another
    content job already holds the global slot.

    Args:
        source_id (UUID):
        x_csrf_token (str | Unset):
        body (ReprocessContentSourceRequest | Unset): Optional pin for which raw snapshot to
            reprocess. Additional properties
            are rejected so a mistyped field cannot silently no-op.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ContentSyncJob | Problem]
    """

    kwargs = _get_kwargs(
        source_id=source_id,
        body=body,
        x_csrf_token=x_csrf_token,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    source_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: ReprocessContentSourceRequest | Unset = UNSET,
    x_csrf_token: str | Unset = UNSET,
) -> ContentSyncJob | Problem | None:
    """Reprocess a source from its last raw snapshot.

     Administrators only (`content.sync`). Creates a `reprocess` job that
    opens the last successful raw snapshot for the named version and runs
    Parse → Normalize → Apply with no network Fetch. Use this after an
    adapter bugfix to repair the catalog without another download.

    The optional body pin `{ "version": "15.1" }` selects which snapshot.
    Rolling sources (Atomic, Sigma, CTID) default to `current` when the
    pin is omitted. ATT&CK requires an explicit version. Answers `409`
    when no raw snapshot exists for that source/version, and when another
    content job already holds the global slot.

    Args:
        source_id (UUID):
        x_csrf_token (str | Unset):
        body (ReprocessContentSourceRequest | Unset): Optional pin for which raw snapshot to
            reprocess. Additional properties
            are rejected so a mistyped field cannot silently no-op.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ContentSyncJob | Problem
    """

    return (
        await asyncio_detailed(
            source_id=source_id,
            client=client,
            body=body,
            x_csrf_token=x_csrf_token,
        )
    ).parsed
