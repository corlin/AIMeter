"""AIMeter Client Module."""

from dataclasses import dataclass
import os
import time
from typing import Any, Dict, Optional
import httpx
import logging

from .context import format_w3c_baggage, get_current_baggage
from .exceptions import CircuitBreakerOpenError
from .reporter import BackgroundReporter

logger = logging.getLogger("aimeter")


@dataclass
class GuardResponse:
    """Active Guard Pre-Check evaluation response."""
    allowed: bool
    decision_code: str
    reason: str
    circuit_state: str
    fallback_model: Optional[str] = None
    retry_after_seconds: Optional[int] = None


class AIMeterClient:
    """Primary client for interacting with AI Meter Control Plane."""

    def __init__(
        self,
        base_url: Optional[str] = None,
        tenant_id: Optional[str] = None,
        api_key: Optional[str] = None,
        guard_timeout_seconds: float = 0.02,  # 20ms fail-open timeout
        batch_size: int = 50,
        flush_interval_seconds: float = 0.5,
    ):
        self.base_url = (base_url or os.getenv("AIMETER_BASE_URL", "http://localhost:8080")).rstrip("/")
        self.tenant_id = tenant_id or os.getenv("AIMETER_TENANT_ID", "default")
        self.api_key = api_key or os.getenv("AIMETER_API_KEY", "")
        self.guard_timeout = guard_timeout_seconds

        self._headers = {"Content-Type": "application/json"}
        if self.api_key:
            self._headers["Authorization"] = f"Bearer {self.api_key}"

        # Sync HTTP Client for Pre-Check (<2ms)
        self._http_client = httpx.Client(
            timeout=self.guard_timeout,
            limits=httpx.Limits(max_keepalive_connections=20, max_connections=50),
        )

        # Background Reporter for non-blocking usage ingestion
        events_endpoint = f"{self.base_url}/api/v1/events"
        self._reporter = BackgroundReporter(
            endpoint_url=events_endpoint,
            batch_size=batch_size,
            flush_interval_seconds=flush_interval_seconds,
            headers=self._headers,
        )

    def check_guard(
        self,
        workflow_id: str,
        model: str,
        tree_depth: int = 0,
        tenant_id: Optional[str] = None,
        estimated_prompt_tokens: int = 0,
    ) -> GuardResponse:
        """Synchronously check Active Guard before making an LLM call with Fail-Open safety."""
        target_tenant = tenant_id or self.tenant_id
        payload = {
            "tenant_id": target_tenant,
            "workflow_id": workflow_id,
            "model": model,
            "current_tree_depth": tree_depth,
            "estimated_prompt_tokens": estimated_prompt_tokens,
        }

        url = f"{self.base_url}/v1/guard/check"
        try:
            resp = self._http_client.post(url, json=payload, headers=self._headers)
            if resp.status_code == 200:
                data = resp.json()
                return GuardResponse(
                    allowed=data.get("allowed", True),
                    decision_code=data.get("decision_code", "OK"),
                    reason=data.get("reason", "Permitted"),
                    circuit_state=data.get("circuit_state", "CLOSED"),
                    fallback_model=data.get("fallback_model"),
                    retry_after_seconds=data.get("retry_after_seconds"),
                )
            elif resp.status_code == 429:  # Blocked / Circuit Breaker OPEN
                data = resp.json()
                return GuardResponse(
                    allowed=False,
                    decision_code=data.get("decision_code", "CIRCUIT_OPEN"),
                    reason=data.get("reason", "Circuit breaker active"),
                    circuit_state=data.get("circuit_state", "OPEN"),
                    fallback_model=data.get("fallback_model"),
                    retry_after_seconds=data.get("retry_after_seconds"),
                )
            else:
                logger.warning(
                    f"[AIMeter] Guard check returned unexpected status {resp.status_code}, failing open."
                )
                return GuardResponse(
                    allowed=True,
                    decision_code="FAIL_OPEN",
                    reason=f"HTTP status {resp.status_code}",
                    circuit_state="CLOSED",
                )
        except Exception as e:
            # Fail-Open: Control plane unreachable or timeout must NEVER crash user application
            logger.debug(f"[AIMeter] Guard pre-check timed out or failed: {e}. Failing open.")
            return GuardResponse(
                allowed=True,
                decision_code="FAIL_OPEN",
                reason=f"Control plane error: {e}",
                circuit_state="CLOSED",
            )

    async def check_guard_async(
        self,
        workflow_id: str,
        model: str,
        tree_depth: int = 0,
        tenant_id: Optional[str] = None,
        estimated_prompt_tokens: int = 0,
    ) -> GuardResponse:
        """Asynchronously check Active Guard before making an LLM call."""
        target_tenant = tenant_id or self.tenant_id
        payload = {
            "tenant_id": target_tenant,
            "workflow_id": workflow_id,
            "model": model,
            "current_tree_depth": tree_depth,
            "estimated_prompt_tokens": estimated_prompt_tokens,
        }
        url = f"{self.base_url}/v1/guard/check"

        try:
            async with httpx.AsyncClient(timeout=self.guard_timeout) as client:
                resp = await client.post(url, json=payload, headers=self._headers)
                data = resp.json()
                return GuardResponse(
                    allowed=data.get("allowed", True),
                    decision_code=data.get("decision_code", "OK"),
                    reason=data.get("reason", "Permitted"),
                    circuit_state=data.get("circuit_state", "CLOSED"),
                    fallback_model=data.get("fallback_model"),
                    retry_after_seconds=data.get("retry_after_seconds"),
                )
        except Exception as e:
            return GuardResponse(
                allowed=True,
                decision_code="FAIL_OPEN",
                reason=f"Async guard check error: {e}",
                circuit_state="CLOSED",
            )

    def record_usage(
        self,
        trace_id: str,
        span_id: str,
        parent_span_id: Optional[str] = None,
        provider: str = "openai",
        model: str = "gpt-4o",
        latency_ms: int = 0,
        input_tokens: int = 0,
        output_tokens: int = 0,
        cached_tokens: int = 0,
        reasoning_tokens: int = 0,
        baggage: Optional[Dict[str, str]] = None,
        attributes: Optional[Dict[str, str]] = None,
    ) -> None:
        """Enqueue an event for non-blocking asynchronous transmission."""
        combined_baggage = get_current_baggage()
        if baggage:
            combined_baggage.update(baggage)
        if "tenant_id" not in combined_baggage:
            combined_baggage["tenant_id"] = self.tenant_id

        attr = attributes or {}
        attr["prompt_tokens"] = str(input_tokens)
        attr["completion_tokens"] = str(output_tokens)
        if cached_tokens > 0:
            attr["cached_tokens"] = str(cached_tokens)
        if reasoning_tokens > 0:
            attr["reasoning_tokens"] = str(reasoning_tokens)

        payload = {
            "trace_id": trace_id,
            "span_id": span_id,
            "parent_span_id": parent_span_id or "",
            "provider": provider,
            "model": model,
            "latency_ms": latency_ms,
            "baggage": format_w3c_baggage(combined_baggage),
            "attributes": attr,
        }

        self._reporter.enqueue(payload)

    def flush(self, timeout: float = 3.0) -> None:
        """Flush the usage event queue."""
        self._reporter.flush(timeout=timeout)

    def close(self) -> None:
        """Close connections and flush pending usage events."""
        self._reporter.shutdown()
        try:
            self._http_client.close()
        except Exception:
            pass


# Global singleton holder
_global_client: Optional[AIMeterClient] = None


def get_default_client() -> AIMeterClient:
    """Retrieve or initialize the global default AIMeterClient instance."""
    global _global_client
    if _global_client is None:
        _global_client = AIMeterClient()
    return _global_client


def init(
    base_url: Optional[str] = None,
    tenant_id: Optional[str] = None,
    api_key: Optional[str] = None,
    **kwargs: Any,
) -> AIMeterClient:
    """Initialize the global default AIMeterClient singleton."""
    global _global_client
    _global_client = AIMeterClient(
        base_url=base_url,
        tenant_id=tenant_id,
        api_key=api_key,
        **kwargs,
    )
    return _global_client
