package workflow

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// CheckpointEntry stores solidified execution result of a completed workflow step.
type CheckpointEntry struct {
	WorkflowID     string    `json:"workflow_id"`
	StepID         string    `json:"step_id"`
	IdempotencyKey string    `json:"idempotency_key"`
	OutputPayload  string    `json:"output_payload"`
	PayloadHash    string    `json:"payload_hash"`
	CostUSD        float64   `json:"cost_usd"`
	InputTokens    int       `json:"input_tokens"`
	OutputTokens   int       `json:"output_tokens"`
	DurationMs     int64     `json:"duration_ms"`
	CreatedAt      time.Time `json:"created_at"`
}

// CheckpointStore manages microsecond-level checkpoint persistence and idempotency verification.
type CheckpointStore struct {
	mu           sync.RWMutex
	byKey        map[string]*CheckpointEntry        // lookup by IdempotencyKey
	byStep       map[string]*CheckpointEntry        // lookup by "wfID:stepID"
	stepToKeyMap map[string]string
}

// NewCheckpointStore initializes an in-memory concurrent checkpoint store.
func NewCheckpointStore() *CheckpointStore {
	return &CheckpointStore{
		byKey:        make(map[string]*CheckpointEntry),
		byStep:       make(map[string]*CheckpointEntry),
		stepToKeyMap: make(map[string]string),
	}
}

// Put records or updates an execution snapshot.
func (cs *CheckpointStore) Put(wfID, stepID, key, payload string, costUSD float64, inTok, outTok int, durMs int64) *CheckpointEntry {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(payload)))
	entry := &CheckpointEntry{
		WorkflowID:     wfID,
		StepID:         stepID,
		IdempotencyKey: key,
		OutputPayload:  payload,
		PayloadHash:    hash,
		CostUSD:        costUSD,
		InputTokens:    inTok,
		OutputTokens:   outTok,
		DurationMs:     durMs,
		CreatedAt:      time.Now().UTC(),
	}

	compositeKey := fmt.Sprintf("%s:%s", wfID, stepID)
	cs.byStep[compositeKey] = entry
	if key != "" {
		cs.byKey[key] = entry
		cs.stepToKeyMap[compositeKey] = key
	}

	return entry
}

// GetByKey retrieves an existing checkpoint by its idempotency key.
func (cs *CheckpointStore) GetByKey(key string) (*CheckpointEntry, bool) {
	if key == "" {
		return nil, false
	}
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	entry, ok := cs.byKey[key]
	return entry, ok
}

// GetByStep retrieves an existing checkpoint by workflow and step ID.
func (cs *CheckpointStore) GetByStep(wfID, stepID string) (*CheckpointEntry, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	compositeKey := fmt.Sprintf("%s:%s", wfID, stepID)
	entry, ok := cs.byStep[compositeKey]
	return entry, ok
}
