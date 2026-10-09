package cluster

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	alertPkg "github.com/corlin/AIMeter/pkg/alert"
	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/throttler"
)

func setupTestCoordinator(t *testing.T) (*ClusterCoordinator, *throttler.ThrottlerEngine, *budget.BudgetManager) {
	throttlerEng := throttler.NewThrottlerEngine()
	budgetMgr := budget.NewBudgetManager()
	client := &http.Client{Timeout: 2 * time.Second}
	dispatcher := alertPkg.NewAlertDispatcher(client)

	coord := NewClusterCoordinator("hub-test", "us-east-1", true, throttlerEng, budgetMgr, dispatcher)
	if err := coord.LoadSeedTopology("configs/demo/cluster_seed.json"); err != nil {
		t.Fatalf("load seed topology: %v", err)
	}
	return coord, throttlerEng, budgetMgr
}

func TestClusterInitialization(t *testing.T) {
	coord, _, _ := setupTestCoordinator(t)

	nodes := coord.GetNodes()
	if len(nodes) == 0 {
		t.Fatalf("Expected seed nodes loaded, got 0")
	}

	hubFound := false
	for _, n := range nodes {
		if n.Role == "hub" {
			hubFound = true
			break
		}
	}
	if !hubFound {
		t.Errorf("Expected hub node in cluster nodes")
	}

	leases := coord.GetLeases("")
	if len(leases) == 0 {
		t.Errorf("Expected initial quota leases provisioned")
	}

	stats := coord.GetStats()
	if stats.TotalNodes < 1 || stats.TotalRegions < 1 {
		t.Errorf("Unexpected stats: %+v", stats)
	}
}

func TestNodeRegistration(t *testing.T) {
	coord, _, _ := setupTestCoordinator(t)

	newNode := domain.ClusterNode{
		RegionID:    "ap-northeast-1",
		Role:        "spoke",
		ClusterType: "edge-worker",
		Weight:      0.8,
		LatencyMs:   45.2,
	}

	reg := coord.RegisterNode(newNode)
	if reg.NodeID == "" {
		t.Fatalf("Expected auto-generated NodeID")
	}
	if reg.Status != domain.NodeStatusHealthy {
		t.Errorf("Expected healthy status, got %s", reg.Status)
	}

	fetched, err := coord.GetNode(reg.NodeID)
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if fetched.RegionID != "ap-northeast-1" {
		t.Errorf("Expected region ap-northeast-1, got %s", fetched.RegionID)
	}
	if fetched.ActiveLeaseCount == 0 {
		t.Errorf("Expected active leases provisioned for new node")
	}
}

func TestHeartbeatAndTrueUp(t *testing.T) {
	coord, throttlerEng, budgetMgr := setupTestCoordinator(t)

	// Seed a budget rule for tenant-test
	budgetMgr.UpsertBudget(domain.BudgetRule{
		TenantID:        "tenant-test",
		MonthlyLimitUSD: 1000.0,
		CurrentSpendUSD: 100.0,
	})

	hbReq := domain.NodeHeartbeatRequest{
		NodeID:   "spoke-eu-central",
		RegionID: "eu-central-1",
		ReportedUsage: map[string]domain.LeaseUsageDelta{
			"tenant-test": {
				Requests: 15,
				Tokens:   3500,
				CostUSD:  0.08,
			},
		},
		LatencyMs: 82.5,
	}

	resp := coord.Heartbeat(hbReq)
	if resp.Status != "ack" {
		t.Fatalf("Expected ack status, got %s", resp.Status)
	}
	if len(resp.GrantedLeases) == 0 {
		t.Errorf("Expected granted leases in heartbeat response")
	}
	if resp.NextHeartbeatIntervalMs != 5000 {
		t.Errorf("Expected 5000ms next heartbeat interval, got %d", resp.NextHeartbeatIntervalMs)
	}

	// Verify TrueUp touched budget
	rules := budgetMgr.GetBudgets("tenant-test")
	if len(rules) > 0 && rules[0].CurrentSpendUSD < 100.08 {
		t.Errorf("Expected budget updated with reported cost, got %v", rules[0].CurrentSpendUSD)
	}

	_ = throttlerEng
}

func TestRebalanceLeases(t *testing.T) {
	coord, _, _ := setupTestCoordinator(t)

	coord.RebalanceLeases()
	leases := coord.GetLeases("")
	if len(leases) == 0 {
		t.Fatalf("Expected leases to exist after rebalance")
	}
	for _, l := range leases {
		if l.AllocatedRPM <= 0 || l.AllocatedTPM <= 0 {
			t.Errorf("Expected positive quota after rebalance, got %+v", l)
		}
	}
}

func TestLivenessCheck(t *testing.T) {
	coord, _, _ := setupTestCoordinator(t)
	ctx := context.Background()

	// Register a spoke node with old heartbeat time
	oldNode := domain.ClusterNode{
		NodeID:          "stale-spoke-1",
		RegionID:        "sa-east-1",
		Role:            "spoke",
		Status:          domain.NodeStatusHealthy,
		LastHeartbeatAt: time.Now().UTC().Add(-20 * time.Second),
	}
	coord.RegisterNode(oldNode)

	// Explicitly set LastHeartbeatAt to stale
	coord.mu.Lock()
	coord.nodes["stale-spoke-1"].LastHeartbeatAt = time.Now().UTC().Add(-20 * time.Second)
	coord.mu.Unlock()

	coord.checkNodeLiveness(ctx)

	stale, err := coord.GetNode("stale-spoke-1")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if stale.Status != domain.NodeStatusDegraded {
		t.Errorf("Expected status degraded for stale node, got %s", stale.Status)
	}
}

func TestClusterSimulate(t *testing.T) {
	coord, _, _ := setupTestCoordinator(t)

	req := domain.ClusterSimulateRequest{
		PartitionedRegionID:       "eu-central-1",
		TrafficSurgeMultiplier:    2.5,
		SimulateDurationSec:       30,
		EnableFailSafeDegradation: true,
	}

	resp := coord.Simulate(req)
	if resp.TargetRegion != "eu-central-1" {
		t.Errorf("Expected target region eu-central-1, got %s", resp.TargetRegion)
	}
	if len(resp.Timeline) == 0 {
		t.Errorf("Expected timeline steps")
	}
	if resp.AutonomousSpendAllowedUSD <= 0 || resp.RunawayOverSpendBlockedUSD <= 0 {
		t.Errorf("Expected non-zero spend protection metrics, got %+v", resp)
	}
	if resp.Analysis == "" {
		t.Errorf("Expected non-empty analysis string")
	}
}

func TestClusterConcurrencyAndRace(t *testing.T) {
	coord, _, _ := setupTestCoordinator(t)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			nodeID := "spoke-eu-central"
			if idx%2 == 0 {
				nodeID = "spoke-ap-southeast"
			}
			coord.Heartbeat(domain.NodeHeartbeatRequest{
				NodeID:   nodeID,
				RegionID: "eu-central-1",
				ReportedUsage: map[string]domain.LeaseUsageDelta{
					"default": {Requests: idx, Tokens: idx * 100, CostUSD: float64(idx) * 0.001},
				},
				LatencyMs: float64(10 + idx),
			})
			_ = coord.GetNodes()
			_ = coord.GetLeases("default")
			_ = coord.GetStats()
			coord.RebalanceLeases()
			_ = coord.Simulate(domain.ClusterSimulateRequest{
				PartitionedRegionID: "ap-southeast-1",
			})
		}(i)
	}
	wg.Wait()
}
