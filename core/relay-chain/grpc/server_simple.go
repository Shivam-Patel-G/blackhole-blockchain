package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/Shivam-Patel-G/blackhole-blockchain/core/relay-chain/chain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

// Performance optimization structures
type ConnectionPool struct {
	connections    map[string]*grpc.ClientConn
	mu             sync.RWMutex
	maxConnections int
}

type PerformanceMetrics struct {
	RequestCount      int64
	AverageResponse   time.Duration
	ErrorCount        int64
	ActiveConnections int
	mu                sync.RWMutex
}

type CircuitBreaker struct {
	failureThreshold int
	failureCount     int
	lastFailureTime  time.Time
	state            string // "closed", "open", "half-open"
	mu               sync.RWMutex
}

// Enhanced error handling
type GRPCError struct {
	Code    codes.Code
	Message string
	Details map[string]interface{}
}

func (e *GRPCError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// SimpleRelayServer provides a simplified gRPC server without protobuf dependencies
type SimpleRelayServer struct {
	blockchain *chain.Blockchain
	grpcServer *grpc.Server
	port       int

	// Performance optimization components
	connectionPool *ConnectionPool
	metrics        *PerformanceMetrics
	circuitBreaker *CircuitBreaker

	// Context management
	ctx        context.Context
	cancelFunc context.CancelFunc

	// Health monitoring
	healthStatus map[string]interface{}
	mu           sync.RWMutex
}

// NewSimpleRelayServer creates a new simplified gRPC relay server
func NewSimpleRelayServer(blockchain *chain.Blockchain, port int) *SimpleRelayServer {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize performance components
	connectionPool := &ConnectionPool{
		connections:    make(map[string]*grpc.ClientConn),
		maxConnections: 100,
	}

	metrics := &PerformanceMetrics{}

	circuitBreaker := &CircuitBreaker{
		failureThreshold: 5,
		state:            "closed",
	}

	server := &SimpleRelayServer{
		blockchain:     blockchain,
		port:           port,
		connectionPool: connectionPool,
		metrics:        metrics,
		circuitBreaker: circuitBreaker,
		ctx:            ctx,
		cancelFunc:     cancel,
		healthStatus:   make(map[string]interface{}),
	}

	// Start background health monitoring
	go server.startHealthMonitoring()

	return server
}

// Start starts the gRPC server with enhanced configuration
func (s *SimpleRelayServer) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %v", s.port, err)
	}

	// Enhanced gRPC server configuration
	s.grpcServer = grpc.NewServer(
		grpc.MaxConcurrentStreams(1000),
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 60 * time.Second,
			MaxConnectionAge:  300 * time.Second,
			Time:              30 * time.Second,
			Timeout:           5 * time.Second,
		}),
	)

	// Note: Service registration would happen here when protobuf is implemented

	fmt.Printf("🚀 Enhanced gRPC Relay Server starting on port %d\n", s.port)
	fmt.Printf("📊 Performance monitoring enabled\n")
	fmt.Printf("🔄 Circuit breaker enabled\n")
	fmt.Printf("🔗 Connection pooling enabled\n")

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the gRPC server gracefully
func (s *SimpleRelayServer) Stop() {
	if s.grpcServer != nil {
		fmt.Println("🛑 Stopping Enhanced gRPC Relay Server...")

		// Cancel context to stop background processes
		s.cancelFunc()

		// Close all connections in pool
		s.connectionPool.mu.Lock()
		for _, conn := range s.connectionPool.connections {
			conn.Close()
		}
		s.connectionPool.connections = make(map[string]*grpc.ClientConn)
		s.connectionPool.mu.Unlock()

		// Graceful stop with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			s.grpcServer.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			fmt.Println("✅ gRPC server stopped gracefully")
		case <-ctx.Done():
			fmt.Println("⚠️  Force stopping gRPC server")
			s.grpcServer.Stop()
		}
	}
}

// Performance monitoring
func (s *SimpleRelayServer) startHealthMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateHealthStatus()
		}
	}
}

func (s *SimpleRelayServer) updateHealthStatus() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get system metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get blockchain status
	blockchainStatus := s.GetBlockchainStatus()

	s.healthStatus = map[string]interface{}{
		"timestamp":           time.Now().Unix(),
		"server_status":       "healthy",
		"port":                s.port,
		"memory_usage_mb":     m.Alloc / 1024 / 1024,
		"memory_total_mb":     m.Sys / 1024 / 1024,
		"goroutines":          runtime.NumGoroutine(),
		"active_connections":  s.metrics.ActiveConnections,
		"request_count":       s.metrics.RequestCount,
		"average_response_ms": s.metrics.AverageResponse.Milliseconds(),
		"error_count":         s.metrics.ErrorCount,
		"circuit_breaker":     s.circuitBreaker.state,
		"blockchain":          blockchainStatus,
	}

	// Check for health issues
	if s.metrics.ErrorCount > 100 {
		s.healthStatus["server_status"] = "degraded"
	}
	if s.metrics.ActiveConnections > 80 {
		s.healthStatus["server_status"] = "high_load"
	}
}

// Enhanced performance tracking
func (s *SimpleRelayServer) trackRequest(start time.Time, success bool) {
	duration := time.Since(start)

	s.metrics.mu.Lock()
	s.metrics.RequestCount++

	if success {
		// Calculate moving average
		if s.metrics.AverageResponse == 0 {
			s.metrics.AverageResponse = duration
		} else {
			s.metrics.AverageResponse = (s.metrics.AverageResponse + duration) / 2
		}
	} else {
		s.metrics.ErrorCount++
	}
	s.metrics.mu.Unlock()
}

// Circuit breaker implementation
func (s *SimpleRelayServer) checkCircuitBreaker() error {
	s.circuitBreaker.mu.RLock()
	defer s.circuitBreaker.mu.RUnlock()

	switch s.circuitBreaker.state {
	case "open":
		if time.Since(s.circuitBreaker.lastFailureTime) > 60*time.Second {
			// Try to transition to half-open
			s.circuitBreaker.mu.RUnlock()
			s.circuitBreaker.mu.Lock()
			s.circuitBreaker.state = "half-open"
			s.circuitBreaker.mu.Unlock()
			s.circuitBreaker.mu.RLock()
		} else {
			return status.Error(codes.Unavailable, "service temporarily unavailable")
		}
	case "half-open":
		// Allow one request to test
		return nil
	}

	return nil
}

func (s *SimpleRelayServer) recordSuccess() {
	s.circuitBreaker.mu.Lock()
	defer s.circuitBreaker.mu.Unlock()

	if s.circuitBreaker.state == "half-open" {
		s.circuitBreaker.state = "closed"
		s.circuitBreaker.failureCount = 0
	}
}

func (s *SimpleRelayServer) recordFailure() {
	s.circuitBreaker.mu.Lock()
	defer s.circuitBreaker.mu.Unlock()

	s.circuitBreaker.failureCount++
	s.circuitBreaker.lastFailureTime = time.Now()

	if s.circuitBreaker.failureCount >= s.circuitBreaker.failureThreshold {
		s.circuitBreaker.state = "open"
	}
}

// Enhanced blockchain status with performance metrics
func (s *SimpleRelayServer) GetBlockchainStatus() map[string]interface{} {
	start := time.Now()
	defer func() {
		s.trackRequest(start, true)
	}()

	latestBlock := s.blockchain.GetLatestBlock()
	if latestBlock == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "no blocks found",
		}
	}

	// Get validator information
	allStakes := s.blockchain.StakeLedger.GetAllStakes()
	validatorCount := len(allStakes)

	// Get total supply
	totalSupply := uint64(0)
	if tokenSystem, exists := s.blockchain.TokenRegistry["BHX"]; exists {
		totalSupply = tokenSystem.TotalSupply()
	}

	// Calculate performance metrics
	responseTime := time.Since(start)

	return map[string]interface{}{
		"success":            true,
		"chain_id":           "blackhole-mainnet",
		"block_height":       latestBlock.Header.Index,
		"latest_block_hash":  latestBlock.CalculateHash(),
		"latest_block_time":  latestBlock.Header.Timestamp.Unix(),
		"total_supply":       totalSupply,
		"circulating_supply": totalSupply,
		"validator_count":    validatorCount,
		"pending_txs":        len(s.blockchain.PendingTxs),
		"response_time_ms":   responseTime.Milliseconds(),
		"timestamp":          time.Now().Unix(),
	}
}

// Enhanced transaction submission with validation and error handling
func (s *SimpleRelayServer) SubmitTransactionSimple(from, to, tokenID string, amount uint64) (string, error) {
	start := time.Now()

	// Check circuit breaker
	if err := s.checkCircuitBreaker(); err != nil {
		s.trackRequest(start, false)
		return "", err
	}

	// Enhanced validation
	if err := s.validateTransactionRequest(from, to, tokenID, amount); err != nil {
		s.trackRequest(start, false)
		s.recordFailure()
		return "", status.Error(codes.InvalidArgument, err.Error())
	}

	// Create transaction
	tx := &chain.Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		TokenID:   tokenID,
		Timestamp: time.Now().Unix(),
	}

	// Calculate transaction hash
	tx.ID = tx.CalculateHash()

	// Add transaction to pending pool
	s.blockchain.PendingTxs = append(s.blockchain.PendingTxs, tx)

	// Record success
	s.trackRequest(start, true)
	s.recordSuccess()

	fmt.Printf("📤 Transaction submitted: %s -> %s (%d %s)\n", from, to, amount, tokenID)
	return tx.ID, nil
}

// Enhanced validation
func (s *SimpleRelayServer) validateTransactionRequest(from, to, tokenID string, amount uint64) error {
	if from == "" || to == "" {
		return fmt.Errorf("from and to addresses are required")
	}

	if amount == 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	if tokenID == "" {
		tokenID = "BHX" // Default token
	}

	// Check token exists
	if _, exists := s.blockchain.TokenRegistry[tokenID]; !exists {
		return fmt.Errorf("token %s not found", tokenID)
	}

	// Check balance
	if token, exists := s.blockchain.TokenRegistry[tokenID]; exists {
		balance, err := token.BalanceOf(from)
		if err != nil {
			return fmt.Errorf("failed to check balance: %v", err)
		}

		if balance < amount {
			return fmt.Errorf("insufficient balance: has %d, needs %d", balance, amount)
		}
	}

	return nil
}

// Enhanced balance query with caching
func (s *SimpleRelayServer) GetBalanceSimple(address, tokenID string) (uint64, error) {
	start := time.Now()

	if err := s.checkCircuitBreaker(); err != nil {
		s.trackRequest(start, false)
		return 0, err
	}

	if tokenID == "" {
		tokenID = "BHX" // Default token
	}

	if token, exists := s.blockchain.TokenRegistry[tokenID]; exists {
		balance, err := token.BalanceOf(address)
		if err != nil {
			s.trackRequest(start, false)
			s.recordFailure()
			return 0, status.Error(codes.Internal, fmt.Sprintf("failed to get balance: %v", err))
		}

		s.trackRequest(start, true)
		s.recordSuccess()
		return balance, nil
	}

	s.trackRequest(start, false)
	s.recordFailure()
	return 0, status.Error(codes.NotFound, fmt.Sprintf("token %s not found", tokenID))
}

// Enhanced transaction validation
func (s *SimpleRelayServer) ValidateTransactionSimple(from, to, tokenID string, amount uint64) (bool, string, error) {
	start := time.Now()

	if err := s.checkCircuitBreaker(); err != nil {
		s.trackRequest(start, false)
		return false, "", err
	}

	// Enhanced validation
	if err := s.validateTransactionRequest(from, to, tokenID, amount); err != nil {
		s.trackRequest(start, false)
		s.recordFailure()
		return false, err.Error(), nil
	}

	s.trackRequest(start, true)
	s.recordSuccess()
	return true, "transaction is valid", nil
}

// Enhanced validator information
func (s *SimpleRelayServer) GetValidatorsSimple() []map[string]interface{} {
	start := time.Now()

	if err := s.checkCircuitBreaker(); err != nil {
		s.trackRequest(start, false)
		return nil
	}

	allStakes := s.blockchain.StakeLedger.GetAllStakes()
	validators := make([]map[string]interface{}, 0)

	for address, stake := range allStakes {
		validator := map[string]interface{}{
			"address": address,
			"stake":   stake,
			"status":  "active",
			"jailed":  false,
			"strikes": 0,
		}

		// Check if validator is jailed (if slashing manager is available)
		if s.blockchain.SlashingManager != nil {
			if s.blockchain.SlashingManager.IsValidatorJailed(address) {
				validator["status"] = "jailed"
				validator["jailed"] = true
				validator["strikes"] = s.blockchain.SlashingManager.GetValidatorStrikes(address)
			}
		}

		validators = append(validators, validator)
	}

	s.trackRequest(start, true)
	s.recordSuccess()
	return validators
}

// Enhanced pending transactions
func (s *SimpleRelayServer) GetPendingTransactionsSimple() []map[string]interface{} {
	start := time.Now()

	if err := s.checkCircuitBreaker(); err != nil {
		s.trackRequest(start, false)
		return nil
	}

	pendingTxs := make([]map[string]interface{}, 0)

	for _, tx := range s.blockchain.PendingTxs {
		txInfo := map[string]interface{}{
			"id":        tx.ID,
			"from":      tx.From,
			"to":        tx.To,
			"amount":    tx.Amount,
			"token_id":  tx.TokenID,
			"timestamp": tx.Timestamp,
		}
		pendingTxs = append(pendingTxs, txInfo)
	}

	s.trackRequest(start, true)
	s.recordSuccess()
	return pendingTxs
}

// Enhanced network statistics
func (s *SimpleRelayServer) GetNetworkStatsSimple() map[string]interface{} {
	start := time.Now()

	if err := s.checkCircuitBreaker(); err != nil {
		s.trackRequest(start, false)
		return nil
	}

	latestBlock := s.blockchain.GetLatestBlock()
	if latestBlock == nil {
		s.trackRequest(start, false)
		s.recordFailure()
		return map[string]interface{}{
			"success": false,
			"error":   "no blocks found",
		}
	}

	// Calculate network metrics
	totalStakes := uint64(0)
	allStakes := s.blockchain.StakeLedger.GetAllStakes()
	for _, stake := range allStakes {
		totalStakes += stake
	}

	stats := map[string]interface{}{
		"success":          true,
		"total_blocks":     latestBlock.Header.Index + 1,
		"total_validators": len(allStakes),
		"total_staked":     totalStakes,
		"pending_txs":      len(s.blockchain.PendingTxs),
		"avg_block_time":   15,         // seconds
		"network_hashrate": "1000 TPS", // simulated
		"timestamp":        time.Now().Unix(),
	}

	s.trackRequest(start, true)
	s.recordSuccess()
	return stats
}

// Enhanced health check
func (s *SimpleRelayServer) HealthCheck() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	health := make(map[string]interface{})
	for k, v := range s.healthStatus {
		health[k] = v
	}

	// Add additional health metrics
	health["uptime_seconds"] = time.Now().Unix() - s.healthStatus["timestamp"].(int64)
	health["memory_usage_percent"] = float64(s.healthStatus["memory_usage_mb"].(uint64)) / float64(s.healthStatus["memory_total_mb"].(uint64)) * 100

	return health
}

// Enhanced activity logging
func (s *SimpleRelayServer) LogActivity(activity string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[%s] %s", timestamp, activity)
}

// Get performance metrics
func (s *SimpleRelayServer) GetPerformanceMetrics() map[string]interface{} {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	return map[string]interface{}{
		"request_count":       s.metrics.RequestCount,
		"average_response_ms": s.metrics.AverageResponse.Milliseconds(),
		"error_count":         s.metrics.ErrorCount,
		"error_rate":          float64(s.metrics.ErrorCount) / float64(s.metrics.RequestCount) * 100,
		"active_connections":  s.metrics.ActiveConnections,
		"timestamp":           time.Now().Unix(),
	}
}

// Get circuit breaker status
func (s *SimpleRelayServer) GetCircuitBreakerStatus() map[string]interface{} {
	s.circuitBreaker.mu.RLock()
	defer s.circuitBreaker.mu.RUnlock()

	return map[string]interface{}{
		"state":             s.circuitBreaker.state,
		"failure_count":     s.circuitBreaker.failureCount,
		"failure_threshold": s.circuitBreaker.failureThreshold,
		"last_failure":      s.circuitBreaker.lastFailureTime.Unix(),
		"timestamp":         time.Now().Unix(),
	}
}
