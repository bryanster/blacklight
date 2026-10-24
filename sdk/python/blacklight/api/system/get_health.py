from http import HTTPStatus
from typing import Any, cast
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...types import Response, UNSET
from ... import errors

from ...models.health import Health
from ...models.problem import Problem
from typing import cast


def _get_kwargs() -> dict[str, Any]:
    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/healthz",
    }

    return _kwargs


def _parse_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> Health | Problem | None:
    if response.status_code == 200:
        response_200 = Health.from_dict(response.json())

        return response_200

    if response.status_code == 500:
        response_500 = Problem.from_dict(response.json())

        return response_500

    if response.status_code == 503:
        response_503 = Health.from_dict(response.json())

        return response_503

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> Response[Health | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
) -> Response[Health | Problem]:
    """Report whether the server and its dependencies are healthy.

     Public by design: an orchestrator or uptime monitor has to be able to call
    this without credentials, and a health check that needs a session reports
    "unhealthy" whenever authentication itself breaks.

    Returns 200 when every check is `ok`, and 503 when any check is `error`.
    The body is the same shape either way — the caller reads `checks` to find
    out *what* is wrong — so the 503 is a health report, not a problem
    document.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Health | Problem]
    """

    kwargs = _get_kwargs()

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient | Client,
) -> Health | Problem | None:
    """Report whether the server and its dependencies are healthy.

     Public by design: an orchestrator or uptime monitor has to be able to call
    this without credentials, and a health check that needs a session reports
    "unhealthy" whenever authentication itself breaks.

    Returns 200 when every check is `ok`, and 503 when any check is `error`.
    The body is the same shape either way — the caller reads `checks` to find
    out *what* is wrong — so the 503 is a health report, not a problem
    document.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Health | Problem
    """

    return sync_detailed(
        client=client,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
) -> Response[Health | Problem]:
    """Report whether the server and its dependencies are healthy.

     Public by design: an orchestrator or uptime monitor has to be able to call
    this without credentials, and a health check that needs a session reports
    "unhealthy" whenever authentication itself breaks.

    Returns 200 when every check is `ok`, and 503 when any check is `error`.
    The body is the same shape either way — the caller reads `checks` to find
    out *what* is wrong — so the 503 is a health report, not a problem
    document.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Health | Problem]
    """

    kwargs = _get_kwargs()

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
) -> Health | Problem | None:
    """Report whether the server and its dependencies are healthy.

     Public by design: an orchestrator or uptime monitor has to be able to call
    this without credentials, and a health check that needs a session reports
    "unhealthy" whenever authentication itself breaks.

    Returns 200 when every check is `ok`, and 503 when any check is `error`.
    The body is the same shape either way — the caller reads `checks` to find
    out *what* is wrong — so the 503 is a health report, not a problem
    document.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Health | Problem
    """

    return (
        await asyncio_detailed(
            client=client,
        )
    ).parsed
