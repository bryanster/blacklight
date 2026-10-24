from http import HTTPStatus
from typing import Any, cast
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...types import Response, UNSET
from ... import errors

from ...models.create_service_token_request import CreateServiceTokenRequest
from ...models.created_service_token import CreatedServiceToken
from ...models.problem import Problem
from ...types import UNSET, Unset
from typing import cast


def _get_kwargs(
    *,
    body: CreateServiceTokenRequest,
    x_csrf_token: str | Unset = UNSET,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}
    if not isinstance(x_csrf_token, Unset):
        headers["X-CSRF-Token"] = x_csrf_token

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/auth/tokens",
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> CreatedServiceToken | Problem | None:
    if response.status_code == 201:
        response_201 = CreatedServiceToken.from_dict(response.json())

        return response_201

    if response.status_code == 400:
        response_400 = Problem.from_dict(response.json())

        return response_400

    if response.status_code == 401:
        response_401 = Problem.from_dict(response.json())

        return response_401

    if response.status_code == 403:
        response_403 = Problem.from_dict(response.json())

        return response_403

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
) -> Response[CreatedServiceToken | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
    body: CreateServiceTokenRequest,
    x_csrf_token: str | Unset = UNSET,
) -> Response[CreatedServiceToken | Problem]:
    """Create a service token, shown once.

     Mints a token owned by you. **The response is the only time its secret
    exists outside your client** — it is stored as a hash and cannot be
    recovered, so a caller that does not save it has to create another one.

    What the token may do is fenced twice, and the narrower fence wins on
    every request:

    1. The scopes you give it here.
    2. Your own permissions, read live on every request the token makes.
       Being demoted narrows every token you hold, immediately, without
       touching them; having your account disabled stops them entirely.

    `engagementId` adds a third fence, and only ever subtracts: a token
    bound to an engagement reaches that engagement and nothing else — not
    another engagement, not the installation — whatever its scopes say. It
    cannot let a token into an engagement you are not a member of.

    `expiresAt` is required and is capped at a year out. A credential with
    no expiry is one nobody remembers to revoke.

    A service token cannot call this. Creating credentials takes a signed-in
    session, so that a leaked token cannot mint a longer-lived replacement
    for itself and outlive its own revocation — `403` with `code:
    "forbidden"` when one tries, whatever scopes it carries.

    Args:
        x_csrf_token (str | Unset):
        body (CreateServiceTokenRequest): Body of `POST /auth/tokens`. The owner is the caller and
            is not a field.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[CreatedServiceToken | Problem]
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
    body: CreateServiceTokenRequest,
    x_csrf_token: str | Unset = UNSET,
) -> CreatedServiceToken | Problem | None:
    """Create a service token, shown once.

     Mints a token owned by you. **The response is the only time its secret
    exists outside your client** — it is stored as a hash and cannot be
    recovered, so a caller that does not save it has to create another one.

    What the token may do is fenced twice, and the narrower fence wins on
    every request:

    1. The scopes you give it here.
    2. Your own permissions, read live on every request the token makes.
       Being demoted narrows every token you hold, immediately, without
       touching them; having your account disabled stops them entirely.

    `engagementId` adds a third fence, and only ever subtracts: a token
    bound to an engagement reaches that engagement and nothing else — not
    another engagement, not the installation — whatever its scopes say. It
    cannot let a token into an engagement you are not a member of.

    `expiresAt` is required and is capped at a year out. A credential with
    no expiry is one nobody remembers to revoke.

    A service token cannot call this. Creating credentials takes a signed-in
    session, so that a leaked token cannot mint a longer-lived replacement
    for itself and outlive its own revocation — `403` with `code:
    "forbidden"` when one tries, whatever scopes it carries.

    Args:
        x_csrf_token (str | Unset):
        body (CreateServiceTokenRequest): Body of `POST /auth/tokens`. The owner is the caller and
            is not a field.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        CreatedServiceToken | Problem
    """

    return sync_detailed(
        client=client,
        body=body,
        x_csrf_token=x_csrf_token,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
    body: CreateServiceTokenRequest,
    x_csrf_token: str | Unset = UNSET,
) -> Response[CreatedServiceToken | Problem]:
    """Create a service token, shown once.

     Mints a token owned by you. **The response is the only time its secret
    exists outside your client** — it is stored as a hash and cannot be
    recovered, so a caller that does not save it has to create another one.

    What the token may do is fenced twice, and the narrower fence wins on
    every request:

    1. The scopes you give it here.
    2. Your own permissions, read live on every request the token makes.
       Being demoted narrows every token you hold, immediately, without
       touching them; having your account disabled stops them entirely.

    `engagementId` adds a third fence, and only ever subtracts: a token
    bound to an engagement reaches that engagement and nothing else — not
    another engagement, not the installation — whatever its scopes say. It
    cannot let a token into an engagement you are not a member of.

    `expiresAt` is required and is capped at a year out. A credential with
    no expiry is one nobody remembers to revoke.

    A service token cannot call this. Creating credentials takes a signed-in
    session, so that a leaked token cannot mint a longer-lived replacement
    for itself and outlive its own revocation — `403` with `code:
    "forbidden"` when one tries, whatever scopes it carries.

    Args:
        x_csrf_token (str | Unset):
        body (CreateServiceTokenRequest): Body of `POST /auth/tokens`. The owner is the caller and
            is not a field.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[CreatedServiceToken | Problem]
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
    body: CreateServiceTokenRequest,
    x_csrf_token: str | Unset = UNSET,
) -> CreatedServiceToken | Problem | None:
    """Create a service token, shown once.

     Mints a token owned by you. **The response is the only time its secret
    exists outside your client** — it is stored as a hash and cannot be
    recovered, so a caller that does not save it has to create another one.

    What the token may do is fenced twice, and the narrower fence wins on
    every request:

    1. The scopes you give it here.
    2. Your own permissions, read live on every request the token makes.
       Being demoted narrows every token you hold, immediately, without
       touching them; having your account disabled stops them entirely.

    `engagementId` adds a third fence, and only ever subtracts: a token
    bound to an engagement reaches that engagement and nothing else — not
    another engagement, not the installation — whatever its scopes say. It
    cannot let a token into an engagement you are not a member of.

    `expiresAt` is required and is capped at a year out. A credential with
    no expiry is one nobody remembers to revoke.

    A service token cannot call this. Creating credentials takes a signed-in
    session, so that a leaked token cannot mint a longer-lived replacement
    for itself and outlive its own revocation — `403` with `code:
    "forbidden"` when one tries, whatever scopes it carries.

    Args:
        x_csrf_token (str | Unset):
        body (CreateServiceTokenRequest): Body of `POST /auth/tokens`. The owner is the caller and
            is not a field.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        CreatedServiceToken | Problem
    """

    return (
        await asyncio_detailed(
            client=client,
            body=body,
            x_csrf_token=x_csrf_token,
        )
    ).parsed
