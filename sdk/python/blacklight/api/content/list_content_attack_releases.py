from http import HTTPStatus
from typing import Any, cast
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...types import Response, UNSET
from ... import errors

from ...models.content_attack_release_list import ContentAttackReleaseList
from ...models.problem import Problem
from typing import cast


def _get_kwargs() -> dict[str, Any]:
    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/content/attack/releases",
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> ContentAttackReleaseList | Problem | None:
    if response.status_code == 200:
        response_200 = ContentAttackReleaseList.from_dict(response.json())

        return response_200

    if response.status_code == 401:
        response_401 = Problem.from_dict(response.json())

        return response_401

    if response.status_code == 403:
        response_403 = Problem.from_dict(response.json())

        return response_403

    if response.status_code == 404:
        response_404 = Problem.from_dict(response.json())

        return response_404

    if response.status_code == 500:
        response_500 = Problem.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Response[ContentAttackReleaseList | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
) -> Response[ContentAttackReleaseList | Problem]:
    """List the ATT&CK releases upstream offers, and which are installed.

     Any authenticated subject (`content.read`). Reads the ATT&CK source's
    published release index — the same index a sync with no pin uses to
    resolve "latest" — and merges it with the versions this installation
    already holds. It is what the first-run version picker is built on.

    This reaches upstream while the request is open. It takes no job slot
    and writes nothing; choosing a release and calling
    `POST /content/sources/{sourceId}/sync` is what installs one.

    **An upstream that cannot be reached is a `200`, not a `502`.**
    Air-gapped installations are supported, and for them an unreachable
    index is the normal case rather than a fault. The answer says
    `reachable: false`, carries the transport error in `unreachable`, and
    still lists what is installed — so a client can offer the offline bundle
    path (`docs/content-bundles.md`) instead of a dead end. `latest` is
    absent from every item when upstream did not answer: nothing local knows
    which release is newest, and ATT&CK labels do not sort.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ContentAttackReleaseList | Problem]
    """

    kwargs = _get_kwargs()

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient | Client,
) -> ContentAttackReleaseList | Problem | None:
    """List the ATT&CK releases upstream offers, and which are installed.

     Any authenticated subject (`content.read`). Reads the ATT&CK source's
    published release index — the same index a sync with no pin uses to
    resolve "latest" — and merges it with the versions this installation
    already holds. It is what the first-run version picker is built on.

    This reaches upstream while the request is open. It takes no job slot
    and writes nothing; choosing a release and calling
    `POST /content/sources/{sourceId}/sync` is what installs one.

    **An upstream that cannot be reached is a `200`, not a `502`.**
    Air-gapped installations are supported, and for them an unreachable
    index is the normal case rather than a fault. The answer says
    `reachable: false`, carries the transport error in `unreachable`, and
    still lists what is installed — so a client can offer the offline bundle
    path (`docs/content-bundles.md`) instead of a dead end. `latest` is
    absent from every item when upstream did not answer: nothing local knows
    which release is newest, and ATT&CK labels do not sort.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ContentAttackReleaseList | Problem
    """

    return sync_detailed(
        client=client,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
) -> Response[ContentAttackReleaseList | Problem]:
    """List the ATT&CK releases upstream offers, and which are installed.

     Any authenticated subject (`content.read`). Reads the ATT&CK source's
    published release index — the same index a sync with no pin uses to
    resolve "latest" — and merges it with the versions this installation
    already holds. It is what the first-run version picker is built on.

    This reaches upstream while the request is open. It takes no job slot
    and writes nothing; choosing a release and calling
    `POST /content/sources/{sourceId}/sync` is what installs one.

    **An upstream that cannot be reached is a `200`, not a `502`.**
    Air-gapped installations are supported, and for them an unreachable
    index is the normal case rather than a fault. The answer says
    `reachable: false`, carries the transport error in `unreachable`, and
    still lists what is installed — so a client can offer the offline bundle
    path (`docs/content-bundles.md`) instead of a dead end. `latest` is
    absent from every item when upstream did not answer: nothing local knows
    which release is newest, and ATT&CK labels do not sort.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ContentAttackReleaseList | Problem]
    """

    kwargs = _get_kwargs()

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
) -> ContentAttackReleaseList | Problem | None:
    """List the ATT&CK releases upstream offers, and which are installed.

     Any authenticated subject (`content.read`). Reads the ATT&CK source's
    published release index — the same index a sync with no pin uses to
    resolve "latest" — and merges it with the versions this installation
    already holds. It is what the first-run version picker is built on.

    This reaches upstream while the request is open. It takes no job slot
    and writes nothing; choosing a release and calling
    `POST /content/sources/{sourceId}/sync` is what installs one.

    **An upstream that cannot be reached is a `200`, not a `502`.**
    Air-gapped installations are supported, and for them an unreachable
    index is the normal case rather than a fault. The answer says
    `reachable: false`, carries the transport error in `unreachable`, and
    still lists what is installed — so a client can offer the offline bundle
    path (`docs/content-bundles.md`) instead of a dead end. `latest` is
    absent from every item when upstream did not answer: nothing local knows
    which release is newest, and ATT&CK labels do not sort.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ContentAttackReleaseList | Problem
    """

    return (
        await asyncio_detailed(
            client=client,
        )
    ).parsed
