"""AIMeter Exceptions Module."""

from typing import Optional


class AIMeterError(Exception):
    """Base exception for all AIMeter SDK errors."""
    pass


class CircuitBreakerOpenError(AIMeterError):
    """Raised when Active Guard blocks an execution due to circuit breaker trip or runaway loop."""

    def __init__(
        self,
        message: str,
        decision_code: str = "CIRCUIT_OPEN",
        reason: str = "",
        circuit_state: str = "OPEN",
        fallback_model: Optional[str] = None,
    ):
        super().__init__(message)
        self.decision_code = decision_code
        self.reason = reason
        self.circuit_state = circuit_state
        self.fallback_model = fallback_model

    def __repr__(self) -> str:
        return (
            f"CircuitBreakerOpenError(decision_code='{self.decision_code}', "
            f"circuit_state='{self.circuit_state}', fallback_model='{self.fallback_model}', "
            f"reason='{self.reason}')"
        )


class BudgetExceededError(CircuitBreakerOpenError):
    """Raised specifically when a budget limit is exceeded."""
    pass


class ConfigurationError(AIMeterError):
    """Raised when the AIMeter client configuration is invalid."""
    pass
