"""AIMeter Context Management and W3C Baggage / Trace Propagation."""

import contextvars
import uuid
from typing import Any, Dict, Optional


_current_trace_id: contextvars.ContextVar[Optional[str]] = contextvars.ContextVar(
    "_current_trace_id", default=None
)
_current_span_id: contextvars.ContextVar[Optional[str]] = contextvars.ContextVar(
    "_current_span_id", default=None
)
_current_parent_span_id: contextvars.ContextVar[Optional[str]] = contextvars.ContextVar(
    "_current_parent_span_id", default=None
)
_current_tree_depth: contextvars.ContextVar[int] = contextvars.ContextVar(
    "_current_tree_depth", default=0
)
_current_baggage: contextvars.ContextVar[Dict[str, str]] = contextvars.ContextVar(
    "_current_baggage", default={}
)


def generate_id() -> str:
    """Generate a clean UUID string without hyphens."""
    return uuid.uuid4().hex


def get_current_trace_id() -> str:
    """Get the active trace ID or initialize a new one if none exists."""
    tid = _current_trace_id.get()
    if not tid:
        tid = f"tr_{generate_id()}"
        _current_trace_id.set(tid)
    return tid


def set_current_trace_id(trace_id: str) -> contextvars.Token:
    """Explicitly set the trace ID in the current execution context."""
    return _current_trace_id.set(trace_id)


def get_current_span_id() -> Optional[str]:
    """Get the current active span ID."""
    return _current_span_id.get()


def get_current_parent_span_id() -> Optional[str]:
    """Get the current active parent span ID."""
    return _current_parent_span_id.get()


def get_current_tree_depth() -> int:
    """Get the current recursive/DAG execution depth."""
    return _current_tree_depth.get()


def get_current_baggage() -> Dict[str, str]:
    """Get the current W3C baggage dictionary."""
    return dict(_current_baggage.get() or {})


def update_current_baggage(updates: Dict[str, str]) -> contextvars.Token:
    """Merge new metadata into the active baggage."""
    existing = get_current_baggage()
    existing.update(updates)
    return _current_baggage.set(existing)


def format_w3c_baggage(baggage: Dict[str, str]) -> str:
    """Format a dictionary into a W3C baggage header string."""
    pairs = [f"{k.strip()}={v.strip()}" for k, v in baggage.items() if k and v]
    return ",".join(pairs)


def parse_w3c_baggage(baggage_str: str) -> Dict[str, str]:
    """Parse a W3C baggage header string into a dictionary."""
    result = {}
    if not baggage_str:
        return result
    for item in baggage_str.split(","):
        if "=" in item:
            k, v = item.strip().split("=", 1)
            result[k.strip()] = v.strip()
    return result
