import pytest
from aimeter.context import (
    format_w3c_baggage,
    get_current_baggage,
    get_current_trace_id,
    parse_w3c_baggage,
    set_current_trace_id,
    update_current_baggage,
)


def test_trace_id_generation():
    tid1 = get_current_trace_id()
    assert tid1.startswith("tr_")
    tid2 = get_current_trace_id()
    assert tid1 == tid2

    set_current_trace_id("tr_custom_123")
    assert get_current_trace_id() == "tr_custom_123"


def test_w3c_baggage_formatting_and_parsing():
    baggage = {"tenant_id": "org-corp", "workflow_id": "review-agent"}
    formatted = format_w3c_baggage(baggage)
    assert "tenant_id=org-corp" in formatted
    assert "workflow_id=review-agent" in formatted

    parsed = parse_w3c_baggage(formatted)
    assert parsed["tenant_id"] == "org-corp"
    assert parsed["workflow_id"] == "review-agent"


def test_baggage_updates():
    update_current_baggage({"tenant_id": "tenant-a"})
    assert get_current_baggage()["tenant_id"] == "tenant-a"

    update_current_baggage({"workflow_id": "flow-b"})
    b = get_current_baggage()
    assert b["tenant_id"] == "tenant-a"
    assert b["workflow_id"] == "flow-b"
