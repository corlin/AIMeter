"use client";

import { useState, useEffect } from "react";
import { fetchCircuitBreakers, resetCircuitBreaker, checkGuard, fetchTenants } from "@/lib/api";
import { CircuitBreakerRecord, GuardCheckResponse, Tenant } from "@/types";
import { 
  ZapOff, 
  ShieldCheck, 
  AlertTriangle, 
  RefreshCw, 
  RotateCcw, 
  CheckCircle2, 
  XCircle,
  Play,
  Flame,
  Layers,
  Clock
} from "lucide-react";

export default function CircuitBreakerPage() {
  const [breakers, setBreakers] = useState<CircuitBreakerRecord[]>([]);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [selectedTenant, setSelectedTenant] = useState("all");
  const [loading, setLoading] = useState(true);
  const [resettingKey, setResettingKey] = useState<string | null>(null);

  // Playground state
  const [testTenant, setTestTenant] = useState("org-enterprise-1");
  const [testWorkflow, setTestWorkflow] = useState("contract-review-agent");
  const [testModel, setTestModel] = useState("gpt-4o");
  const [testDepth, setTestDepth] = useState(3);
  const [testTesting, setTestTesting] = useState(false);
  const [testResult, setTestResult] = useState<GuardCheckResponse | null>(null);

  const loadData = async () => {
    setLoading(true);
    try {
      const [bData, tData] = await Promise.all([
        fetchCircuitBreakers(selectedTenant),
        fetchTenants(),
      ]);
      setBreakers(bData);
      setTenants(tData);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedTenant]);

  const handleReset = async (tenantId: string, workflowId: string) => {
    const key = `${tenantId}:${workflowId}`;
    setResettingKey(key);
    try {
      await resetCircuitBreaker(tenantId, workflowId);
      await loadData();
    } catch (err) {
      console.error("Failed to reset breaker", err);
    } finally {
      setResettingKey(null);
    }
  };

  const handleRunPlayground = async () => {
    setTestTesting(true);
    try {
      const res = await checkGuard({
        tenant_id: testTenant,
        workflow_id: testWorkflow,
        model: testModel,
        current_tree_depth: Number(testDepth),
      });
      setTestResult(res);
      await loadData();
    } catch (err) {
      console.error("Playground test failed", err);
    } finally {
      setTestTesting(false);
    }
  };

  const openBreakers = breakers.filter((b) => b.state === "OPEN").length;
  const halfOpenBreakers = breakers.filter((b) => b.state === "HALF_OPEN").length;
  const totalBlocked = breakers.reduce((sum, b) => sum + b.blocked_count, 0);

  const getStateBadge = (state: string) => {
    switch (state) {
      case "OPEN":
        return (
          <span className="flex items-center gap-1 text-[11px] font-semibold px-2.5 py-0.5 rounded-full bg-rose-500/10 text-rose-400 border border-rose-500/30 animate-pulse">
            <Flame className="h-3 w-3" />
            <span>OPEN (Blocked)</span>
          </span>
        );
      case "HALF_OPEN":
        return (
          <span className="flex items-center gap-1 text-[11px] font-semibold px-2.5 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/30">
            <AlertTriangle className="h-3 w-3" />
            <span>HALF-OPEN (Canary)</span>
          </span>
        );
      default:
        return (
          <span className="flex items-center gap-1 text-[11px] font-semibold px-2.5 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
            <CheckCircle2 className="h-3 w-3" />
            <span>CLOSED (Normal)</span>
          </span>
        );
    }
  };

  return (
    <div className="space-y-8">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
            <ZapOff className="h-6 w-6 text-rose-400" />
            Active Guard & Circuit Breaker Control Center
          </h1>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Real-time invocation pre-check (<span className="font-mono text-emerald-400">&lt;2ms</span>), budget ceiling enforcement, and runaway loop circuit breakers.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <select
            value={selectedTenant}
            onChange={(e) => setSelectedTenant(e.target.value)}
            className="bg-zinc-900 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-zinc-700 font-medium"
          >
            <option value="all">All Tenants / Orgs</option>
            {tenants.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name} ({t.id})
              </option>
            ))}
          </select>

          <button
            onClick={loadData}
            className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-zinc-900 border border-zinc-800 text-xs font-medium text-zinc-300 hover:text-white hover:bg-zinc-800 transition-colors"
          >
            <RefreshCw className="h-4 w-4" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* KPI Overview Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
          <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">Tripped Breakers (OPEN)</span>
          <span className="text-2xl font-bold font-mono text-rose-400 mt-1 block">{openBreakers}</span>
          <span className="text-[11px] text-zinc-500 mt-0.5 block">Workflows actively blocked</span>
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
          <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">Half-Open Canaries</span>
          <span className="text-2xl font-bold font-mono text-amber-400 mt-1 block">{halfOpenBreakers}</span>
          <span className="text-[11px] text-zinc-500 mt-0.5 block">Testing probe requests</span>
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
          <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">Invocations Blocked</span>
          <span className="text-2xl font-bold font-mono text-white mt-1 block">{totalBlocked}</span>
          <span className="text-[11px] text-zinc-500 mt-0.5 block">Prevented runaway waste</span>
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm flex flex-col justify-between">
          <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">Fail-Open Protection</span>
          <div className="flex items-center gap-1.5 mt-2">
            <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" />
            <span className="text-sm font-semibold text-emerald-400">Fail-Open Active</span>
          </div>
          <span className="text-[11px] text-zinc-500 mt-1 block">Zero business outage risk</span>
        </div>
      </div>

      {/* Interactive Guard Pre-Check Playground Sandbox */}
      <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 backdrop-blur-sm space-y-4">
        <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
          <div>
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              <Play className="h-4 w-4 text-emerald-400" />
              Active Guard Pre-Check Simulator (POST /v1/guard/check)
            </h3>
            <p className="text-xs text-zinc-400 mt-0.5">
              Simulate an invocation from an AI Gateway or SDK to test real-time admission and circuit breaker decisions.
            </p>
          </div>
          <span className="text-xs font-mono text-zinc-500">Latency: &lt; 2ms</span>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
          <div>
            <label className="block text-xs text-zinc-400 mb-1 font-medium">Tenant ID</label>
            <input
              type="text"
              value={testTenant}
              onChange={(e) => setTestTenant(e.target.value)}
              className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-zinc-700 font-mono"
            />
          </div>

          <div>
            <label className="block text-xs text-zinc-400 mb-1 font-medium">Workflow ID</label>
            <input
              type="text"
              value={testWorkflow}
              onChange={(e) => setTestWorkflow(e.target.value)}
              className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-zinc-700 font-mono"
            />
          </div>

          <div>
            <label className="block text-xs text-zinc-400 mb-1 font-medium">Target Model</label>
            <select
              value={testModel}
              onChange={(e) => setTestModel(e.target.value)}
              className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-zinc-700"
            >
              <option value="gpt-4o">GPT-4o</option>
              <option value="claude-3-5-sonnet">Claude 3.5 Sonnet</option>
              <option value="deepseek-reasoner">DeepSeek Reasoner (R1)</option>
              <option value="gpt-4o-mini">GPT-4o-mini</option>
            </select>
          </div>

          <div>
            <label className="block text-xs text-zinc-400 mb-1 font-medium">Execution Depth (Limit: 12)</label>
            <div className="flex gap-2">
              <input
                type="number"
                min="1"
                max="25"
                value={testDepth}
                onChange={(e) => setTestDepth(Number(e.target.value))}
                className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-zinc-700 font-mono"
              />
              <button
                onClick={handleRunPlayground}
                disabled={testTesting}
                className="px-4 py-2 rounded-lg bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 text-xs font-semibold whitespace-nowrap transition-colors"
              >
                {testTesting ? "Checking..." : "Execute Check"}
              </button>
            </div>
          </div>
        </div>

        {testResult && (
          <div className={`p-4 rounded-lg border text-xs space-y-2 ${
            testResult.allowed 
              ? "bg-emerald-500/10 border-emerald-500/20 text-emerald-300"
              : "bg-rose-500/10 border-rose-500/20 text-rose-300"
          }`}>
            <div className="flex items-center justify-between">
              <span className="font-bold flex items-center gap-1.5 text-sm">
                {testResult.allowed ? <CheckCircle2 className="h-4 w-4" /> : <XCircle className="h-4 w-4" />}
                Decision: {testResult.decision_code} ({testResult.allowed ? "ALLOWED" : "BLOCKED"})
              </span>
              <span className="font-mono text-[11px] text-zinc-400">
                Circuit: {testResult.circuit_state}
              </span>
            </div>
            <p className="font-mono text-[11px]">{testResult.reason}</p>
            {testResult.fallback_model && (
              <div className="pt-2 border-t border-rose-500/20 flex items-center gap-2">
                <span className="font-semibold">Recommended Fallback Model:</span>
                <code className="px-2 py-0.5 rounded bg-zinc-900 border border-zinc-800 text-amber-400 font-mono">
                  {testResult.fallback_model}
                </code>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Circuit Breaker Instances Table */}
      <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 backdrop-blur-sm space-y-4">
        <div className="flex items-center justify-between pb-2 border-b border-zinc-800">
          <div>
            <h3 className="text-sm font-semibold text-white">Registered Circuit Breaker Instances</h3>
            <p className="text-xs text-zinc-400">Active protection state machines per tenant and workflow.</p>
          </div>
          <span className="text-xs font-mono text-zinc-400">
            {breakers.length} Breaker Instances
          </span>
        </div>

        {breakers.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead>
                <tr className="border-b border-zinc-800 text-zinc-400 uppercase tracking-wider font-semibold">
                  <th className="pb-3">Tenant & Workflow</th>
                  <th className="pb-3">Circuit State</th>
                  <th className="pb-3">Blocked Count</th>
                  <th className="pb-3">Reason / Trigger</th>
                  <th className="pb-3">Cooldown</th>
                  <th className="pb-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800/60 font-mono">
                {breakers.map((b) => (
                  <tr key={b.key} className="hover:bg-zinc-800/30">
                    <td className="py-3.5 font-sans">
                      <div className="font-bold text-white">{b.workflow_id}</div>
                      <div className="text-[11px] text-zinc-500 font-mono">{b.tenant_id}</div>
                    </td>
                    <td className="py-3.5">{getStateBadge(b.state)}</td>
                    <td className="py-3.5 font-bold text-zinc-200">
                      {b.blocked_count} calls
                    </td>
                    <td className="py-3.5 text-zinc-300 max-w-md font-sans text-xs">
                      {b.reason || "Operating within safety limits"}
                    </td>
                    <td className="py-3.5 text-zinc-400 text-xs">
                      {b.cooldown_seconds}s
                    </td>
                    <td className="py-3.5 text-right font-sans">
                      {b.state === "OPEN" || b.state === "HALF_OPEN" ? (
                        <button
                          onClick={() => handleReset(b.tenant_id, b.workflow_id)}
                          disabled={resettingKey === b.key}
                          className="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 text-xs font-medium transition-colors"
                        >
                          <RotateCcw className="h-3 w-3" />
                          <span>{resettingKey === b.key ? "Resetting..." : "Reset / Unblock"}</span>
                        </button>
                      ) : (
                        <span className="text-zinc-600 text-xs font-mono">Healthy</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="rounded-xl border border-dashed border-zinc-800 bg-zinc-900/30 p-12 text-center space-y-3">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-400">
              <ShieldCheck className="h-6 w-6" />
            </div>
            <h3 className="text-sm font-semibold text-white">All Workflows Protected and Normal</h3>
            <p className="text-xs text-zinc-400 max-w-sm mx-auto">
              No circuit breakers currently tripped. As invocations occur, runaway loops or budget excesses will be blocked automatically.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
