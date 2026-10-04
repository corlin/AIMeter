from dataclasses import dataclass
from typing import Any, Optional
import pytest

from aimeter import CircuitBreakerOpenError, meter
from aimeter.client import AIMeterClient, GuardResponse
from aimeter.context import get_current_tree_depth
from aimeter.tracer import Tracer


@dataclass
class MockUsage:
    prompt_tokens: int = 120
    completion_tokens: int = 40
    prompt_tokens_details: Optional[Any] = None
    completion_tokens_details: Optional[Any] = None


@dataclass
class MockResponse:
    usage: MockUsage


def test_span_context_manager():
    initial_depth = get_current_tree_depth()

    with meter.span("root_task", model="gpt-4o") as root_span:
        assert get_current_tree_depth() == initial_depth + 1
        assert root_span.name == "root_task"

        with meter.span("child_agent", model="claude-3-5-sonnet") as child_span:
            assert get_current_tree_depth() == initial_depth + 2
            assert child_span.parent_span_id == root_span.span_id

            child_span.record_tokens(input_tokens=500, output_tokens=100)

        assert get_current_tree_depth() == initial_depth + 1

    assert get_current_tree_depth() == initial_depth


def test_decorator_sniffs_tokens():
    @meter.trace(name="generate_reply", model="gpt-4o")
    def mock_llm_call():
        return MockResponse(usage=MockUsage(prompt_tokens=250, completion_tokens=80))

    resp = mock_llm_call()
    assert resp.usage.prompt_tokens == 250


@pytest.mark.asyncio
async def test_async_decorator():
    @meter.trace(name="async_generate", model="gpt-4o-mini")
    async def mock_async_llm():
        return {"usage": {"prompt_tokens": 100, "completion_tokens": 50}}

    res = await mock_async_llm()
    assert res["usage"]["prompt_tokens"] == 100


def test_guard_pre_check_circuit_broken():
    class MockBlockedClient(AIMeterClient):
        def check_guard(self, *args, **kwargs):
            return GuardResponse(
                allowed=False,
                decision_code="CIRCUIT_OPEN",
                reason="Monthly budget of $500.00 exceeded",
                circuit_state="OPEN",
                fallback_model="gpt-4o-mini",
            )

    custom_tracer = Tracer(client=MockBlockedClient())

    @custom_tracer.trace(name="risky_agent", model="gpt-4o", pre_check=True)
    def expensive_agent():
        return "completed"

    with pytest.raises(CircuitBreakerOpenError) as exc_info:
        expensive_agent()

    assert exc_info.value.circuit_state == "OPEN"
    assert exc_info.value.fallback_model == "gpt-4o-mini"
    assert "budget" in exc_info.value.reason.lower()
