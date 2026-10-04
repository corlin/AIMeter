"""AIMeter LlamaIndex Callback Handler Adapter."""

from typing import Any, Dict, List, Optional

try:
    from llama_index.core.callbacks.base_handler import BaseCallbackHandler
    from llama_index.core.callbacks.schema import CBEventType, EventPayload
except ImportError:
    class BaseCallbackHandler:  # type: ignore
        """Fallback mock if llama_index is not installed."""
        def __init__(self, *args: Any, **kwargs: Any) -> None:
            pass

    class CBEventType:  # type: ignore
        LLM = "llm"
        FUNCTION_CALL = "function_call"

    class EventPayload:  # type: ignore
        RESPONSE = "response"
        PROMPT = "prompt"

from ..client import AIMeterClient, get_default_client
from ..context import get_current_tree_depth
from ..exceptions import CircuitBreakerOpenError
from ..tracer import Span


class AIMeterLlamaIndexCallbackHandler(BaseCallbackHandler):
    """LlamaIndex callback handler that intercepts LLM and tool querying for cost accounting."""

    def __init__(
        self,
        client: Optional[AIMeterClient] = None,
        workflow_id: str = "llamaindex-flow",
        tenant_id: Optional[str] = None,
        pre_check: bool = True,
    ):
        super().__init__(event_starts_to_ignore=[], event_ends_to_ignore=[])
        self.client = client or get_default_client()
        self.workflow_id = workflow_id
        self.tenant_id = tenant_id or self.client.tenant_id
        self.pre_check = pre_check
        self._active_spans: Dict[str, Span] = {}

    def on_event_start(
        self,
        event_type: CBEventType,
        payload: Optional[Dict[str, Any]] = None,
        event_id: str = "",
        parent_id: str = "",
        **kwargs: Any,
    ) -> str:
        """Run Active Guard and start span when LLM/Function event begins."""
        if str(event_type) in [str(CBEventType.LLM), "llm"]:
            model = "gpt-4o"
            if payload and "serialized" in payload:
                model = payload["serialized"].get("model_name") or model

            if self.pre_check:
                depth = get_current_tree_depth()
                guard_resp = self.client.check_guard(
                    workflow_id=self.workflow_id,
                    model=model,
                    tree_depth=depth,
                    tenant_id=self.tenant_id,
                )
                if not guard_resp.allowed:
                    raise CircuitBreakerOpenError(
                        message=f"Active Guard blocked LlamaIndex execution: {guard_resp.reason}",
                        decision_code=guard_resp.decision_code,
                        reason=guard_resp.reason,
                        circuit_state=guard_resp.circuit_state,
                        fallback_model=guard_resp.fallback_model,
                    )
                if guard_resp.fallback_model:
                    model = guard_resp.fallback_model

            span = Span(
                name=f"llamaindex_{event_type}",
                client=self.client,
                model=model,
                workflow_id=self.workflow_id,
            )
            span.__enter__()
            self._active_spans[event_id] = span

        return event_id

    def on_event_end(
        self,
        event_type: CBEventType,
        payload: Optional[Dict[str, Any]] = None,
        event_id: str = "",
        **kwargs: Any,
    ) -> None:
        """Extract tokens and close span when event ends."""
        span = self._active_spans.pop(event_id, None)
        if not span:
            return

        if payload:
            response = payload.get(EventPayload.RESPONSE) or payload.get("response")
            # Check for raw response / usage
            if hasattr(response, "raw") and hasattr(response.raw, "usage"):
                u = response.raw.usage
                span.record_tokens(
                    input_tokens=getattr(u, "prompt_tokens", 0),
                    output_tokens=getattr(u, "completion_tokens", 0),
                )
            elif isinstance(response, dict) and "usage" in response:
                u = response["usage"]
                span.record_tokens(
                    input_tokens=u.get("prompt_tokens", 0),
                    output_tokens=u.get("completion_tokens", 0),
                )

        span.__exit__(None, None, None)

    def start_trace(self, trace_id: Optional[str] = None) -> None:
        pass

    def end_trace(
        self,
        trace_id: Optional[str] = None,
        trace_map: Optional[Dict[str, List[str]]] = None,
    ) -> None:
        pass
