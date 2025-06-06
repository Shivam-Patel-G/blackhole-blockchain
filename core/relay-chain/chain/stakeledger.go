package chain

import (
	"errors"
	"fmt"
	"sync"
	"time"
	
	"github.com/Shivam-Patel-G/blackhole-blockchain/core/relay-chain/token"
)

type StakeLedger struct {
	Stakes map[string]uint64
	mu     sync.RWMutex
}

func NewStakeLedger() *StakeLedger {
	sl := &StakeLedger{
		Stakes: make(map[string]uint64),
	}
	sl.InitializeDefaultStakes()
	return sl
}

func (sl *StakeLedger) ToMap() map[string]uint64 {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	stakes := make(map[string]uint64)
	for addr, stake := range sl.Stakes {
		stakes[addr] = stake
	}
	return stakes
}

func (sl *StakeLedger) GetStake(address string) uint64 {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.Stakes[address]
}

func (sl *StakeLedger) SetStake(address string, stake uint64) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.Stakes[address] = stake
}

func (sl *StakeLedger) AddStake(address string, amount uint64) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.Stakes[address] += amount
}

func (sl *StakeLedger) GetAllStakes() map[string]uint64 {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	stakes := make(map[string]uint64)
	for addr, stake := range sl.Stakes {
		stakes[addr] = stake
	}
	return stakes
}

func (sl *StakeLedger) InitializeDefaultStakes() {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	// Set default stakes for testing
	sl.Stakes["node1"] = 1000 // First node (port 3000)
	sl.Stakes["node2"] = 500  // Second node (port 3001)
}

func (sl *StakeLedger) IsSelectedValidator(address string) bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	// Simple PoS: Select validator with highest stake
	// Deterministic based on current time slot for coordination
	maxStake := uint64(0)
	maxAddr := ""
	for addr, stake := range sl.Stakes {
		if stake > maxStake {
			maxStake = stake
			maxAddr = addr
		}
	}

	// Ensure the address has stake and is the highest
	if maxStake == 0 || address != maxAddr {
		return false
	}

	// Time-based slot to mimic PoS slot selection
	slotDuration := 5 * time.Second
	currentSlot := uint64(time.Now().UTC().UnixNano()) / uint64(slotDuration.Nanoseconds())
	fmt.Println("Current Slot:", currentSlot)
	// Use slot to ensure only one validator per slot (simplified)
	return true
}

func (sl *StakeLedger) GetHighestStakeValidator() string {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	maxStake := uint64(0)
	maxAddr := ""
	for addr, stake := range sl.Stakes {
		if stake > maxStake {
			maxStake = stake
			maxAddr = addr
		}
	}
	return maxAddr
}

// Integrate with token system
func (sl *StakeLedger) StakeTokens(address string, amount uint64, tokenSystem *token.Token) error {
	// Verify token balance
	balance, err := tokenSystem.BalanceOf(address)
	if err != nil {
		return err
	}
	if balance < amount {
		return errors.New("insufficient token balance for staking")
	}
	
	// Lock tokens (could implement as internal transfer to staking contract address)
	err = tokenSystem.Transfer(address, "staking_contract", amount)
	if err != nil {
		return err
	}
	
	// Add stake
	sl.AddStake(address, amount)
	return nil
}
