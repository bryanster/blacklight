from enum import StrEnum


class ProblemCode(StrEnum):
    CONFLICT = "conflict"
    FORBIDDEN = "forbidden"
    INTERNAL = "internal"
    METHOD_NOT_ALLOWED = "method_not_allowed"
    MFA_ENROLMENT_REQUIRED = "mfa_enrolment_required"
    NOT_FOUND = "not_found"
    PAYLOAD_TOO_LARGE = "payload_too_large"
    RATE_LIMITED = "rate_limited"
    UNAUTHENTICATED = "unauthenticated"
    VALIDATION_FAILED = "validation_failed"

    def __str__(self) -> str:
        return str(self.value)
