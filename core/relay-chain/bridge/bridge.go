package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Shivam-Patel-G/blackhole-blockchain/core/relay-chain/chain"
)

// ChainType represents different blockchain types
type ChainType string

const (
	ChainTypeBlackhole ChainType = "blackhole"
	ChainTypeEthereum  ChainType = "ethereum"
	ChainTypeSolana    ChainType = "solana"
	ChainTypePolkadot  ChainType = "polkadot"
)

// Enhanced retry configuration
type RetryConfig struct {
	MaxAttempts   int           `json:"max_attempts"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
	JitterFactor  float64       `json:"jitter_factor"`
	Timeout       time.Duration `json:"timeout"`
}

// Event streaming structures
type BridgeEvent struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	BridgeTxID string                 `json:"bridge_tx_id"`
	Chain      ChainType              `json:"chain"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  int64                  `json:"timestamp"`
	Status     string                 `json:"status"`
	Error      string                 `json:"error,omitempty"`
}

type EventStream struct {
	subscribers map[string]chan *BridgeEvent
	mu          sync.RWMutex
	events      []*BridgeEvent
	maxEvents   int
}

// Enhanced bridge transaction with retry tracking
type BridgeTransaction struct {
	ID              string    `json:"id"`
	SourceChain     ChainType `json:"source_chain"`
	DestChain       ChainType `json:"dest_chain"`
	SourceAddress   string    `json:"source_address"`
	DestAddress     string    `json:"dest_address"`
	TokenSymbol     string    `json:"token_symbol"`
	Amount          uint64    `json:"amount"`
	Status          string    `json:"status"` // "pending", "confirmed", "completed", "failed", "retrying"
	CreatedAt       int64     `json:"created_at"`
	ConfirmedAt     int64     `json:"confirmed_at,omitempty"`
	CompletedAt     int64     `json:"completed_at,omitempty"`
	SourceTxHash    string    `json:"source_tx_hash,omitempty"`
	DestTxHash      string    `json:"dest_tx_hash,omitempty"`
	RelaySignatures []string  `json:"relay_signatures"`

	// Enhanced retry tracking
	RetryAttempts int    `json:"retry_attempts"`
	LastRetryAt   int64  `json:"last_retry_at,omitempty"`
	NextRetryAt   int64  `json:"next_retry_at,omitempty"`
	RetryReason   string `json:"retry_reason,omitempty"`
	FailureCount  int    `json:"failure_count"`
	LastError     string `json:"last_error,omitempty"`

	// Event tracking
	Events []*BridgeEvent `json:"events,omitempty"`

	mu sync.RWMutex
}

// RelayNode represents a bridge relay node
type RelayNode struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
	Active    bool   `json:"active"`
}

// Bridge manages cross-chain operations
type Bridge struct {
	SupportedChains map[ChainType]bool              `json:"supported_chains"`
	Transactions    map[string]*BridgeTransaction   `json:"transactions"`
	RelayNodes      map[string]*RelayNode           `json:"relay_nodes"`
	TokenMappings   map[ChainType]map[string]string `json:"token_mappings"` // chain -> original_token -> wrapped_token
	Blockchain      *chain.Blockchain               `json:"-"`

	// Enhanced components
	RetryConfig *RetryConfig       `json:"retry_config"`
	EventStream *EventStream       `json:"-"`
	Context     context.Context    `json:"-"`
	CancelFunc  context.CancelFunc `json:"-"`

	mu sync.RWMutex
}

// NewBridge creates a new bridge instance
func NewBridge(blockchain *chain.Blockchain) *Bridge {
	ctx, cancel := context.WithCancel(context.Background())

	// Enhanced retry configuration
	retryConfig := &RetryConfig{
		MaxAttempts:   5,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.1,
		Timeout:       60 * time.Second,
	}

	// Initialize event streaming
	eventStream := &EventStream{
		subscribers: make(map[string]chan *BridgeEvent),
		events:      make([]*BridgeEvent, 0),
		maxEvents:   1000,
	}

	bridge := &Bridge{
		SupportedChains: make(map[ChainType]bool),
		Transactions:    make(map[string]*BridgeTransaction),
		RelayNodes:      make(map[string]*RelayNode),
		TokenMappings:   make(map[ChainType]map[string]string),
		Blockchain:      blockchain,
		RetryConfig:     retryConfig,
		EventStream:     eventStream,
		Context:         ctx,
		CancelFunc:      cancel,
	}

	// Initialize supported chains
	bridge.SupportedChains[ChainTypeBlackhole] = true
	bridge.SupportedChains[ChainTypeEthereum] = true
	bridge.SupportedChains[ChainTypePolkadot] = true

	// Initialize token mappings
	bridge.TokenMappings[ChainTypeBlackhole] = make(map[string]string)
	bridge.TokenMappings[ChainTypeEthereum] = make(map[string]string)
	bridge.TokenMappings[ChainTypePolkadot] = make(map[string]string)

	// Mock token mappings
	bridge.TokenMappings[ChainTypeBlackhole]["BHX"] = "BHX"
	bridge.TokenMappings[ChainTypeEthereum]["BHX"] = "wBHX" // Wrapped BHX on Ethereum
	bridge.TokenMappings[ChainTypePolkadot]["BHX"] = "pBHX" // Polkadot BHX

	// Initialize mock relay nodes
	bridge.AddRelayNode("relay1", "relay1_address", "relay1_pubkey")
	bridge.AddRelayNode("relay2", "relay2_address", "relay2_pubkey")
	bridge.AddRelayNode("relay3", "relay3_address", "relay3_pubkey")

	// Start background retry processor
	go bridge.startRetryProcessor()

	return bridge
}

// AddRelayNode adds a new relay node
func (b *Bridge) AddRelayNode(id, address, publicKey string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.RelayNodes[id] = &RelayNode{
		ID:        id,
		Address:   address,
		PublicKey: publicKey,
		Active:    true,
	}
}

// Enhanced retry processor
func (b *Bridge) startRetryProcessor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-b.Context.Done():
			return
		case <-ticker.C:
			b.processRetryQueue()
		}
	}
}

func (b *Bridge) processRetryQueue() {
	b.mu.RLock()
	var retryTxs []*BridgeTransaction
	for _, tx := range b.Transactions {
		if tx.Status == "retrying" && time.Now().Unix() >= tx.NextRetryAt {
			retryTxs = append(retryTxs, tx)
		}
	}
	b.mu.RUnlock()

	for _, tx := range retryTxs {
		go b.retryBridgeTransaction(tx)
	}
}

func (b *Bridge) retryBridgeTransaction(tx *BridgeTransaction) {
	tx.mu.Lock()
	tx.RetryAttempts++
	tx.LastRetryAt = time.Now().Unix()
	tx.Status = "pending"
	tx.mu.Unlock()

	// Emit retry event
	b.emitEvent(&BridgeEvent{
		ID:         fmt.Sprintf("retry_%s_%d", tx.ID, tx.RetryAttempts),
		Type:       "retry_attempt",
		BridgeTxID: tx.ID,
		Chain:      tx.SourceChain,
		Data: map[string]interface{}{
			"attempt":    tx.RetryAttempts,
			"reason":     tx.RetryReason,
			"last_error": tx.LastError,
		},
		Timestamp: time.Now().Unix(),
		Status:    "retrying",
	})

	// Attempt the operation with timeout
	ctx, cancel := context.WithTimeout(b.Context, b.RetryConfig.Timeout)
	defer cancel()

	// Simulate relay processing with retry logic
	go b.processRelayConfirmationWithRetry(ctx, tx.ID)
}

func (b *Bridge) processRelayConfirmationWithRetry(ctx context.Context, bridgeTxID string) {
	// Simulate relay processing time with potential failure
	processingTime := time.Duration(3+rand.Intn(7)) * time.Second

	select {
	case <-ctx.Done():
		b.handleRetryFailure(bridgeTxID, "timeout")
		return
	case <-time.After(processingTime):
		// Continue processing
	}

	b.mu.Lock()
	bridgeTx, exists := b.Transactions[bridgeTxID]
	if !exists {
		b.mu.Unlock()
		return
	}
	b.mu.Unlock()

	bridgeTx.mu.Lock()
	defer bridgeTx.mu.Unlock()

	// Simulate potential failure (10% chance)
	if rand.Float64() < 0.1 {
		bridgeTx.Status = "failed"
		bridgeTx.LastError = "relay confirmation failed"
		bridgeTx.FailureCount++

		// Check if we should retry
		if bridgeTx.RetryAttempts < b.RetryConfig.MaxAttempts {
			bridgeTx.Status = "retrying"
			bridgeTx.RetryReason = "relay confirmation failed"
			bridgeTx.NextRetryAt = time.Now().Add(b.calculateRetryDelay(bridgeTx.RetryAttempts)).Unix()

			b.emitEvent(&BridgeEvent{
				ID:         fmt.Sprintf("retry_scheduled_%s", bridgeTxID),
				Type:       "retry_scheduled",
				BridgeTxID: bridgeTxID,
				Chain:      bridgeTx.SourceChain,
				Data: map[string]interface{}{
					"next_retry_at": bridgeTx.NextRetryAt,
					"attempt":       bridgeTx.RetryAttempts,
				},
				Timestamp: time.Now().Unix(),
				Status:    "scheduled",
			})
		} else {
			b.emitEvent(&BridgeEvent{
				ID:         fmt.Sprintf("max_retries_exceeded_%s", bridgeTxID),
				Type:       "max_retries_exceeded",
				BridgeTxID: bridgeTxID,
				Chain:      bridgeTx.SourceChain,
				Data: map[string]interface{}{
					"max_attempts":   b.RetryConfig.MaxAttempts,
					"total_attempts": bridgeTx.RetryAttempts,
				},
				Timestamp: time.Now().Unix(),
				Status:    "failed",
				Error:     "max retry attempts exceeded",
			})
		}
		return
	}

	// Success - simulate relay signatures
	relayCount := 0
	for relayID := range b.RelayNodes {
		if relayCount >= 2 {
			break
		}
		bridgeTx.RelaySignatures = append(bridgeTx.RelaySignatures, fmt.Sprintf("sig_%s_%s", relayID, bridgeTxID))
		relayCount++
	}

	bridgeTx.Status = "confirmed"
	bridgeTx.ConfirmedAt = time.Now().Unix()

	// Emit confirmation event
	b.emitEvent(&BridgeEvent{
		ID:         fmt.Sprintf("confirmed_%s", bridgeTxID),
		Type:       "relay_confirmed",
		BridgeTxID: bridgeTxID,
		Chain:      bridgeTx.SourceChain,
		Data: map[string]interface{}{
			"relay_signatures": bridgeTx.RelaySignatures,
			"relay_count":      len(bridgeTx.RelaySignatures),
		},
		Timestamp: time.Now().Unix(),
		Status:    "confirmed",
	})

	fmt.Printf("✅ Bridge transaction %s confirmed by %d relays\n", bridgeTxID, len(bridgeTx.RelaySignatures))

	// Continue with destination transfer
	go b.processDestinationTransferWithRetry(ctx, bridgeTxID)
}

func (b *Bridge) processDestinationTransferWithRetry(ctx context.Context, bridgeTxID string) {
	// Simulate destination processing time
	processingTime := time.Duration(2+rand.Intn(4)) * time.Second

	select {
	case <-ctx.Done():
		b.handleRetryFailure(bridgeTxID, "destination timeout")
		return
	case <-time.After(processingTime):
		// Continue processing
	}

	b.mu.Lock()
	bridgeTx, exists := b.Transactions[bridgeTxID]
	if !exists {
		b.mu.Unlock()
		return
	}
	b.mu.Unlock()

	bridgeTx.mu.Lock()
	defer bridgeTx.mu.Unlock()

	// Simulate potential failure (5% chance)
	if rand.Float64() < 0.05 {
		bridgeTx.Status = "failed"
		bridgeTx.LastError = "destination transfer failed"
		bridgeTx.FailureCount++

		// Check if we should retry
		if bridgeTx.RetryAttempts < b.RetryConfig.MaxAttempts {
			bridgeTx.Status = "retrying"
			bridgeTx.RetryReason = "destination transfer failed"
			bridgeTx.NextRetryAt = time.Now().Add(b.calculateRetryDelay(bridgeTx.RetryAttempts)).Unix()

			b.emitEvent(&BridgeEvent{
				ID:         fmt.Sprintf("dest_retry_scheduled_%s", bridgeTxID),
				Type:       "destination_retry_scheduled",
				BridgeTxID: bridgeTxID,
				Chain:      bridgeTx.DestChain,
				Data: map[string]interface{}{
					"next_retry_at": bridgeTx.NextRetryAt,
					"attempt":       bridgeTx.RetryAttempts,
				},
				Timestamp: time.Now().Unix(),
				Status:    "scheduled",
			})
		} else {
			b.emitEvent(&BridgeEvent{
				ID:         fmt.Sprintf("dest_max_retries_exceeded_%s", bridgeTxID),
				Type:       "destination_max_retries_exceeded",
				BridgeTxID: bridgeTxID,
				Chain:      bridgeTx.DestChain,
				Data: map[string]interface{}{
					"max_attempts":   b.RetryConfig.MaxAttempts,
					"total_attempts": bridgeTx.RetryAttempts,
				},
				Timestamp: time.Now().Unix(),
				Status:    "failed",
				Error:     "max retry attempts exceeded",
			})
		}
		return
	}

	// Success - process destination transfer
	if bridgeTx.DestChain == ChainTypeBlackhole {
		destToken := b.TokenMappings[bridgeTx.DestChain][bridgeTx.TokenSymbol]
		token, exists := b.Blockchain.TokenRegistry[destToken]
		if exists {
			err := token.Mint(bridgeTx.DestAddress, bridgeTx.Amount)
			if err == nil {
				bridgeTx.DestTxHash = fmt.Sprintf("blackhole_mint_%d", time.Now().UnixNano())
			}
		}
	} else {
		bridgeTx.DestTxHash = fmt.Sprintf("%s_tx_%d", bridgeTx.DestChain, time.Now().UnixNano())
	}

	bridgeTx.Status = "completed"
	bridgeTx.CompletedAt = time.Now().Unix()

	// Emit completion event
	b.emitEvent(&BridgeEvent{
		ID:         fmt.Sprintf("completed_%s", bridgeTxID),
		Type:       "transfer_completed",
		BridgeTxID: bridgeTxID,
		Chain:      bridgeTx.DestChain,
		Data: map[string]interface{}{
			"dest_tx_hash": bridgeTx.DestTxHash,
			"amount":       bridgeTx.Amount,
			"token":        bridgeTx.TokenSymbol,
		},
		Timestamp: time.Now().Unix(),
		Status:    "completed",
	})

	fmt.Printf("✅ Bridge transfer completed: %s (tx: %s)\n", bridgeTxID, bridgeTx.DestTxHash)
}

func (b *Bridge) handleRetryFailure(bridgeTxID string, reason string) {
	b.mu.Lock()
	bridgeTx, exists := b.Transactions[bridgeTxID]
	b.mu.Unlock()

	if !exists {
		return
	}

	bridgeTx.mu.Lock()
	bridgeTx.Status = "failed"
	bridgeTx.LastError = reason
	bridgeTx.FailureCount++
	bridgeTx.mu.Unlock()

	b.emitEvent(&BridgeEvent{
		ID:         fmt.Sprintf("retry_failed_%s", bridgeTxID),
		Type:       "retry_failed",
		BridgeTxID: bridgeTxID,
		Chain:      bridgeTx.SourceChain,
		Data: map[string]interface{}{
			"reason": reason,
		},
		Timestamp: time.Now().Unix(),
		Status:    "failed",
		Error:     reason,
	})
}

func (b *Bridge) calculateRetryDelay(attempt int) time.Duration {
	delay := b.RetryConfig.InitialDelay
	for i := 0; i < attempt; i++ {
		delay = time.Duration(float64(delay) * b.RetryConfig.BackoffFactor)
		if delay > b.RetryConfig.MaxDelay {
			delay = b.RetryConfig.MaxDelay
			break
		}
	}

	// Add jitter
	jitter := time.Duration(float64(delay) * b.RetryConfig.JitterFactor * (rand.Float64() - 0.5))
	delay += jitter

	if delay < 0 {
		delay = b.RetryConfig.InitialDelay
	}

	return delay
}

// Event streaming methods
func (b *Bridge) emitEvent(event *BridgeEvent) {
	// Add to event history
	b.EventStream.mu.Lock()
	b.EventStream.events = append(b.EventStream.events, event)

	// Keep only recent events
	if len(b.EventStream.events) > b.EventStream.maxEvents {
		b.EventStream.events = b.EventStream.events[1:]
	}
	b.EventStream.mu.Unlock()

	// Broadcast to subscribers
	b.EventStream.mu.RLock()
	for _, ch := range b.EventStream.subscribers {
		select {
		case ch <- event:
		default:
			// Channel is full, skip this subscriber
		}
	}
	b.EventStream.mu.RUnlock()

	// Add to transaction events
	b.mu.RLock()
	if tx, exists := b.Transactions[event.BridgeTxID]; exists {
		tx.mu.Lock()
		tx.Events = append(tx.Events, event)
		tx.mu.Unlock()
	}
	b.mu.RUnlock()
}

func (b *Bridge) SubscribeToEvents(subscriberID string) chan *BridgeEvent {
	ch := make(chan *BridgeEvent, 100)

	b.EventStream.mu.Lock()
	b.EventStream.subscribers[subscriberID] = ch
	b.EventStream.mu.Unlock()

	return ch
}

func (b *Bridge) UnsubscribeFromEvents(subscriberID string) {
	b.EventStream.mu.Lock()
	if ch, exists := b.EventStream.subscribers[subscriberID]; exists {
		close(ch)
		delete(b.EventStream.subscribers, subscriberID)
	}
	b.EventStream.mu.Unlock()
}

func (b *Bridge) GetRecentEvents(limit int) []*BridgeEvent {
	b.EventStream.mu.RLock()
	defer b.EventStream.mu.RUnlock()

	if limit > len(b.EventStream.events) {
		limit = len(b.EventStream.events)
	}

	events := make([]*BridgeEvent, limit)
	copy(events, b.EventStream.events[len(b.EventStream.events)-limit:])

	return events
}

// Enhanced bridge transfer with retry support
func (b *Bridge) InitiateBridgeTransfer(sourceChain, destChain ChainType, sourceAddr, destAddr, tokenSymbol string, amount uint64) (*BridgeTransaction, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Validate chains
	if !b.SupportedChains[sourceChain] {
		return nil, fmt.Errorf("unsupported source chain: %s", sourceChain)
	}
	if !b.SupportedChains[destChain] {
		return nil, fmt.Errorf("unsupported destination chain: %s", destChain)
	}

	// Validate addresses
	if sourceAddr == "" || destAddr == "" {
		return nil, fmt.Errorf("source and destination addresses are required")
	}

	// Validate amount
	if amount == 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// Check token mapping
	if _, exists := b.TokenMappings[destChain][tokenSymbol]; !exists {
		return nil, fmt.Errorf("token %s not supported on destination chain %s", tokenSymbol, destChain)
	}

	// Generate bridge transaction ID
	bridgeTxID := fmt.Sprintf("bridge_%s_%d", sourceChain, time.Now().UnixNano())

	// Create bridge transaction
	bridgeTx := &BridgeTransaction{
		ID:            bridgeTxID,
		SourceChain:   sourceChain,
		DestChain:     destChain,
		SourceAddress: sourceAddr,
		DestAddress:   destAddr,
		TokenSymbol:   tokenSymbol,
		Amount:        amount,
		Status:        "pending",
		CreatedAt:     time.Now().Unix(),
		Events:        make([]*BridgeEvent, 0),
	}

	// Store transaction
	b.Transactions[bridgeTxID] = bridgeTx

	// Emit initiation event
	b.emitEvent(&BridgeEvent{
		ID:         fmt.Sprintf("initiated_%s", bridgeTxID),
		Type:       "transfer_initiated",
		BridgeTxID: bridgeTxID,
		Chain:      sourceChain,
		Data: map[string]interface{}{
			"source_chain": sourceChain,
			"dest_chain":   destChain,
			"amount":       amount,
			"token":        tokenSymbol,
		},
		Timestamp: time.Now().Unix(),
		Status:    "initiated",
	})

	// Start processing with retry logic
	ctx, cancel := context.WithTimeout(b.Context, b.RetryConfig.Timeout)
	defer cancel()

	go b.processRelayConfirmationWithRetry(ctx, bridgeTxID)

	return bridgeTx, nil
}

// GetBridgeTransaction returns a bridge transaction
func (b *Bridge) GetBridgeTransaction(bridgeTxID string) (*BridgeTransaction, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	bridgeTx, exists := b.Transactions[bridgeTxID]
	if !exists {
		return nil, fmt.Errorf("bridge transaction %s not found", bridgeTxID)
	}

	// Return a copy
	bridgeTxCopy := *bridgeTx
	return &bridgeTxCopy, nil
}

// GetUserBridgeTransactions returns all bridge transactions for a user
func (b *Bridge) GetUserBridgeTransactions(userAddress string) []*BridgeTransaction {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var userTxs []*BridgeTransaction
	for _, bridgeTx := range b.Transactions {
		if bridgeTx.SourceAddress == userAddress || bridgeTx.DestAddress == userAddress {
			bridgeTxCopy := *bridgeTx
			userTxs = append(userTxs, &bridgeTxCopy)
		}
	}

	return userTxs
}

// GetSupportedChains returns list of supported chains
func (b *Bridge) GetSupportedChains() []ChainType {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var chains []ChainType
	for chain, supported := range b.SupportedChains {
		if supported {
			chains = append(chains, chain)
		}
	}

	return chains
}

// GetTokenMapping returns token mapping for a chain
func (b *Bridge) GetTokenMapping(chain ChainType) map[string]string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	mapping := make(map[string]string)
	if chainMapping, exists := b.TokenMappings[chain]; exists {
		for k, v := range chainMapping {
			mapping[k] = v
		}
	}

	return mapping
}

// GenerateTestBridgeTransaction creates a test bridge transaction JSON
func (b *Bridge) GenerateTestBridgeTransaction() string {
	testTx := map[string]interface{}{
		"id":               "bridge_test_12345",
		"source_chain":     "blackhole",
		"dest_chain":       "ethereum",
		"source_address":   "blackhole_addr_123",
		"dest_address":     "0x742d35Cc6634C0532925a3b8D4C9db96590b5",
		"token_symbol":     "BHX",
		"amount":           1000,
		"status":           "pending",
		"created_at":       time.Now().Unix(),
		"relay_signatures": []string{},
	}

	jsonData, _ := json.MarshalIndent(testTx, "", "  ")
	return string(jsonData)
}

// ApprovalSimulation represents the result of a bridge approval simulation
type ApprovalSimulation struct {
	Valid               bool     `json:"valid"`
	TokenSymbol         string   `json:"token_symbol"`
	Owner               string   `json:"owner"`
	Spender             string   `json:"spender"`
	RequestedAmount     uint64   `json:"requested_amount"`
	CurrentAllowance    uint64   `json:"current_allowance"`
	CurrentBalance      uint64   `json:"current_balance"`
	SufficientBalance   bool     `json:"sufficient_balance"`
	SufficientAllowance bool     `json:"sufficient_allowance"`
	Warnings            []string `json:"warnings"`
	EstimatedGasCost    uint64   `json:"estimated_gas_cost"`
	Timestamp           int64    `json:"timestamp"`
}

// SimulateApproval simulates a token approval for bridge operations
func (b *Bridge) SimulateApproval(sourceChain ChainType, tokenSymbol, owner, spender string, amount uint64) (*ApprovalSimulation, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	simulation := &ApprovalSimulation{
		TokenSymbol:      tokenSymbol,
		Owner:            owner,
		Spender:          spender,
		RequestedAmount:  amount,
		Warnings:         make([]string, 0),
		EstimatedGasCost: 45000, // Standard ERC-20 approval gas cost
		Timestamp:        time.Now().Unix(),
	}

	// Get token from blockchain registry
	token, exists := b.Blockchain.TokenRegistry[tokenSymbol]
	if !exists {
		simulation.Valid = false
		simulation.Warnings = append(simulation.Warnings, fmt.Sprintf("Token %s not found", tokenSymbol))
		return simulation, nil
	}

	// Check current balance
	balance, err := token.BalanceOf(owner)
	if err != nil {
		simulation.Valid = false
		simulation.Warnings = append(simulation.Warnings, fmt.Sprintf("Failed to check balance: %v", err))
		return simulation, nil
	}
	simulation.CurrentBalance = balance
	simulation.SufficientBalance = balance >= amount

	// Check current allowance
	allowance, err := token.Allowance(owner, spender)
	if err != nil {
		simulation.Valid = false
		simulation.Warnings = append(simulation.Warnings, fmt.Sprintf("Failed to check allowance: %v", err))
		return simulation, nil
	}
	simulation.CurrentAllowance = allowance
	simulation.SufficientAllowance = allowance >= amount

	// Validate approval requirements
	if !simulation.SufficientBalance {
		simulation.Warnings = append(simulation.Warnings,
			fmt.Sprintf("Insufficient balance: has %d, needs %d", balance, amount))
	}

	if !simulation.SufficientAllowance {
		simulation.Warnings = append(simulation.Warnings,
			fmt.Sprintf("Insufficient allowance: has %d, needs %d", allowance, amount))
	}

	// Check for common issues
	if amount > 1000000000 { // Very large amount
		simulation.Warnings = append(simulation.Warnings, "Large amount detected - please verify")
	}

	if owner == spender {
		simulation.Warnings = append(simulation.Warnings, "Owner and spender are the same address")
	}

	// Simulation is valid if balance and allowance are sufficient
	simulation.Valid = simulation.SufficientBalance && simulation.SufficientAllowance

	return simulation, nil
}

// ValidateApprovalForBridge validates that a bridge transaction has proper approvals
func (b *Bridge) ValidateApprovalForBridge(bridgeTx *BridgeTransaction) error {
	if bridgeTx.SourceChain != ChainTypeBlackhole {
		// For external chains, we assume approvals are handled externally
		return nil
	}

	// For Blackhole chain, validate token approval
	simulation, err := b.SimulateApproval(
		bridgeTx.SourceChain,
		bridgeTx.TokenSymbol,
		bridgeTx.SourceAddress,
		"bridge_contract", // Bridge contract as spender
		bridgeTx.Amount,
	)
	if err != nil {
		return fmt.Errorf("approval simulation failed: %v", err)
	}

	if !simulation.Valid {
		return fmt.Errorf("bridge approval validation failed: %v", simulation.Warnings)
	}

	if len(simulation.Warnings) > 0 {
		fmt.Printf("⚠️ Bridge approval warnings: %v\n", simulation.Warnings)
	}

	return nil
}

// PreValidateBridgeTransfer performs pre-flight validation of a bridge transfer
func (b *Bridge) PreValidateBridgeTransfer(sourceAddr, tokenSymbol string, amount uint64) error {
	// Check if token exists
	token, exists := b.Blockchain.TokenRegistry[tokenSymbol]
	if !exists {
		return fmt.Errorf("token %s not found", tokenSymbol)
	}

	// Check balance
	balance, err := token.BalanceOf(sourceAddr)
	if err != nil {
		return fmt.Errorf("failed to check balance: %v", err)
	}

	if balance < amount {
		return fmt.Errorf("insufficient balance: has %d, needs %d", balance, amount)
	}

	// Check allowance for bridge contract
	allowance, err := token.Allowance(sourceAddr, "bridge_contract")
	if err != nil {
		return fmt.Errorf("failed to check bridge allowance: %v", err)
	}

	if allowance < amount {
		return fmt.Errorf("insufficient bridge allowance: has %d, needs %d. Please approve bridge contract first", allowance, amount)
	}

	return nil
}
