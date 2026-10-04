"""AI Meter Python SDK

Enterprise AI Usage & Cost Control Plane - Official Python SDK.
"""

from .client import AIMeterClient, get_default_client, init
from .exceptions import (
    AIMeterError,
    BudgetExceededError,
    CircuitBreakerOpenError,
    ConfigurationError,
)
from .tracer import Span, Tracer, meter

AIMeter = AIMeterClient

__version__ = "0.1.0"

__all__ = [
    "AIMeter",
    "AIMeterClient",
    "Tracer",
    "Span",
    "meter",
    "init",
    "get_default_client",
    "AIMeterError",
    "CircuitBreakerOpenError",
    "BudgetExceededError",
    "ConfigurationError",
]
