"""AI Meter Quickstart Example."""

import time
from aimeter import AIMeter, CircuitBreakerOpenError, meter

# 1. Initialize client pointing to AI Meter Control Plane
client = AIMeter(
    base_url="http://localhost:8080",
    tenant_id="org-enterprise-1",
    guard_timeout_seconds=0.02,  # 20ms fail-open timeout
)


# 2. Trace complex agent workflow
@meter.trace(name="contract_analyzer", model="gpt-4o", pre_check=True)
def analyze_contract(document_text: str):
    print(f"[Agent] Analyzing document of length {len(document_text)}...")

    # Sub-task 1: Search legal clauses using Tavily
    with meter.span(name="sec_edgar_search", provider="tavily", model="search") as search_span:
        time.sleep(0.05)
        # Record 2 search queries
        search_span.record_tokens(input_tokens=2, output_tokens=0)

    # Sub-task 2: Reasoning & Extraction using DeepSeek-R1
    with meter.span(name="clause_extraction", provider="deepseek", model="deepseek-reasoner") as llm_span:
        time.sleep(0.1)
        # DeepSeek R1 reasoning tokens
        llm_span.record_tokens(
            input_tokens=1500,
            output_tokens=300,
            reasoning_tokens=800,
        )

    return {"status": "SUCCESS", "risks_found": 3}


if __name__ == "__main__":
    try:
        result = analyze_contract("Standard Mutual Non-Disclosure Agreement ...")
        print("[Result]", result)
    except CircuitBreakerOpenError as e:
        print(f"[BLOCKED] Circuit Breaker Active: {e.reason}")
        if e.fallback_model:
            print(f"[FALLBACK] Recommended alternative model: {e.fallback_model}")
    finally:
        # Gracefully flush events to AI Meter Control Plane
        client.flush()
