package consensus

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// MemoryCheckpointStore provides an in-memory implementation of CheckpointStore
type MemoryCheckpointStore struct {
	mu            sync.RWMutex
	checkpoints   map[int64]*Checkpoint
	committees    map[int64]*Committee
	signingTasks  map[string]*SigningTask
	currentCommittee *Committee
	currentTask   *SigningTask
}

// NewMemoryCheckpointStore creates a new in-memory checkpoint store
func NewMemoryCheckpointStore() *MemoryCheckpointStore {
	return &MemoryCheckpointStore{
		checkpoints:  make(map[int64]*Checkpoint),
		committees:   make(map[int64]*Committee),
		signingTasks: make(map[string]*SigningTask),
	}
}

// StoreCheckpoint stores a checkpoint
func (s *MemoryCheckpointStore) StoreCheckpoint(ctx context.Context, checkpoint *Checkpoint) error {
	if checkpoint == nil {
		return fmt.Errorf("checkpoint cannot be nil")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Create a copy to avoid external modifications
	stored := &Checkpoint{
		Root:      make([]byte, len(checkpoint.Root)),
		Epoch:     checkpoint.Epoch,
		TreeSize:  checkpoint.TreeSize,
		Signers:   make([]string, len(checkpoint.Signers)),
		Signature: make([]byte, len(checkpoint.Signature)),
		Timestamp: checkpoint.Timestamp,
	}
	
	copy(stored.Root, checkpoint.Root)
	copy(stored.Signers, checkpoint.Signers)
	copy(stored.Signature, checkpoint.Signature)
	
	s.checkpoints[checkpoint.Epoch] = stored
	return nil
}

// GetCheckpoint retrieves a checkpoint by epoch
func (s *MemoryCheckpointStore) GetCheckpoint(ctx context.Context, epoch int64) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	checkpoint, exists := s.checkpoints[epoch]
	if !exists {
		return nil, fmt.Errorf("checkpoint not found for epoch %d", epoch)
	}
	
	// Return a copy to prevent external modifications
	return s.copyCheckpoint(checkpoint), nil
}

// GetLatestCheckpoint retrieves the latest checkpoint
func (s *MemoryCheckpointStore) GetLatestCheckpoint(ctx context.Context) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if len(s.checkpoints) == 0 {
		return nil, fmt.Errorf("no checkpoints found")
	}
	
	// Find the checkpoint with the highest epoch
	var latestEpoch int64
	var latestCheckpoint *Checkpoint
	
	for epoch, checkpoint := range s.checkpoints {
		if epoch > latestEpoch {
			latestEpoch = epoch
			latestCheckpoint = checkpoint
		}
	}
	
	return s.copyCheckpoint(latestCheckpoint), nil
}

// ListCheckpoints retrieves checkpoints in a range
func (s *MemoryCheckpointStore) ListCheckpoints(ctx context.Context, startEpoch, endEpoch int64) ([]*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var checkpoints []*Checkpoint
	
	for epoch, checkpoint := range s.checkpoints {
		if epoch >= startEpoch && epoch <= endEpoch {
			checkpoints = append(checkpoints, s.copyCheckpoint(checkpoint))
		}
	}
	
	// Sort by epoch
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].Epoch < checkpoints[j].Epoch
	})
	
	return checkpoints, nil
}

// StoreCommittee stores a committee
func (s *MemoryCheckpointStore) StoreCommittee(ctx context.Context, committee *Committee) error {
	if committee == nil {
		return fmt.Errorf("committee cannot be nil")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Create a copy to avoid external modifications
	stored := &Committee{
		Members:   make([]string, len(committee.Members)),
		Threshold: committee.Threshold,
		Epoch:     committee.Epoch,
		VRFSeed:   make([]byte, len(committee.VRFSeed)),
		StartTime: committee.StartTime,
		EndTime:   committee.EndTime,
	}
	
	copy(stored.Members, committee.Members)
	copy(stored.VRFSeed, committee.VRFSeed)
	
	s.committees[committee.Epoch] = stored
	s.currentCommittee = stored
	
	return nil
}

// GetCurrentCommittee retrieves the current committee
func (s *MemoryCheckpointStore) GetCurrentCommittee(ctx context.Context) (*Committee, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.currentCommittee == nil {
		return nil, fmt.Errorf("no current committee")
	}
	
	return s.copyCommittee(s.currentCommittee), nil
}

// StoreSigningTask stores a signing task
func (s *MemoryCheckpointStore) StoreSigningTask(ctx context.Context, task *SigningTask) error {
	if task == nil {
		return fmt.Errorf("signing task cannot be nil")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Create task ID from epoch
	taskID := fmt.Sprintf("task-%d", task.Epoch)
	
	// Create a copy to avoid external modifications
	stored := &SigningTask{
		Root:     make([]byte, len(task.Root)),
		Epoch:    task.Epoch,
		TreeSize: task.TreeSize,
		Deadline: task.Deadline,
		Partials: make([]PartialSignature, len(task.Partials)),
		Status:   task.Status,
	}
	
	copy(stored.Root, task.Root)
	for i, partial := range task.Partials {
		stored.Partials[i] = s.copyPartialSignature(partial)
	}
	
	s.signingTasks[taskID] = stored
	s.currentTask = stored
	
	return nil
}

// GetCurrentSigningTask retrieves the current signing task
func (s *MemoryCheckpointStore) GetCurrentSigningTask(ctx context.Context) (*SigningTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.currentTask == nil {
		return nil, fmt.Errorf("no current signing task")
	}
	
	return s.copySigningTask(s.currentTask), nil
}

// AddPartialSignature adds a partial signature to a signing task
func (s *MemoryCheckpointStore) AddPartialSignature(ctx context.Context, taskID string, partial *PartialSignature) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	task, exists := s.signingTasks[taskID]
	if !exists {
		return fmt.Errorf("signing task %s not found", taskID)
	}
	
	// Check if signature from this signer already exists
	for _, existing := range task.Partials {
		if existing.SignerDID == partial.SignerDID {
			return fmt.Errorf("partial signature from %s already exists", partial.SignerDID)
		}
	}
	
	// Add the partial signature
	task.Partials = append(task.Partials, s.copyPartialSignature(*partial))
	
	return nil
}

// Helper functions for deep copying

func (s *MemoryCheckpointStore) copyCheckpoint(checkpoint *Checkpoint) *Checkpoint {
	if checkpoint == nil {
		return nil
	}
	
	copied := &Checkpoint{
		Root:      make([]byte, len(checkpoint.Root)),
		Epoch:     checkpoint.Epoch,
		TreeSize:  checkpoint.TreeSize,
		Signers:   make([]string, len(checkpoint.Signers)),
		Signature: make([]byte, len(checkpoint.Signature)),
		Timestamp: checkpoint.Timestamp,
	}
	
	copy(copied.Root, checkpoint.Root)
	copy(copied.Signers, checkpoint.Signers)
	copy(copied.Signature, checkpoint.Signature)
	
	return copied
}

func (s *MemoryCheckpointStore) copyCommittee(committee *Committee) *Committee {
	if committee == nil {
		return nil
	}
	
	copied := &Committee{
		Members:   make([]string, len(committee.Members)),
		Threshold: committee.Threshold,
		Epoch:     committee.Epoch,
		VRFSeed:   make([]byte, len(committee.VRFSeed)),
		StartTime: committee.StartTime,
		EndTime:   committee.EndTime,
	}
	
	copy(copied.Members, committee.Members)
	copy(copied.VRFSeed, committee.VRFSeed)
	
	return copied
}

func (s *MemoryCheckpointStore) copySigningTask(task *SigningTask) *SigningTask {
	if task == nil {
		return nil
	}
	
	copied := &SigningTask{
		Root:     make([]byte, len(task.Root)),
		Epoch:    task.Epoch,
		TreeSize: task.TreeSize,
		Deadline: task.Deadline,
		Partials: make([]PartialSignature, len(task.Partials)),
		Status:   task.Status,
	}
	
	copy(copied.Root, task.Root)
	for i, partial := range task.Partials {
		copied.Partials[i] = s.copyPartialSignature(partial)
	}
	
	return copied
}

func (s *MemoryCheckpointStore) copyPartialSignature(partial PartialSignature) PartialSignature {
	copied := PartialSignature{
		SignerDID: partial.SignerDID,
		Signature: make([]byte, len(partial.Signature)),
		Root:      make([]byte, len(partial.Root)),
		Epoch:     partial.Epoch,
		Timestamp: partial.Timestamp,
	}
	
	copy(copied.Signature, partial.Signature)
	copy(copied.Root, partial.Root)
	
	return copied
}