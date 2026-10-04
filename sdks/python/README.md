# AI Meter Python SDK

Official Python Client & Instrumentation SDK for [AI Meter](https://github.com/corlin/AIMeter) — The Enterprise AI Usage & Cost Control Plane.

## Features

- **⚡ Active Guard Pre-Check**: Fast synchronous pre-check (`<2ms` backend target, default 20ms client timeout with fail-open safety) preventing runaway loops and budget overruns.
- **🔄 Non-blocking Usage Reporting**: Background daemon worker thread with memory queue, batching and auto-flush on exit.
- **🎯 Context & Tree Tracing**: Automatic DAG tree depth tracking, parent/child span linking via Python `contextvars`.
- **🔍 Auto Token Sniffing**: Automatically extracts token counts from OpenAI, Anthropic, and other standard completion responses.
- **🧩 Framework Integrations**: Native `LangChain` and `LlamaIndex` callback handlers.

## Installation

```bash
# Core SDK
pip install aimeter

# With LangChain integration
pip install "aimeter[langchain]"

# With LlamaIndex integration
pip install "aimeter[llamaindex]"
```

## Quick Start

```python
from aimeter import AIMeter, meter

# 1. Initialize client
client = AIMeter(base_url="http://localhost:8080", tenant_id="org-enterprise-1")

# 2. Use @meter.trace decorator
@meter.trace(name="contract_review", model="gpt-4o", pre_check=True)
def run_review(prompt: str):
    # Call OpenAI / Anthropic
    response = openai_client.chat.completions.create(
        model="gpt-4o",
        messages=[{"role": "user", "content": prompt}]
    )
    return response

# 3. Use with context manager
with meter.span(name="search_tool", provider="tavily", model="search") as span:
    span.record_tokens(input_tokens=150, output_tokens=0)
```
