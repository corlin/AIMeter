"""AIMeter Tracer, @meter.trace Decorator and Context Manager."""

import asyncio
from functools import wraps
import time
from typing import Any, Callable, Dict, Optional, TypeVar, Union

from .client import AIMeterClient, get_default_client
from .context import (
    _current_parent_span_id,
    _current_span_id,
    _current_tree_depth,
    generate_id,
    get_current_parent_span_id,
    get_current_span_id,
    get_current_trace_id,
    get_current_tree_depth,
    update_current_baggage,
)
from .exceptions import CircuitBreakerOpenError

F = TypeVar("F", bound=Callable[..., Any])


class Span:
    """Represents a single unit of execution or AI call within an execution trace."""

    def __init__(
        self,
        name: str,
        client: Optional[AIMeterClient] = None,
        model: str = "gpt-4o",
        provider: str = "openai",
        workflow_id: Optional[str] = None,
        attributes: Optional[Dict[str, str]] = None,
    ):
        self.name = name
        self.client = client or get_default_client()
        self.model = model
        self.provider = provider
        self.workflow_id = workflow_id
        self.attributes: Dict[str, str] = attributes or {}

        self.trace_id = get_current_trace_id()
        self.parent_span_id = get_current_span_id()
        self.span_id = f"sp_{generate_id()}"
        self.tree_depth = get_current_tree_depth()

        self.start_time = time.time()
        self._tokens_recorded = False
        self._input_tokens = 0
        self._output_tokens = 0
        self._cached_tokens = 0
        self._reasoning_tokens = 0

        self._token_span: Any = None
        self._token_parent: Any = None
        self._token_depth: Any = None

    def __enter__(self) -> "Span":
        self._token_parent = _current_parent_span_id.set(self.parent_span_id)
        self._token_span = _current_span_id.set(self.span_id)
        self._token_depth = _current_tree_depth.set(self.tree_depth + 1)
        if self.workflow_id:
            update_current_baggage({"workflow_id": self.workflow_id})
        return self

    def __exit__(self, exc_type: Any, exc_val: Any, exc_tb: Any) -> None:
        self.end()
        if self._token_span:
            _current_span_id.reset(self._token_span)
        if self._token_parent:
            _current_parent_span_id.reset(self._token_parent)
        if self._token_depth:
            _current_tree_depth.reset(self._token_depth)

    def record_tokens(
        self,
        input_tokens: int = 0,
        output_tokens: int = 0,
        cached_tokens: int = 0,
        reasoning_tokens: int = 0,
    ) -> None:
        """Manually record token counts for this span."""
        self._input_tokens = input_tokens
        self._output_tokens = output_tokens
        self._cached_tokens = cached_tokens
        self._reasoning_tokens = reasoning_tokens
        self._tokens_recorded = True

    def set_attribute(self, key: str, value: str) -> None:
        """Add metadata attributes to this span."""
        self.attributes[key] = str(value)

    def end(self) -> None:
        """Compute latency and enqueue the usage event."""
        latency_ms = max(1, int((time.time() - self.start_time) * 1000))
        baggage = {}
        if self.workflow_id:
            baggage["workflow_id"] = self.workflow_id

        self.client.record_usage(
            trace_id=self.trace_id,
            span_id=self.span_id,
            parent_span_id=self.parent_span_id,
            provider=self.provider,
            model=self.model,
            latency_ms=latency_ms,
            input_tokens=self._input_tokens,
            output_tokens=self._output_tokens,
            cached_tokens=self._cached_tokens,
            reasoning_tokens=self._reasoning_tokens,
            baggage=baggage,
            attributes=self.attributes,
        )


def _sniff_tokens(result: Any, span: Span) -> None:
    """Introspect common LLM provider response objects and extract token metrics."""
    if span._tokens_recorded or result is None:
        return

    # 1. OpenAI / LiteLLM Completion Object
    if hasattr(result, "usage") and result.usage is not None:
        u = result.usage
        prompt_tokens = getattr(u, "prompt_tokens", 0)
        completion_tokens = getattr(u, "completion_tokens", 0)

        cached_tokens = 0
        prompt_details = getattr(u, "prompt_tokens_details", None)
        if prompt_details:
            cached_tokens = getattr(prompt_details, "cached_tokens", 0)

        reasoning_tokens = 0
        completion_details = getattr(u, "completion_tokens_details", None)
        if completion_details:
            reasoning_tokens = getattr(completion_details, "reasoning_tokens", 0)

        span.record_tokens(
            input_tokens=prompt_tokens,
            output_tokens=completion_tokens,
            cached_tokens=cached_tokens,
            reasoning_tokens=reasoning_tokens,
        )
        return

    # 2. Anthropic Message Object
    if hasattr(result, "usage") and hasattr(result.usage, "input_tokens"):
        u = result.usage
        input_tokens = getattr(u, "input_tokens", 0)
        output_tokens = getattr(u, "output_tokens", 0)
        cached_tokens = getattr(u, "cache_read_input_tokens", 0)
        span.record_tokens(
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            cached_tokens=cached_tokens,
        )
        return

    # 3. Dict-like response
    if isinstance(result, dict) and "usage" in result:
        u = result["usage"]
        if isinstance(u, dict):
            span.record_tokens(
                input_tokens=u.get("prompt_tokens", u.get("input_tokens", 0)),
                output_tokens=u.get("completion_tokens", u.get("output_tokens", 0)),
                cached_tokens=u.get("cached_tokens", 0),
                reasoning_tokens=u.get("reasoning_tokens", 0),
            )


class Tracer:
    """Tracer interface for decorating functions and initiating spans."""

    def __init__(self, client: Optional[AIMeterClient] = None):
        self._client = client

    @property
    def client(self) -> AIMeterClient:
        return self._client or get_default_client()

    def span(
        self,
        name: str,
        model: str = "gpt-4o",
        provider: str = "openai",
        workflow_id: Optional[str] = None,
        attributes: Optional[Dict[str, str]] = None,
    ) -> Span:
        """Create a new execution span context manager."""
        return Span(
            name=name,
            client=self.client,
            model=model,
            provider=provider,
            workflow_id=workflow_id,
            attributes=attributes,
        )

    def trace(
        self,
        name: Optional[str] = None,
        model: str = "gpt-4o",
        provider: str = "openai",
        workflow_id: Optional[str] = None,
        pre_check: bool = False,
        fallback_fn: Optional[Callable[..., Any]] = None,
    ) -> Callable[[F], F]:
        """Decorator to trace function execution and optionally perform Active Guard pre-checks."""

        def decorator(func: F) -> F:
            span_name = name or func.__name__
            wf_id = workflow_id or "default-flow"

            if asyncio.iscoroutinefunction(func):
                @wraps(func)
                async def async_wrapper(*args: Any, **kwargs: Any) -> Any:
                    target_model = model
                    if pre_check:
                        depth = get_current_tree_depth()
                        guard_resp = await self.client.check_guard_async(
                            workflow_id=wf_id,
                            model=target_model,
                            tree_depth=depth,
                        )
                        if not guard_resp.allowed:
                            if fallback_fn:
                                return await fallback_fn(guard_resp.fallback_model, *args, **kwargs)
                            raise CircuitBreakerOpenError(
                                message=f"Active Guard blocked execution: {guard_resp.reason}",
                                decision_code=guard_resp.decision_code,
                                reason=guard_resp.reason,
                                circuit_state=guard_resp.circuit_state,
                                fallback_model=guard_resp.fallback_model,
                            )
                        if guard_resp.fallback_model:
                            target_model = guard_resp.fallback_model

                    with self.span(
                        name=span_name,
                        model=target_model,
                        provider=provider,
                        workflow_id=wf_id,
                    ) as sp:
                        result = await func(*args, **kwargs)
                        _sniff_tokens(result, sp)
                        return result

                return async_wrapper  # type: ignore
            else:
                @wraps(func)
                def sync_wrapper(*args: Any, **kwargs: Any) -> Any:
                    target_model = model
                    if pre_check:
                        depth = get_current_tree_depth()
                        guard_resp = self.client.check_guard(
                            workflow_id=wf_id,
                            model=target_model,
                            tree_depth=depth,
                        )
                        if not guard_resp.allowed:
                            if fallback_fn:
                                return fallback_fn(guard_resp.fallback_model, *args, **kwargs)
                            raise CircuitBreakerOpenError(
                                message=f"Active Guard blocked execution: {guard_resp.reason}",
                                decision_code=guard_resp.decision_code,
                                reason=guard_resp.reason,
                                circuit_state=guard_resp.circuit_state,
                                fallback_model=guard_resp.fallback_model,
                            )
                        if guard_resp.fallback_model:
                            target_model = guard_resp.fallback_model

                    with self.span(
                        name=span_name,
                        model=target_model,
                        provider=provider,
                        workflow_id=wf_id,
                    ) as sp:
                        result = func(*args, **kwargs)
                        _sniff_tokens(result, sp)
                        return result

                return sync_wrapper  # type: ignore

        return decorator


# Global tracer singleton
meter = Tracer()
