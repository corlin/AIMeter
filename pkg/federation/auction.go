package federation

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// TaskAuctionHouse manages federated tasks and agent bidding matches
type TaskAuctionHouse struct {
	mu    sync.RWMutex
	tasks map[string]*domain.FederatedTask
}

// NewTaskAuctionHouse initializes an auction house
func NewTaskAuctionHouse() *TaskAuctionHouse {
	return &TaskAuctionHouse{
		tasks: make(map[string]*domain.FederatedTask),
	}
}

// CreateTask registers a new collaborative task
func (h *TaskAuctionHouse) CreateTask(req domain.FederationTaskCreateRequest, voucherID string) *domain.FederatedTask {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now().UTC()
	taskID := fmt.Sprintf("task-fed-%d", now.UnixNano()%10000000)
	tenantID := req.TenantID
	if tenantID == "" {
		tenantID = "default"
	}

	task := &domain.FederatedTask{
		ID:              taskID,
		TenantID:        tenantID,
		Title:           req.Title,
		Description:     req.Description,
		Category:        req.Category,
		SourceWorkspace: req.SourceWorkspace,
		CreatorAgent:    req.CreatorAgent,
		BountyCapUSD:    math.Round(req.BountyCapUSD*10000) / 10000,
		Status:          domain.FederatedTaskOpen,
		VoucherID:       voucherID,
		Bids:            make([]*domain.FederationBid, 0),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	h.tasks[taskID] = task
	cp := *task
	return &cp
}

// GetTask retrieves a task by ID
func (h *TaskAuctionHouse) GetTask(id string) (*domain.FederatedTask, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	task, found := h.tasks[id]
	if !found {
		return nil, false
	}
	cp := *task
	return &cp, true
}

// ListTasks returns tasks matching optional category and status
func (h *TaskAuctionHouse) ListTasks(category, status string) []*domain.FederatedTask {
	h.mu.RLock()
	defer h.mu.RUnlock()

	res := make([]*domain.FederatedTask, 0, len(h.tasks))
	for _, t := range h.tasks {
		if category != "" && t.Category != category {
			continue
		}
		if status != "" && string(t.Status) != status {
			continue
		}
		cp := *t
		res = append(res, &cp)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].CreatedAt.After(res[j].CreatedAt)
	})
	return res
}

// CalculateCompositeScore calculates weighted bid score
func CalculateCompositeScore(bountyCap float64, quoteUSD float64, durationMs int64, reputationScore float64) float64 {
	// 1. Price score (50% weight): Lower price compared to bounty cap gets higher score
	priceRatio := quoteUSD / (bountyCap * 1.2)
	if priceRatio > 1.0 {
		priceRatio = 1.0
	}
	priceScore := (1.0 - priceRatio) * 100.0

	// 2. SLA Duration score (30% weight): Lower latency gets higher score
	durRatio := float64(durationMs) / 20000.0
	if durRatio > 1.0 {
		durRatio = 1.0
	}
	slaScore := (1.0 - durRatio) * 100.0

	// 3. Reputation score (20% weight): 0 to 100
	repScore := reputationScore
	if repScore <= 0 {
		repScore = 80.0
	}

	composite := (0.50 * priceScore) + (0.30 * slaScore) + (0.20 * repScore)
	return math.Round(composite*10) / 10
}

// AddBid adds an agent bid to a task and evaluates winner
func (h *TaskAuctionHouse) AddBid(taskID string, bidReq domain.FederationBidCreateRequest, bidderRep float64) (*domain.FederationBid, *domain.FederatedTask, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	task, found := h.tasks[taskID]
	if !found {
		return nil, nil, fmt.Errorf("task not found: %s", taskID)
	}

	if task.Status == domain.FederatedTaskCompleted || task.Status == domain.FederatedTaskCancelled {
		return nil, nil, fmt.Errorf("task is closed for bidding: %s", task.Status)
	}

	now := time.Now().UTC()
	bidID := fmt.Sprintf("bid-%d", now.UnixNano()%10000000)

	composite := CalculateCompositeScore(task.BountyCapUSD, bidReq.QuotedPriceUSD, bidReq.EstimatedDurationMs, bidderRep)

	bid := &domain.FederationBid{
		ID:                  bidID,
		TaskID:              taskID,
		BidderWorkspace:     bidReq.BidderWorkspace,
		BidderAgent:         bidReq.BidderAgent,
		QuotedPriceUSD:      math.Round(bidReq.QuotedPriceUSD*10000) / 10000,
		EstimatedDurationMs: bidReq.EstimatedDurationMs,
		ReputationScore:     bidderRep,
		CompositeScore:      composite,
		CreatedAt:           now,
	}

	task.Bids = append(task.Bids, bid)
	task.Status = domain.FederatedTaskBidding
	task.UpdatedAt = now

	// Auto-match: Select the bid with highest composite score
	bestBid := bid
	for _, b := range task.Bids {
		if b.CompositeScore > bestBid.CompositeScore {
			bestBid = b
		}
	}

	// Assign winner
	task.AssignedWorkspace = bestBid.BidderWorkspace
	task.AssignedAgent = bestBid.BidderAgent
	task.Status = domain.FederatedTaskInProgress

	taskCopy := *task
	bidCopy := *bid
	return &bidCopy, &taskCopy, nil
}

// SetTaskStatus transitions a task status
func (h *TaskAuctionHouse) SetTaskStatus(taskID string, status domain.FederatedTaskStatus) (*domain.FederatedTask, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	task, found := h.tasks[taskID]
	if !found {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	task.Status = status
	task.UpdatedAt = time.Now().UTC()
	cp := *task
	return &cp, nil
}
