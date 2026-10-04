import http from 'k6/http';
import { check, sleep } from 'k6';

// k6 Stress Testing Options for AI Meter Production Gateway
export const options = {
  scenarios: {
    // Scenario 1: High frequency Active Guard synchronous pre-check
    guard_pre_check: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '15s', target: 50 },  // Ramp up
        { duration: '30s', target: 200 }, // Peak load
        { duration: '15s', target: 0 },   // Ramp down
      ],
      gracefulRampDown: '5s',
      exec: 'testGuardCheck',
    },
    // Scenario 2: Smart LLM Reverse Proxy invocations
    proxy_chat: {
      executor: 'constant-vus',
      vus: 30,
      duration: '45s',
      startTime: '10s',
      exec: 'testProxyChat',
    },
    // Scenario 3: Asynchronous AI Gateway Telemetry ingest
    telemetry_ingest: {
      executor: 'constant-vus',
      vus: 50,
      duration: '45s',
      startTime: '5s',
      exec: 'testTelemetryIngest',
    },
  },
  thresholds: {
    // SLA Gates
    'http_req_duration{scenario:guard_pre_check}': ['p(95)<2.0', 'p(99)<4.0'],
    'http_req_duration{scenario:proxy_chat}': ['p(95)<8.0', 'p(99)<25.0'],
    'http_req_duration{scenario:telemetry_ingest}': ['p(95)<2.0', 'p(99)<5.0'],
    'http_req_failed': ['rate<0.005'], // Error rate < 0.5%
  },
};

const BASE_URL = __ENV.AIMETER_URL || 'http://localhost:8080';
const API_KEY = __ENV.AIMETER_KEY || 'sk-aimeter-live-benchmark-demo';

const defaultHeaders = {
  'Content-Type': 'application/json',
  'Authorization': `Bearer ${API_KEY}`,
  'X-Tenant-ID': 'org-enterprise-1',
};

// 1. Guard Check
export function testGuardCheck() {
  const payload = JSON.stringify({
    tenant_id: 'org-enterprise-1',
    workflow_id: 'contract-review-agent',
    model: 'gpt-4o',
    current_tree_depth: 3,
  });

  const res = http.post(`${BASE_URL}/v1/guard/check`, payload, { headers: defaultHeaders });
  check(res, {
    'status is 200': (r) => r.status === 200,
    'guard allowed': (r) => {
      try {
        return JSON.parse(r.body).allowed === true;
      } catch (e) {
        return false;
      }
    },
  });
}

// 2. Proxy Chat Completions
export function testProxyChat() {
  const payload = JSON.stringify({
    model: 'gpt-4o',
    messages: [{ role: 'user', content: 'Benchmark probe message' }],
  });

  const res = http.post(`${BASE_URL}/v1/chat/completions`, payload, {
    headers: Object.assign({}, defaultHeaders, {
      'X-Workflow-ID': 'benchmark-flow',
    }),
  });

  check(res, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
  });
}

// 3. Telemetry Ingest
export function testTelemetryIngest() {
  const payload = JSON.stringify({
    provider: 'openai',
    model: 'gpt-4o',
    prompt_tokens: 1500,
    completion_tokens: 400,
    cached_tokens: 800,
    latency_ms: 380,
  });

  const res = http.post(`${BASE_URL}/v1/gateway/litellm`, payload, { headers: defaultHeaders });
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}
