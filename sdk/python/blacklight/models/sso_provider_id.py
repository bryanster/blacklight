from enum import StrEnum


class SSOProviderId(StrEnum):
    OIDC = "oidc"
    SAML = "saml"

    def __str__(self) -> str:
        return str(self.value)
