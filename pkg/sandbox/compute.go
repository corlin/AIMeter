package sandbox

import (
	"math"

	"github.com/corlin/AIMeter/pkg/domain"
)

// Default rates for ephemeral sandboxes (vCPU-sec and RAM-GB-sec)
const (
	DefaultRatePerCPUSec   = 0.000015 // $0.000015 / vCPU-sec (~$0.054 / vCPU-hour)
	DefaultRatePerRAMGBSec = 0.000004 // $0.000004 / GB-sec   (~$0.0144 / GB-hour)
	DefaultColdStartBase   = 0.0010   // $0.0010 base spin-up fee
	DefaultTimeoutSeconds  = 60       // 60s hard ceiling
)

// DefaultSpec returns a validated computation hardware specification
func DefaultSpec(runtime domain.SandboxRuntime, cpu, ramMB, timeoutSec int) domain.SandboxComputeSpec {
	if cpu <= 0 {
		cpu = 2
	}
	if ramMB <= 0 {
		ramMB = 2048
	}
	if timeoutSec <= 0 {
		timeoutSec = DefaultTimeoutSeconds
	}

	coldStart := DefaultColdStartBase
	if runtime == domain.SandboxRuntimeWasm {
		coldStart = 0.0002 // lighter cold start for WebAssembly
	} else if runtime == domain.SandboxRuntimeFirecracker {
		coldStart = 0.0005
	}

	return domain.SandboxComputeSpec{
		CPU:              cpu,
		RAMMB:            ramMB,
		ColdStartBaseUSD: coldStart,
		RatePerCPUSec:    DefaultRatePerCPUSec,
		RatePerRAMGBSec:  DefaultRatePerRAMGBSec,
		TimeoutSec:       timeoutSec,
	}
}

// CalculateComputeCost calculates exact micro-VM ephemeral runtime cost and detects timeout cut-offs
func CalculateComputeCost(spec domain.SandboxComputeSpec, durationMs int64) (float64, bool) {
	if durationMs <= 0 {
		return spec.ColdStartBaseUSD, false
	}

	isTimeoutCapped := false
	maxMs := int64(spec.TimeoutSec * 1000)
	effectiveMs := durationMs
	if durationMs > maxMs {
		effectiveMs = maxMs
		isTimeoutCapped = true
	}

	durSec := float64(effectiveMs) / 1000.0
	cpuCost := float64(spec.CPU) * durSec * spec.RatePerCPUSec
	ramGBCost := (float64(spec.RAMMB) / 1024.0) * durSec * spec.RatePerRAMGBSec
	total := spec.ColdStartBaseUSD + cpuCost + ramGBCost

	rounded := math.Round(total*1000000) / 1000000
	return rounded, isTimeoutCapped
}
