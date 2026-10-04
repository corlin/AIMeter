from dataclasses import dataclass
from typing import Any, Dict
import uuid
import pytest

from aimeter.adapters import LangChainCallbackHandler, LlamaIndexCallbackHandler
from aimeter.client import AIMeterClient, GuardResponse
from aimeter.exceptions import CircuitBreakerOpenError


@dataclass
class MockLLMResult:
    llm_output: Dict[str, Any]
    generations: Any = None


def test_langchain_callback_lifecycle():
    client = AIMeterClient(base_url="http://127.0.0.1:59999", guard_timeout_seconds=0.01)
    handler = LangChainCallbackHandler(client=client, workflow_id="flow-langchain", pre_check=False)

    run_id = uuid.uuid4()

    # 1. Start LLM
    handler.on_llm_start(
        serialized={"name": "ChatOpenAI"},
        prompts=["Summarize this document"],
        run_id=run_id,
        invocation_params={"model_name": "gpt-4o", "_type": "openai"},
    )
    assert str(run_id) in handler._active_spans

    # 2. End LLM with tokens
    mock_result = MockLLMResult(
        llm_output={
            "token_usage": {
                "prompt_tokens": 500,
                "completion_tokens": 120,
                "cached_tokens": 200,
            }
        }
    )
    handler.on_llm_end(mock_result, run_id=run_id)
    assert str(run_id) not in handler._active_spans

    client.close()


def test_langchain_callback_pre_check_rejection():
    class BlockedClient(AIMeterClient):
        def check_guard(self, *args, **kwargs):
            return GuardResponse(
                allowed=False,
                decision_code="RUNAWAY_LOOP_PREVENTED",
                reason="DAG tree depth exceeded limit 12",
                circuit_state="OPEN",
            )

    handler = LangChainCallbackHandler(
        client=BlockedClient(),
        workflow_id="loop-flow",
        pre_check=True,
    )

    with pytest.raises(CircuitBreakerOpenError) as exc_info:
        handler.on_llm_start(
            serialized={"name": "ChatOpenAI"},
            prompts=["Looping prompt"],
            run_id=uuid.uuid4(),
            invocation_params={"model_name": "gpt-4o"},
        )

    assert exc_info.value.decision_code == "RUNAWAY_LOOP_PREVENTED"


def test_llamaindex_callback_lifecycle():
    client = AIMeterClient(base_url="http://127.0.0.1:59999", guard_timeout_seconds=0.01)
    handler = LlamaIndexCallbackHandler(client=client, workflow_id="flow-llama", pre_check=False)

    event_id = "event_123"
    handler.on_event_start(
        event_type="llm",
        payload={"serialized": {"model_name": "gpt-4o"}},
        event_id=event_id,
    )
    assert event_id in handler._active_spans

    handler.on_event_end(
        event_type="llm",
        payload={"response": {"usage": {"prompt_tokens": 300, "completion_tokens": 50}}},
        event_id=event_id,
    )
    assert event_id not in handler._active_spans

    client.close()
