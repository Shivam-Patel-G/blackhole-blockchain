package pool

import (
	"errors"
	"math"
)

type LiquidityPool struct {
	TokenA       string
	TokenB       string
	ReserveA     float64
	ReserveB     float64
	TotalShare   float64
	FeeReserveA  float64 
	FeeReserveB  float64 
	ProtocolShare  float64
}


// CreatePool creates a new pool with zero reserves and shares
func CreatePool(tokenA, tokenB string) (*LiquidityPool, error) {
	if tokenA == "" || tokenB == "" || tokenA == tokenB {
		return nil, errors.New("invalid token pair")
	}

	return &LiquidityPool{
		TokenA:     tokenA,
		TokenB:     tokenB,
		ReserveA:   0,
		ReserveB:   0,
		TotalShare: 0,
	}, nil
}

// AddLiquidity lets a user add tokens to the pool
func (p *LiquidityPool) AddLiquidity(amountA, amountB float64) (float64, error) {
	if amountA <= 0 || amountB <= 0 {
		return 0, errors.New("token amounts must be positive")
	}

	// First liquidity provider (bootstrap)
	if p.TotalShare == 0 {
		share := math.Sqrt(amountA * amountB)
		p.ReserveA += amountA
		p.ReserveB += amountB
		p.TotalShare += share
		return share, nil
	}

	// Ensure correct ratio
	if err := validateRatio(amountA/p.ReserveA, amountB/p.ReserveB); err != nil {
		return 0, err
	}

	// Calculate shares based on existing liquidity
	shareA := (amountA * p.TotalShare) / p.ReserveA
	shareB := (amountB * p.TotalShare) / p.ReserveB
	shares := math.Min(shareA, shareB)

	p.ReserveA += amountA
	p.ReserveB += amountB
	p.TotalShare += shares

	return shares, nil
}

const SwapFeePercent = 0.003 // 0.3% fee

// ExecuteSwap executes a swap transaction
func (p *LiquidityPool) ExecuteSwap(fromToken string, amountIn, minExpected float64) (float64, error) {
	if amountIn <= 0 {
		return 0, errors.New("input amount must be positive")
	}

	fee := amountIn * SwapFeePercent
	amountInWithFee := amountIn - fee

	var amountOut float64

	switch fromToken {
	case p.TokenA:
		if p.ReserveA <= 0 || p.ReserveB <= 0 {
			return 0, errors.New("invalid pool state")
		}
		amountOut = (amountInWithFee * p.ReserveB) / (p.ReserveA + amountInWithFee)
		if amountOut > p.ReserveB {
			return 0, errors.New("not enough liquidity for output token")
		}
		if amountOut < minExpected {
			return 0, errors.New("slippage too high")
		}

		// ✅ Update reserves
		p.ReserveA += amountInWithFee
		p.ReserveB -= amountOut
		p.FeeReserveA += fee // ✅ Accumulate fee

	case p.TokenB:
		if p.ReserveB <= 0 || p.ReserveA <= 0 {
			return 0, errors.New("invalid pool state")
		}
		amountOut = (amountInWithFee * p.ReserveA) / (p.ReserveB + amountInWithFee)
		if amountOut > p.ReserveA {
			return 0, errors.New("not enough liquidity for output token")
		}
		if amountOut < minExpected {
			return 0, errors.New("slippage too high")
		}

		// ✅ Update reserves
		p.ReserveB += amountInWithFee
		p.ReserveA -= amountOut
		p.FeeReserveB += fee // ✅ Accumulate fee

	default:
		return 0, errors.New("invalid fromToken")
	}

	// Optional: Round everything cleanly
	p.ReserveA = roundToPrecision(p.ReserveA)
	p.ReserveB = roundToPrecision(p.ReserveB)
	p.FeeReserveA = roundToPrecision(p.FeeReserveA)
	p.FeeReserveB = roundToPrecision(p.FeeReserveB)

	return roundToPrecision(amountOut), nil
}


func roundToPrecision(value float64) float64 {
    return math.Round(value*1e12) / 1e12
}

// RemoveLiquidity allows a user to redeem their share of the pool's reserves
func (p *LiquidityPool) RemoveLiquidity(shareAmount float64) (amountA float64, amountB float64, err error) {
    // === Step 1: Validation ===
    if shareAmount <= 0 {
        return 0, 0, errors.New("share amount must be positive")
    }
    if p.TotalShare == 0 {
        return 0, 0, errors.New("pool is empty")
    }
    if shareAmount > p.TotalShare {
        return 0, 0, errors.New("insufficient pool shares")
    }

    // === Step 2: Calculate user's proportional reserves ===
    shareRatio := shareAmount / p.TotalShare

    // Calculate amounts without rounding first
    amountA = p.ReserveA * shareRatio
    amountB = p.ReserveB * shareRatio

    // === Step 3: Update state ===
    p.ReserveA -= amountA
    p.ReserveB -= amountB
    p.TotalShare -= shareAmount

    // If all shares removed, zero out residuals (fix floating point inaccuracies)
    if p.TotalShare <= 1e-12 {
        p.TotalShare = 0
    }
    if p.ReserveA <= 1e-12 {
        p.ReserveA = 0
    }
    if p.ReserveB <= 1e-12 {
        p.ReserveB = 0
    }

    // Round only the returned amounts for cleaner output
    amountA = roundToPrecision(amountA)
    amountB = roundToPrecision(amountB)
    p.TotalShare = roundToPrecision(p.TotalShare)
    p.ReserveA = roundToPrecision(p.ReserveA)
    p.ReserveB = roundToPrecision(p.ReserveB)

    // === Step 4: Final sanity check (optional safeguard) ===
    if p.ReserveA < 0 || p.ReserveB < 0 || p.TotalShare < 0 {
        return 0, 0, errors.New("internal accounting error")
    }

    return amountA, amountB, nil
}

// DistributeFees allocates accumulated swap fees between LPs and the protocol.
func (p *LiquidityPool) DistributeFees() (protocolA, protocolB float64, err error) {
	// === Step 1: Sanity Check ===
	if p.FeeReserveA < 0 || p.FeeReserveB < 0 {
		return 0, 0, errors.New("invalid fee reserve state")
	}
	if p.ProtocolShare < 0 || p.ProtocolShare > 1 {
		return 0, 0, errors.New("protocol share must be between 0 and 1")
	}

	// === Step 2: Compute protocol and LP shares ===
	protocolA = roundToPrecision(p.FeeReserveA * p.ProtocolShare)
	protocolB = roundToPrecision(p.FeeReserveB * p.ProtocolShare)

	lpShareA := roundToPrecision(p.FeeReserveA - protocolA)
	lpShareB := roundToPrecision(p.FeeReserveB - protocolB)

	// === Step 3: Update state ===
	p.ReserveA = roundToPrecision(p.ReserveA + lpShareA)
	p.ReserveB = roundToPrecision(p.ReserveB + lpShareB)

	p.FeeReserveA = 0
	p.FeeReserveB = 0

	// Optionally: You could track protocolA/B amounts in a treasury wallet
	return protocolA, protocolB, nil
}



// validateRatio checks if the added tokens are in the right ratio to the reserves
func validateRatio(ratioA, ratioB float64) error {
	const tolerance = 0.01
	if math.Abs(ratioA-ratioB) > tolerance {
		return errors.New("token amounts not in correct ratio")
	}
	return nil
}

// GetPoolStatus returns current pool state
func (p *LiquidityPool) GetPoolStatus() map[string]interface{} {
	return map[string]interface{}{
		"TokenA":       p.TokenA,
		"TokenB":       p.TokenB,
		"ReserveA":     p.ReserveA,
		"ReserveB":     p.ReserveB,
		"TotalShare":   p.TotalShare,
		"FeeReserveA":  p.FeeReserveA, 
		"FeeReserveB":  p.FeeReserveB, 
	}
}

