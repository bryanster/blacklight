from enum import StrEnum


class LoginStatus(StrEnum):
    AUTHENTICATED = "authenticated"
    MFA_ENROLMENT_REQUIRED = "mfa_enrolment_required"
    MFA_REQUIRED = "mfa_required"

    def __str__(self) -> str:
        return str(self.value)
