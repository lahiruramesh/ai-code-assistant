package agents

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// RequestStatus represents the status of a request
type RequestStatus string

const (
	RequestPending    RequestStatus = "pending"
	RequestProcessing RequestStatus = "processing"
	RequestCompleted  RequestStatus = "completed"
	RequestFailed     RequestStatus = "failed"
	RequestTimeout    RequestStatus = "timeout"
)

// AgentLoop represents a single agent processing loop for a request
type AgentLoop struct {
	ID          string
	RequestID   string
	UserRequest string
	Status      RequestStatus
	StartTime   time.Time
	EndTime     *time.Time
	Coordinator *Coordinator
	Context     context.Context
	Cancel      context.CancelFunc
	Result      chan AgentLoopResult
	ErrorChan   chan error
	mutex       sync.RWMutex
}

// AgentLoopResult contains the result of an agent loop
type AgentLoopResult struct {
	RequestID   string
	Status      RequestStatus
	Messages    []AgentMessage
	Error       error
	Duration    time.Duration
	CompletedAt time.Time
}

// LoopManager manages multiple concurrent agent loops
type LoopManager struct {
	loops       map[string]*AgentLoop
	coordinator *Coordinator
	mutex       sync.RWMutex
	resultChan  chan AgentLoopResult
	maxTimeout  time.Duration
}

// NewLoopManager creates a new loop manager
func NewLoopManager(coordinator *Coordinator) *LoopManager {
	return &LoopManager{
		loops:       make(map[string]*AgentLoop),
		coordinator: coordinator,
		resultChan:  make(chan AgentLoopResult, 100),
		maxTimeout:  20 * time.Minute, // 20 minute timeout
	}
}

// StartLoop starts a new agent loop for a request
func (lm *LoopManager) StartLoop(requestID, userRequest string) (*AgentLoop, error) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// Check if request already exists
	if _, exists := lm.loops[requestID]; exists {
		return nil, fmt.Errorf("request %s already being processed", requestID)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), lm.maxTimeout)

	// Create new agent loop
	loop := &AgentLoop{
		ID:          generateLoopID(),
		RequestID:   requestID,
		UserRequest: userRequest,
		Status:      RequestPending,
		StartTime:   time.Now(),
		Coordinator: lm.coordinator,
		Context:     ctx,
		Cancel:      cancel,
		Result:      make(chan AgentLoopResult, 1),
		ErrorChan:   make(chan error, 1),
	}

	// Store the loop
	lm.loops[requestID] = loop

	// Start the loop in a goroutine
	go lm.runLoop(loop)

	log.Printf("Started agent loop %s for request %s", loop.ID, requestID)
	return loop, nil
}

// runLoop executes the agent loop
func (lm *LoopManager) runLoop(loop *AgentLoop) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Agent loop %s panicked: %v", loop.ID, r)
			lm.completeLoop(loop, RequestFailed, fmt.Errorf("loop panicked: %v", r))
		}
	}()

	loop.setStatus(RequestProcessing)
	log.Printf("Running agent loop %s for request: %s", loop.ID, loop.UserRequest)

	// Process the request through the coordinator
	err := loop.Coordinator.ProcessUserRequest(loop.UserRequest)
	if err != nil {
		log.Printf("Error starting request processing in loop %s: %v", loop.ID, err)
		lm.completeLoop(loop, RequestFailed, err)
		return
	}

	// Monitor the loop until completion or timeout
	go lm.monitorLoop(loop)
}

// monitorLoop monitors a loop for completion or timeout
func (lm *LoopManager) monitorLoop(loop *AgentLoop) {
	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()

	lastActivityTime := time.Now()
	consecutiveIdleChecks := 0

	for {
		select {
		case <-loop.Context.Done():
			// Timeout or cancellation
			if loop.Context.Err() == context.DeadlineExceeded {
				log.Printf("Agent loop %s timed out after %v", loop.ID, time.Since(loop.StartTime))
				lm.completeLoop(loop, RequestTimeout, fmt.Errorf("request timed out after %v", lm.maxTimeout))
			} else {
				log.Printf("Agent loop %s was cancelled", loop.ID)
				lm.completeLoop(loop, RequestFailed, fmt.Errorf("request was cancelled"))
			}
			return

		case <-ticker.C:
			// Check if agents are still processing
			if lm.areAgentsActive(loop) {
				lastActivityTime = time.Now()
				consecutiveIdleChecks = 0
				log.Printf("Agent loop %s: agents still active", loop.ID)
			} else {
				consecutiveIdleChecks++
				log.Printf("Agent loop %s: no agent activity (check %d)", loop.ID, consecutiveIdleChecks)

				// If no activity for 30 seconds (6 checks), consider complete
				if time.Since(lastActivityTime) > 30*time.Second && consecutiveIdleChecks >= 6 {
					log.Printf("Agent loop %s completed after %v", loop.ID, time.Since(loop.StartTime))
					lm.completeLoop(loop, RequestCompleted, nil)
					return
				}
			}
		}
	}
}

// areAgentsActive checks if any agents are currently processing
func (lm *LoopManager) areAgentsActive(loop *AgentLoop) bool {
	coordinator := loop.Coordinator

	// Check for pending messages
	totalPending := coordinator.getTotalPendingMessages()
	if totalPending > 0 {
		return true
	}

	// Check for agents currently processing
	activeProcessing := coordinator.getActiveProcessingCount()
	if activeProcessing > 0 {
		return true
	}

	return false
}

// completeLoop marks a loop as complete and cleans up
func (lm *LoopManager) completeLoop(loop *AgentLoop, status RequestStatus, err error) {
	loop.setStatus(status)
	endTime := time.Now()
	loop.setEndTime(&endTime)

	// Create result
	result := AgentLoopResult{
		RequestID:   loop.RequestID,
		Status:      status,
		Error:       err,
		Duration:    time.Since(loop.StartTime),
		CompletedAt: endTime,
	}

	// Send result to channel
	select {
	case loop.Result <- result:
	case <-time.After(1 * time.Second):
		log.Printf("Warning: Failed to send result for loop %s", loop.ID)
	}

	// Send to manager result channel
	select {
	case lm.resultChan <- result:
	case <-time.After(1 * time.Second):
		log.Printf("Warning: Failed to send result to manager for loop %s", loop.ID)
	}

	// Clean up
	loop.Cancel()

	// Remove from active loops
	lm.mutex.Lock()
	delete(lm.loops, loop.RequestID)
	lm.mutex.Unlock()

	log.Printf("Completed agent loop %s for request %s with status %s", loop.ID, loop.RequestID, status)
}

// GetLoop returns a loop by request ID
func (lm *LoopManager) GetLoop(requestID string) (*AgentLoop, bool) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	loop, exists := lm.loops[requestID]
	return loop, exists
}

// GetActiveLoops returns all currently active loops
func (lm *LoopManager) GetActiveLoops() []*AgentLoop {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	loops := make([]*AgentLoop, 0, len(lm.loops))
	for _, loop := range lm.loops {
		loops = append(loops, loop)
	}
	return loops
}

// CancelLoop cancels a specific loop
func (lm *LoopManager) CancelLoop(requestID string) error {
	lm.mutex.RLock()
	loop, exists := lm.loops[requestID]
	lm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("loop for request %s not found", requestID)
	}

	loop.Cancel()
	return nil
}

// GetResultChannel returns the result channel for listening to completed loops
func (lm *LoopManager) GetResultChannel() <-chan AgentLoopResult {
	return lm.resultChan
}

// Stop stops all active loops and cleans up
func (lm *LoopManager) Stop() {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	log.Println("Stopping all agent loops...")
	for requestID, loop := range lm.loops {
		log.Printf("Cancelling loop for request %s", requestID)
		loop.Cancel()
	}

	// Clear all loops
	lm.loops = make(map[string]*AgentLoop)
}

// Helper methods for AgentLoop

func (al *AgentLoop) setStatus(status RequestStatus) {
	al.mutex.Lock()
	defer al.mutex.Unlock()
	al.Status = status
}

func (al *AgentLoop) GetStatus() RequestStatus {
	al.mutex.RLock()
	defer al.mutex.RUnlock()
	return al.Status
}

func (al *AgentLoop) setEndTime(endTime *time.Time) {
	al.mutex.Lock()
	defer al.mutex.Unlock()
	al.EndTime = endTime
}

func (al *AgentLoop) GetDuration() time.Duration {
	al.mutex.RLock()
	defer al.mutex.RUnlock()
	if al.EndTime != nil {
		return al.EndTime.Sub(al.StartTime)
	}
	return time.Since(al.StartTime)
}

// generateLoopID generates a unique ID for agent loops
func generateLoopID() string {
	return fmt.Sprintf("loop_%d", time.Now().UnixNano())
}
