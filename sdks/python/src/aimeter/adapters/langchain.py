"""AIMeter LangChain Callback Handler Adapter."""

import time
from typing import Any, Dict, List, Optional
from uuid import UUID

try:
    from langchain_core.callbacks import BaseCallbackHandler
    from langchain_core.outputs import LLMResult
except ImportError:
    class BaseCallbackHandler:  # type: ignore
        """Fallback mock if langchain_core is not installed."""
        pass
    LLMResult = Any  # type: ignore

from ..client import AIMeterClient, get_default_client
from ..context import get_current_tree_depth
from ..exceptions import CircuitBreakerOpenError
from ..tracer import Span


class AIMeterCallbackHandler(BaseCallbackHandler):
    """LangChain callback handler that automatically tracks LLM/Tool costs and enforces Active Guard."""

    def __init__(
        self,
        client: Optional[AIMeterClient] = None,
        workflow_id: str = "langchain-workflow",
        tenant_id: Optional[str] = None,
        pre_check: bool = True,
    ):
        super().__init__()
        self.client = client or get_default_client()
        self.workflow_id = workflow_id
        self.tenant_id = tenant_id or self.client.tenant_id
        self.pre_check = pre_check
        self._active_spans: Dict[str, Span] = {}

    def on_llm_start(
        self,
        serialized: Dict[str, Any],
        prompts: List[str],
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        tags: Optional[List[str]] = None,
        metadata: Optional[Dict[str, Any]] = None,
        **kwargs: Any,
    ) -> None:
        """Invoked when LLM starts running. Performs Active Guard check and initializes Span."""
        params = kwargs.get("invocation_params", {})
        model = params.get("model_name") or params.get("model") or "gpt-4o"
        provider = params.get("_type") or "openai"

        # 1. Active Guard Pre-Check
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
                    message=f"Active Guard blocked LangChain execution: {guard_resp.reason}",
                    decision_code=guard_resp.decision_code,
                    reason=guard_resp.reason,
                    circuit_state=guard_resp.circuit_state,
                    fallback_model=guard_resp.fallback_model,
                )
            if guard_resp.fallback_model:
                model = guard_resp.fallback_model

        # 2. Open Span
        span = Span(
            name=f"langchain_llm_{model}",
            client=self.client,
            model=model,
            provider=provider,
            workflow_id=self.workflow_id,
        )
        span.__enter__()
        self._active_spans[str(run_id)] = span

    def on_llm_end(
        self,
        response: LLMResult,
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        **kwargs: Any,
    ) -> None:
        """Invoked when LLM finishes running. Extracts token counts and records usage."""
        span = self._active_spans.pop(str(run_id), None)
        if not span:
            return

        input_tokens = 0
        output_tokens = 0
        cached_tokens = 0
        reasoning_tokens = 0

        # Try extracting from llm_output
        if response.llm_output:
            usage = response.llm_output.get("token_usage") or response.llm_output.get("usage") or {}
            if isinstance(usage, dict):
                input_tokens = usage.get("prompt_tokens") or usage.get("input_tokens") or 0
                output_tokens = usage.get("completion_tokens") or usage.get("output_tokens") or 0
                cached_tokens = usage.get("cached_tokens") or 0
                reasoning_tokens = usage.get("reasoning_tokens") or 0

        # Try extracting from generation metadata
        if input_tokens == 0 and response.generations:
            for gen_list in response.generations:
                for gen in gen_list:
                    gen_info = getattr(gen, "generation_info", {}) or {}
                    meta = gen_info.get("usage_metadata") or {}
                    if meta:
                        input_tokens += meta.get("input_tokens", 0)
                        output_tokens += meta.get("output_tokens", 0)

        span.record_tokens(
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            cached_tokens=cached_tokens,
            reasoning_tokens=reasoning_tokens,
        )
        span.__exit__(None, None, None)

    def on_llm_error(
        self,
        error: BaseException,
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        **kwargs: Any,
    ) -> None:
        """Close span on LLM failure."""
        span = self._active_spans.pop(str(run_id), None)
        if span:
            span.set_attribute("error", str(error))
            span.__exit__(type(error), error, None)

    def on_tool_start(
        self,
        serialized: Dict[str, Any],
        input_str: str,
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        tags: Optional[List[str]] = None,
        metadata: Optional[Dict[str, Any]] = None,
        **kwargs: Any,
    ) -> None:
        """Track Tool execution start."""
        tool_name = serialized.get("name") or "tool"
        span = Span(
            name=f"tool_{tool_name}",
            client=self.client,
            model="tool",
            provider="local",
            workflow_id=self.workflow_id,
        )
        span.__enter__()
        self._active_spans[str(run_id)] = span

    def on_tool_end(
        self,
        output: str,
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        **kwargs: Any,
    ) -> None:
        """Track Tool execution completion."""
        span = self._active_spans.pop(str(run_id), None)
        if span:
            span.set_attribute("tool.output_length", str(len(output)))
            span.__exit__(None, None, None)
