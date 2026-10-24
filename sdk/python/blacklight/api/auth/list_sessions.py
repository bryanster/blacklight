from http import HTTPStatus
from typing import Any, cast
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...types import Response, UNSET
from ... import errors

from ...models.problem import Problem
from ...models.sessions import Sessions
from typing import cast


def _get_kwargs() -> dict[str, Any]:
    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/auth/sessions",
    }

    return _kwargs


def _parse_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> Problem | Sessions | None:
    if response.status_code == 200:
        response_200 = Sessions.from_dict(response.json())

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


def _build_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> Response[Problem | Sessions]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
) -> Response[Problem | Sessions]:
    """List the browsers you are currently signed in on.

     Your own live sessions, newest first, and nobody else's — there is no
    parameter that names another account. An administrator wanting somebody
    else's reaches for `POST /users/{userId}/sessions/revoke`, which ends
    them rather than reading them.

    **Live** means what the authentication middleware means by it: not
    revoked, not past its absolute expiry, and not idle for longer than the
    idle timeout. The same function decides both, so this list cannot
    disagree with what would actually be accepted on the next request — a
    row here is a browser that could act right now, which is what makes
    revoking one worth doing.

    `current` marks the session this request arrived on. It is the one row a
    client should not offer an ordinary "revoke" for: ending it is signing
    out, and `POST /auth/logout` is where that lives, with the cookie
    clearing that goes with it.

    No token and no hash of one. What is in the cookie stays in the cookie.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Problem | Sessions]
    """

    kwargs = _get_kwargs()

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient | Client,
) -> Problem | Sessions | None:
    """List the browsers you are currently signed in on.

     Your own live sessions, newest first, and nobody else's — there is no
    parameter that names another account. An administrator wanting somebody
    else's reaches for `POST /users/{userId}/sessions/revoke`, which ends
    them rather than reading them.

    **Live** means what the authentication middleware means by it: not
    revoked, not past its absolute expiry, and not idle for longer than the
    idle timeout. The same function decides both, so this list cannot
    disagree with what would actually be accepted on the next request — a
    row here is a browser that could act right now, which is what makes
    revoking one worth doing.

    `current` marks the session this request arrived on. It is the one row a
    client should not offer an ordinary "revoke" for: ending it is signing
    out, and `POST /auth/logout` is where that lives, with the cookie
    clearing that goes with it.

    No token and no hash of one. What is in the cookie stays in the cookie.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Problem | Sessions
    """

    return sync_detailed(
        client=client,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
) -> Response[Problem | Sessions]:
    """List the browsers you are currently signed in on.

     Your own live sessions, newest first, and nobody else's — there is no
    parameter that names another account. An administrator wanting somebody
    else's reaches for `POST /users/{userId}/sessions/revoke`, which ends
    them rather than reading them.

    **Live** means what the authentication middleware means by it: not
    revoked, not past its absolute expiry, and not idle for longer than the
    idle timeout. The same function decides both, so this list cannot
    disagree with what would actually be accepted on the next request — a
    row here is a browser that could act right now, which is what makes
    revoking one worth doing.

    `current` marks the session this request arrived on. It is the one row a
    client should not offer an ordinary "revoke" for: ending it is signing
    out, and `POST /auth/logout` is where that lives, with the cookie
    clearing that goes with it.

    No token and no hash of one. What is in the cookie stays in the cookie.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Problem | Sessions]
    """

    kwargs = _get_kwargs()

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
) -> Problem | Sessions | None:
    """List the browsers you are currently signed in on.

     Your own live sessions, newest first, and nobody else's — there is no
    parameter that names another account. An administrator wanting somebody
    else's reaches for `POST /users/{userId}/sessions/revoke`, which ends
    them rather than reading them.

    **Live** means what the authentication middleware means by it: not
    revoked, not past its absolute expiry, and not idle for longer than the
    idle timeout. The same function decides both, so this list cannot
    disagree with what would actually be accepted on the next request — a
    row here is a browser that could act right now, which is what makes
    revoking one worth doing.

    `current` marks the session this request arrived on. It is the one row a
    client should not offer an ordinary "revoke" for: ending it is signing
    out, and `POST /auth/logout` is where that lives, with the cookie
    clearing that goes with it.

    No token and no hash of one. What is in the cookie stays in the cookie.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Problem | Sessions
    """

    return (
        await asyncio_detailed(
            client=client,
        )
    ).parsed
