import pytest
from aimeter import AIMeterClient


def test_client_init_and_fail_open():
    # Use dummy port to ensure unreachable network
    client = AIMeterClient(
        base_url="http://127.0.0.1:59999",
        tenant_id="test-tenant",
        guard_timeout_seconds=0.05,
    )

    # Active Guard check must Fail-Open without crashing
    resp = client.check_guard(
        workflow_id="test-flow",
        model="gpt-4o",
        tree_depth=2,
    )

    assert resp.allowed is True
    assert resp.decision_code == "FAIL_OPEN"
    assert resp.circuit_state == "CLOSED"

    client.close()


@pytest.mark.asyncio
async def test_client_async_fail_open():
    client = AIMeterClient(
        base_url="http://127.0.0.1:59999",
        tenant_id="test-tenant",
        guard_timeout_seconds=0.05,
    )

    resp = await client.check_guard_async(
        workflow_id="test-flow",
        model="gpt-4o",
        tree_depth=1,
    )

    assert resp.allowed is True
    assert resp.decision_code == "FAIL_OPEN"

    client.close()


def test_record_usage_enqueuing():
    client = AIMeterClient(
        base_url="http://127.0.0.1:59999",
        tenant_id="test-tenant",
        batch_size=10,
    )

    # Enqueue a usage event
    client.record_usage(
        trace_id="tr_123",
        span_id="sp_456",
        provider="openai",
        model="gpt-4o",
        latency_ms=250,
        input_tokens=1000,
        output_tokens=200,
        cached_tokens=500,
        reasoning_tokens=50,
    )

    # Flush should drain queue safely
    client.flush(timeout=0.5)
    client.close()
