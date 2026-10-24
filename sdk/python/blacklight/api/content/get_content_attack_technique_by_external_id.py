from http import HTTPStatus
from typing import Any, cast
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...types import Response, UNSET
from ... import errors

from ...models.content_technique import ContentTechnique
from ...models.problem import Problem
from typing import cast


def _get_kwargs(
    version: str,
    external_id: str,
) -> dict[str, Any]:
    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/content/attack/versions/{version}/techniques/{external_id}".format(
            version=quote(str(version), safe=""),
            external_id=quote(str(external_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> ContentTechnique | Problem | None:
    if response.status_code == 200:
        response_200 = ContentTechnique.from_dict(response.json())

        return response_200

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

    if response.status_code == 500:
        response_500 = Problem.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Response[ContentTechnique | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    version: str,
    external_id: str,
    *,
    client: AuthenticatedClient | Client,
) -> Response[ContentTechnique | Problem]:
    """Resolve a technique by ATT&CK version and MITRE id.

     Any authenticated subject. Looks up the technique natural key inside
    **exactly** the named version. Never falls back to another installed
    version — a technique that exists only in `15.1` is `404` under `14.1`.

    The version path param is required; there is no default "latest" on
    this pin-sensitive endpoint.

    Args:
        version (str):
        external_id (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ContentTechnique | Problem]
    """

    kwargs = _get_kwargs(
        version=version,
        external_id=external_id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    version: str,
    external_id: str,
    *,
    client: AuthenticatedClient | Client,
) -> ContentTechnique | Problem | None:
    """Resolve a technique by ATT&CK version and MITRE id.

     Any authenticated subject. Looks up the technique natural key inside
    **exactly** the named version. Never falls back to another installed
    version — a technique that exists only in `15.1` is `404` under `14.1`.

    The version path param is required; there is no default "latest" on
    this pin-sensitive endpoint.

    Args:
        version (str):
        external_id (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ContentTechnique | Problem
    """

    return sync_detailed(
        version=version,
        external_id=external_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    version: str,
    external_id: str,
    *,
    client: AuthenticatedClient | Client,
) -> Response[ContentTechnique | Problem]:
    """Resolve a technique by ATT&CK version and MITRE id.

     Any authenticated subject. Looks up the technique natural key inside
    **exactly** the named version. Never falls back to another installed
    version — a technique that exists only in `15.1` is `404` under `14.1`.

    The version path param is required; there is no default "latest" on
    this pin-sensitive endpoint.

    Args:
        version (str):
        external_id (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ContentTechnique | Problem]
    """

    kwargs = _get_kwargs(
        version=version,
        external_id=external_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    version: str,
    external_id: str,
    *,
    client: AuthenticatedClient | Client,
) -> ContentTechnique | Problem | None:
    """Resolve a technique by ATT&CK version and MITRE id.

     Any authenticated subject. Looks up the technique natural key inside
    **exactly** the named version. Never falls back to another installed
    version — a technique that exists only in `15.1` is `404` under `14.1`.

    The version path param is required; there is no default "latest" on
    this pin-sensitive endpoint.

    Args:
        version (str):
        external_id (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ContentTechnique | Problem
    """

    return (
        await asyncio_detailed(
            version=version,
            external_id=external_id,
            client=client,
        )
    ).parsed
