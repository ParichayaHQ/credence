package consensus

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultCheckpointor implements the Checkpointor interface
type DefaultCheckpointor struct {
	config       *CheckpointorConfig
	store        CheckpointStore
	logClient    LogNodeClient
	publisher    P2PPublisher
	blsAgg       BLSAggregator
	vrfProvider  VRFProvider
	
	// Internal state
	mu           sync.RWMutex
	running      bool
	currentTask  *SigningTask
	committee    *Committee
	
	// Channels for coordination
	stopCh       chan struct{}
	taskCh       chan *SigningTask
	partialCh    chan *PartialSignature
}

// NewCheckpointor creates a new checkpointor service
func NewCheckpointor(
	config *CheckpointorConfig,
	store CheckpointStore,
	logClient LogNodeClient,
	publisher P2PPublisher,
	blsAgg BLSAggregator,
	vrfProvider VRFProvider,
) *DefaultCheckpointor {
	return &DefaultCheckpointor{
		config:      config,
		store:       store,
		logClient:   logClient,
		publisher:   publisher,
		blsAgg:      blsAgg,
		vrfProvider: vrfProvider,
		stopCh:      make(chan struct{}),
		taskCh:      make(chan *SigningTask, 10),
		partialCh:   make(chan *PartialSignature, 100),
	}
}

// Start begins the checkpointor service
func (c *DefaultCheckpointor) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.running {
		return fmt.Errorf("checkpointor already running")
	}
	
	// Load current committee from storage
	committee, err := c.store.GetCurrentCommittee(ctx)
	if err != nil {
		// Initialize with empty committee if none exists
		committee = &Committee{
			Members:   []string{},
			Threshold: c.config.SignatureThreshold,
			Epoch:     0,
			StartTime: time.Now(),
			EndTime:   time.Now().Add(c.config.RotationPeriod),
		}
		if err := c.store.StoreCommittee(ctx, committee); err != nil {
			return fmt.Errorf("failed to store initial committee: %w", err)
		}
	}
	c.committee = committee
	
	c.running = true
	
	// Start background goroutines
	go c.checkpointScheduler(ctx)
	go c.partialSignatureProcessor(ctx)
	
	return nil
}

// Stop stops the checkpointor service
func (c *DefaultCheckpointor) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.running {
		return nil
	}
	
	c.running = false
	close(c.stopCh)
	
	return nil
}

// GetLatestCheckpoint returns the latest checkpoint
func (c *DefaultCheckpointor) GetLatestCheckpoint(ctx context.Context) (*Checkpoint, error) {
	return c.store.GetLatestCheckpoint(ctx)
}

// GetCheckpoint returns a checkpoint by epoch
func (c *DefaultCheckpointor) GetCheckpoint(ctx context.Context, epoch int64) (*Checkpoint, error) {
	return c.store.GetCheckpoint(ctx, epoch)
}

// SubmitPartialSignature submits a partial signature from a committee member
func (c *DefaultCheckpointor) SubmitPartialSignature(ctx context.Context, partial *PartialSignature) error {
	// Verify the partial signature is from a committee member
	c.mu.RLock()
	committee := c.committee
	c.mu.RUnlock()
	
	if !contains(committee.Members, partial.SignerDID) {
		return fmt.Errorf("signer %s is not a committee member", partial.SignerDID)
	}
	
	// Send to processing channel
	select {
	case c.partialCh <- partial:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetCurrentTask returns the current signing task
func (c *DefaultCheckpointor) GetCurrentTask(ctx context.Context) (*SigningTask, error) {
	c.mu.RLock()
	task := c.currentTask
	c.mu.RUnlock()
	
	if task == nil {
		return nil, fmt.Errorf("no current signing task")
	}
	
	return task, nil
}

// ForceCheckpoint forces creation of a checkpoint (admin/testing)
func (c *DefaultCheckpointor) ForceCheckpoint(ctx context.Context) (*Checkpoint, error) {
	return c.createCheckpointFromSTH(ctx)
}

// checkpointScheduler runs the periodic checkpoint creation
func (c *DefaultCheckpointor) checkpointScheduler(ctx context.Context) {
	ticker := time.NewTicker(c.config.CheckpointInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			checkpoint, err := c.createCheckpointFromSTH(ctx)
			if err != nil {
				// Log error but continue
				fmt.Printf("Failed to create checkpoint: %v\n", err)
				continue
			}
			
			if checkpoint != nil {
				// Publish to p2p network
				if err := c.publisher.PublishCheckpoint(ctx, checkpoint); err != nil {
					fmt.Printf("Failed to publish checkpoint: %v\n", err)
				}
			}
			
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// partialSignatureProcessor handles incoming partial signatures
func (c *DefaultCheckpointor) partialSignatureProcessor(ctx context.Context) {
	for {
		select {
		case partial := <-c.partialCh:
			if err := c.processPartialSignature(ctx, partial); err != nil {
				fmt.Printf("Failed to process partial signature: %v\n", err)
			}
			
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// createCheckpointFromSTH creates a checkpoint from the latest STH
func (c *DefaultCheckpointor) createCheckpointFromSTH(ctx context.Context) (*Checkpoint, error) {
	// Get latest STH from log node
	sth, err := c.logClient.GetLatestSTH(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get STH: %w", err)
	}
	
	// Check if we already have a checkpoint for this tree size
	latest, err := c.store.GetLatestCheckpoint(ctx)
	if err == nil && latest != nil && latest.TreeSize >= sth.TreeSize {
		return nil, nil // No new checkpoint needed
	}
	
	// Create signing task
	epoch := time.Now().Unix() / int64(c.config.CheckpointInterval.Seconds())
	task := &SigningTask{
		Root:     sth.RootHash,
		Epoch:    epoch,
		TreeSize: sth.TreeSize,
		Deadline: time.Now().Add(c.config.SignatureTimeout),
		Partials: []PartialSignature{},
		Status:   TaskStatusPending,
	}
	
	// Store and set as current task
	c.mu.Lock()
	c.currentTask = task
	c.mu.Unlock()
	
	if err := c.store.StoreSigningTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to store signing task: %w", err)
	}
	
	// Start collecting signatures
	task.Status = TaskStatusCollecting
	
	// Wait for threshold signatures with timeout
	deadline := time.NewTimer(c.config.SignatureTimeout)
	defer deadline.Stop()
	
	for {
		select {
		case <-deadline.C:
			task.Status = TaskStatusFailed
			return nil, fmt.Errorf("signature collection timeout")
			
		case <-time.After(1 * time.Second):
			// Check if we have enough signatures
			c.mu.RLock()
			partialCount := len(c.currentTask.Partials)
			threshold := c.committee.Threshold
			c.mu.RUnlock()
			
			if partialCount >= threshold {
				return c.finalizeCheckpoint(ctx, task)
			}
			
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// processPartialSignature processes an incoming partial signature
func (c *DefaultCheckpointor) processPartialSignature(ctx context.Context, partial *PartialSignature) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.currentTask == nil || c.currentTask.Status != TaskStatusCollecting {
		return fmt.Errorf("no active signing task")
	}
	
	// Verify partial is for current task
	if !bytesEqual(c.currentTask.Root, partial.Root) || c.currentTask.Epoch != partial.Epoch {
		return fmt.Errorf("partial signature not for current task")
	}
	
	// Check if we already have signature from this member
	for _, existing := range c.currentTask.Partials {
		if existing.SignerDID == partial.SignerDID {
			return fmt.Errorf("already have signature from %s", partial.SignerDID)
		}
	}
	
	// Add to current task
	c.currentTask.Partials = append(c.currentTask.Partials, *partial)
	
	return nil
}

// finalizeCheckpoint creates the final checkpoint with aggregated signature
func (c *DefaultCheckpointor) finalizeCheckpoint(ctx context.Context, task *SigningTask) (*Checkpoint, error) {
	// Aggregate signatures
	aggregatedSig, err := c.blsAgg.AggregateSignatures(task.Partials, c.committee.Threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate signatures: %w", err)
	}
	
	// Create checkpoint
	signers := make([]string, len(task.Partials))
	for i, partial := range task.Partials {
		signers[i] = partial.SignerDID
	}
	
	checkpoint := &Checkpoint{
		Root:      task.Root,
		Epoch:     task.Epoch,
		TreeSize:  task.TreeSize,
		Signers:   signers,
		Signature: aggregatedSig,
		Timestamp: time.Now(),
	}
	
	// Store checkpoint
	if err := c.store.StoreCheckpoint(ctx, checkpoint); err != nil {
		return nil, fmt.Errorf("failed to store checkpoint: %w", err)
	}
	
	// Mark task as complete
	task.Status = TaskStatusComplete
	
	// Clear current task
	c.mu.Lock()
	c.currentTask = nil
	c.mu.Unlock()
	
	return checkpoint, nil
}

// Utility functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}