from http import HTTPStatus
from typing import Any, cast
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...types import Response, UNSET
from ... import errors

from ...models.content_attack_version_list import ContentAttackVersionList
from ...models.problem import Problem
from typing import cast


def _get_kwargs() -> dict[str, Any]:
    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/content/attack/versions",
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> ContentAttackVersionList | Problem | None:
    if response.status_code == 200:
        response_200 = ContentAttackVersionList.from_dict(response.json())

        return response_200

    if response.status_code == 401:
        response_401 = Problem.from_dict(response.json())

        return response_401

    if response.status_code == 403:
        response_403 = Problem.from_dict(response.json())

        return response_403

    if response.status_code == 500:
        response_500 = Problem.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Response[ContentAttackVersionList | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
) -> Response[ContentAttackVersionList | Problem]:
    """List installed ATT&CK versions.

     Any authenticated subject (`content.read`). Returns every version
    snapshot under the ATT&CK source with item counts and `syncedAt`.

    The version string is an opaque release label equal to
    `content_source_version.version` (for example `15.1`). There is no
    semver rewriting and no implicit "latest" — pin-sensitive callers must
    name the version they mean. See `docs/content-attack.md` and
    `docs/content-copy-on-use.md`.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ContentAttackVersionList | Problem]
    """

    kwargs = _get_kwargs()

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient | Client,
) -> ContentAttackVersionList | Problem | None:
    """List installed ATT&CK versions.

     Any authenticated subject (`content.read`). Returns every version
    snapshot under the ATT&CK source with item counts and `syncedAt`.

    The version string is an opaque release label equal to
    `content_source_version.version` (for example `15.1`). There is no
    semver rewriting and no implicit "latest" — pin-sensitive callers must
    name the version they mean. See `docs/content-attack.md` and
    `docs/content-copy-on-use.md`.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ContentAttackVersionList | Problem
    """

    return sync_detailed(
        client=client,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
) -> Response[ContentAttackVersionList | Problem]:
    """List installed ATT&CK versions.

     Any authenticated subject (`content.read`). Returns every version
    snapshot under the ATT&CK source with item counts and `syncedAt`.

    The version string is an opaque release label equal to
    `content_source_version.version` (for example `15.1`). There is no
    semver rewriting and no implicit "latest" — pin-sensitive callers must
    name the version they mean. See `docs/content-attack.md` and
    `docs/content-copy-on-use.md`.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ContentAttackVersionList | Problem]
    """

    kwargs = _get_kwargs()

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
) -> ContentAttackVersionList | Problem | None:
    """List installed ATT&CK versions.

     Any authenticated subject (`content.read`). Returns every version
    snapshot under the ATT&CK source with item counts and `syncedAt`.

    The version string is an opaque release label equal to
    `content_source_version.version` (for example `15.1`). There is no
    semver rewriting and no implicit "latest" — pin-sensitive callers must
    name the version they mean. See `docs/content-attack.md` and
    `docs/content-copy-on-use.md`.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ContentAttackVersionList | Problem
    """

    return (
        await asyncio_detailed(
            client=client,
        )
    ).parsed
