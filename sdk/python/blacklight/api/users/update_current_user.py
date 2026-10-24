from http import HTTPStatus
from typing import Any, cast
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...types import Response, UNSET
from ... import errors

from ...models.problem import Problem
from ...models.update_self_request import UpdateSelfRequest
from ...models.user import User
from ...types import UNSET, Unset
from typing import cast


def _get_kwargs(
    *,
    body: UpdateSelfRequest,
    x_csrf_token: str | Unset = UNSET,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}
    if not isinstance(x_csrf_token, Unset):
        headers["X-CSRF-Token"] = x_csrf_token

    _kwargs: dict[str, Any] = {
        "method": "patch",
        "url": "/users/me",
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> Problem | User | None:
    if response.status_code == 200:
        response_200 = User.from_dict(response.json())

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

    if response.status_code == 500:
        response_500 = Problem.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> Response[Problem | User]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
    body: UpdateSelfRequest,
    x_csrf_token: str | Unset = UNSET,
) -> Response[Problem | User]:
    """Change your own display name.

     The self-service half of user administration, and deliberately the
    smallest one that could exist: a display name and nothing else.

    Your platform role and your status are **not** fields of this request.
    Not filtered out of it — absent from it, so a body that names one is
    rejected by the request validator with `400` before any handler runs
    (PLAN.md §4: "field safety comes from the schema"). Raising your own
    privilege through this endpoint is not something the server declines to
    do; it is something there is no way to ask for.

    Args:
        x_csrf_token (str | Unset):
        body (UpdateSelfRequest): Body of `PATCH /users/me`. One field, and that is the design: a
            schema
            with no `platformRole` in it is a request that cannot ask for one, which
            is a stronger guarantee than a handler that declines to honour it
            (PLAN.md §4).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Problem | User]
    """

    kwargs = _get_kwargs(
        body=body,
        x_csrf_token=x_csrf_token,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient | Client,
    body: UpdateSelfRequest,
    x_csrf_token: str | Unset = UNSET,
) -> Problem | User | None:
    """Change your own display name.

     The self-service half of user administration, and deliberately the
    smallest one that could exist: a display name and nothing else.

    Your platform role and your status are **not** fields of this request.
    Not filtered out of it — absent from it, so a body that names one is
    rejected by the request validator with `400` before any handler runs
    (PLAN.md §4: "field safety comes from the schema"). Raising your own
    privilege through this endpoint is not something the server declines to
    do; it is something there is no way to ask for.

    Args:
        x_csrf_token (str | Unset):
        body (UpdateSelfRequest): Body of `PATCH /users/me`. One field, and that is the design: a
            schema
            with no `platformRole` in it is a request that cannot ask for one, which
            is a stronger guarantee than a handler that declines to honour it
            (PLAN.md §4).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Problem | User
    """

    return sync_detailed(
        client=client,
        body=body,
        x_csrf_token=x_csrf_token,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
    body: UpdateSelfRequest,
    x_csrf_token: str | Unset = UNSET,
) -> Response[Problem | User]:
    """Change your own display name.

     The self-service half of user administration, and deliberately the
    smallest one that could exist: a display name and nothing else.

    Your platform role and your status are **not** fields of this request.
    Not filtered out of it — absent from it, so a body that names one is
    rejected by the request validator with `400` before any handler runs
    (PLAN.md §4: "field safety comes from the schema"). Raising your own
    privilege through this endpoint is not something the server declines to
    do; it is something there is no way to ask for.

    Args:
        x_csrf_token (str | Unset):
        body (UpdateSelfRequest): Body of `PATCH /users/me`. One field, and that is the design: a
            schema
            with no `platformRole` in it is a request that cannot ask for one, which
            is a stronger guarantee than a handler that declines to honour it
            (PLAN.md §4).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Problem | User]
    """

    kwargs = _get_kwargs(
        body=body,
        x_csrf_token=x_csrf_token,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
    body: UpdateSelfRequest,
    x_csrf_token: str | Unset = UNSET,
) -> Problem | User | None:
    """Change your own display name.

     The self-service half of user administration, and deliberately the
    smallest one that could exist: a display name and nothing else.

    Your platform role and your status are **not** fields of this request.
    Not filtered out of it — absent from it, so a body that names one is
    rejected by the request validator with `400` before any handler runs
    (PLAN.md §4: "field safety comes from the schema"). Raising your own
    privilege through this endpoint is not something the server declines to
    do; it is something there is no way to ask for.

    Args:
        x_csrf_token (str | Unset):
        body (UpdateSelfRequest): Body of `PATCH /users/me`. One field, and that is the design: a
            schema
            with no `platformRole` in it is a request that cannot ask for one, which
            is a stronger guarantee than a handler that declines to honour it
            (PLAN.md §4).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Problem | User
    """

    return (
        await asyncio_detailed(
            client=client,
            body=body,
            x_csrf_token=x_csrf_token,
        )
    ).parsed
