package pool

import (
	"testing"
	"math"
)

func TestNewPool(t *testing.T) {
	tests := []struct {
		name    string
		tokenA  string
		tokenB  string
		wantErr bool
	}{
		{"Valid tokens", "BLH", "USDT", false},
		{"Empty token", "", "USDT", true},
		{"Same token", "USDT", "USDT", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreatePool(tt.tokenA, tt.tokenB)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddLiquidity(t *testing.T) {
	pool, _ := CreatePool("BLH", "USDT")

	t.Run("Bootstrap liquidity", func(t *testing.T) {
		shares, err := pool.AddLiquidity(100, 100)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if shares <= 0 {
			t.Errorf("expected shares > 0, got %f", shares)
		}
	})

	t.Run("Correct ratio", func(t *testing.T) {
		shares, err := pool.AddLiquidity(50, 50)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if shares <= 0 {
			t.Errorf("expected shares > 0, got %f", shares)
		}
	})

	t.Run("Incorrect ratio", func(t *testing.T) {
		_, err := pool.AddLiquidity(100, 10) // Bad ratio
		if err == nil {
			t.Errorf("expected ratio error, got nil")
		}
	})
}

func TestValidateRatio(t *testing.T) {
	tests := []struct {
		name    string
		ratioA  float64
		ratioB  float64
		wantErr bool
	}{
		{"Exact match", 1.0, 1.0, false},
		{"Within tolerance", 1.0, 1.009, false},
		{"Over tolerance", 1.0, 1.02, true},
		{"Big mismatch", 1.0, 2.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRatio(tt.ratioA, tt.ratioB)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRatio() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPoolStatus(t *testing.T) {
	pool, _ := CreatePool("BLH", "USDT")
	_, _ = pool.AddLiquidity(500, 1000)

	status := pool.GetPoolStatus()

	if tokenA, ok := status["TokenA"].(string); !ok || tokenA != "BLH" {
		t.Errorf("Expected TokenA to be 'BLH', got %v", status["TokenA"])
	}
	if tokenB, ok := status["TokenB"].(string); !ok || tokenB != "USDT" {
		t.Errorf("Expected TokenB to be 'USDT', got %v", status["TokenB"])
	}
	if reserveA, ok := status["ReserveA"].(float64); !ok || reserveA != 500.0 {
		t.Errorf("Expected ReserveA to be 500, got %v", status["ReserveA"])
	}
	if reserveB, ok := status["ReserveB"].(float64); !ok || reserveB != 1000.0 {
		t.Errorf("Expected ReserveB to be 1000, got %v", status["ReserveB"])
	}
	if totalShare, ok := status["TotalShare"].(float64); !ok || totalShare <= 0 {
		t.Errorf("Expected TotalShare to be > 0, got %v", status["TotalShare"])
	}
}


func TestExecuteSwap(t *testing.T) {
	pool, _ := CreatePool("BLH", "USDT")
	_, _ = pool.AddLiquidity(1000, 2000) // BLH:USDT ratio is 1:2

	t.Run("Swap BLH for USDT", func(t *testing.T) {
		fromToken := "BLH"
		amountIn := 100.0
		minExpected := 180.0 // set slippage guard low for test

		out, err := pool.ExecuteSwap(fromToken, amountIn, minExpected)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out <= 0 {
			t.Errorf("expected output > 0, got %f", out)
		}
	})

	t.Run("Swap USDT for BLH", func(t *testing.T) {
		fromToken := "USDT"
		amountIn := 200.0
		minExpected := 90.0

		out, err := pool.ExecuteSwap(fromToken, amountIn, minExpected)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out <= 0 {
			t.Errorf("expected output > 0, got %f", out)
		}
	})

	t.Run("Invalid token swap", func(t *testing.T) {
		_, err := pool.ExecuteSwap("INVALID", 100, 0)
		if err == nil {
			t.Errorf("expected error for invalid token")
		}
	})

	t.Run("Slippage too high", func(t *testing.T) {
		// Intentionally set minExpected to a high value
		_, err := pool.ExecuteSwap("BLH", 10, 1000)
		if err == nil {
			t.Errorf("expected slippage error, got nil")
		}
	})

	t.Run("Zero input", func(t *testing.T) {
		_, err := pool.ExecuteSwap("BLH", 0, 0)
		if err == nil {
			t.Errorf("expected error for zero input")
		}
	})
}

func TestRemoveLiquidity(t *testing.T) {
    const epsilon = 1e-9

    // Helper to check float equality within epsilon tolerance
    floatsEqual := func(a, b float64) bool {
        return math.Abs(a-b) < epsilon
    }

    t.Run("Invalid share amount zero or negative", func(t *testing.T) {
        pool := &LiquidityPool{TokenA: "A", TokenB: "B", ReserveA: 1000, ReserveB: 2000, TotalShare: 100}
        _, _, err := pool.RemoveLiquidity(0)
        if err == nil {
            t.Error("expected error for zero share amount")
        }
        _, _, err = pool.RemoveLiquidity(-10)
        if err == nil {
            t.Error("expected error for negative share amount")
        }
    })

    t.Run("Empty pool", func(t *testing.T) {
        pool := &LiquidityPool{TokenA: "A", TokenB: "B"}
        _, _, err := pool.RemoveLiquidity(10)
        if err == nil {
            t.Error("expected error for empty pool")
        }
    })

    t.Run("Insufficient pool shares", func(t *testing.T) {
        pool := &LiquidityPool{TokenA: "A", TokenB: "B", ReserveA: 1000, ReserveB: 2000, TotalShare: 100}
        _, _, err := pool.RemoveLiquidity(200)
        if err == nil {
            t.Error("expected error for removing more shares than total")
        }
    })

    t.Run("Partial removal", func(t *testing.T) {
        pool := &LiquidityPool{TokenA: "A", TokenB: "B", ReserveA: 1000, ReserveB: 2000, TotalShare: 100}
        shareToRemove := 25.0 // 25% shares
        amountA, amountB, err := pool.RemoveLiquidity(shareToRemove)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        expectedA := 1000 * (shareToRemove / 100)
        expectedB := 2000 * (shareToRemove / 100)

        if !floatsEqual(amountA, expectedA) {
            t.Errorf("expected amountA %.9f, got %.9f", expectedA, amountA)
        }
        if !floatsEqual(amountB, expectedB) {
            t.Errorf("expected amountB %.9f, got %.9f", expectedB, amountB)
        }

        expectedReserveA := 1000 - expectedA
        expectedReserveB := 2000 - expectedB
        expectedTotalShare := 100 - shareToRemove

        if !floatsEqual(pool.ReserveA, expectedReserveA) {
            t.Errorf("expected pool reserveA %.9f, got %.9f", expectedReserveA, pool.ReserveA)
        }
        if !floatsEqual(pool.ReserveB, expectedReserveB) {
            t.Errorf("expected pool reserveB %.9f, got %.9f", expectedReserveB, pool.ReserveB)
        }
        if !floatsEqual(pool.TotalShare, expectedTotalShare) {
            t.Errorf("expected total share %.9f, got %.9f", expectedTotalShare, pool.TotalShare)
        }
    })

    t.Run("Remove all liquidity", func(t *testing.T) {
        pool := &LiquidityPool{TokenA: "A", TokenB: "B", ReserveA: 500, ReserveB: 1000, TotalShare: 50}
        amountA, amountB, err := pool.RemoveLiquidity(50) // remove 100% shares
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        if !floatsEqual(amountA, 500) {
            t.Errorf("expected amountA 500, got %.9f", amountA)
        }
        if !floatsEqual(amountB, 1000) {
            t.Errorf("expected amountB 1000, got %.9f", amountB)
        }

        // After full removal, reserves and total share should be zero
        if pool.ReserveA != 0 {
            t.Errorf("expected pool reserveA 0 after full removal, got %.9f", pool.ReserveA)
        }
        if pool.ReserveB != 0 {
            t.Errorf("expected pool reserveB 0 after full removal, got %.9f", pool.ReserveB)
        }
        if pool.TotalShare != 0 {
            t.Errorf("expected total share 0 after full removal, got %.9f", pool.TotalShare)
        }
    })

    t.Run("Floating point precision boundary", func(t *testing.T) {
        pool := &LiquidityPool{TokenA: "A", TokenB: "B", ReserveA: 1e-12, ReserveB: 1e-12, TotalShare: 1e-12}
        amountA, amountB, err := pool.RemoveLiquidity(1e-12)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if amountA != 1e-12 || amountB != 1e-12 {
            t.Errorf("expected amounts close to 1e-12, got %g and %g", amountA, amountB)
        }
        if pool.ReserveA != 0 || pool.ReserveB != 0 || pool.TotalShare != 0 {
            t.Error("expected pool state to zero out small float residues")
        }
    })
}

func almostEqual(a, b float64) bool {
	const eps = 1e-9
	return math.Abs(a-b) < eps
}

func TestExecuteSwapWithFees(t *testing.T) {
	pool, err := CreatePool("USDC", "ETH")
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	_, err = pool.AddLiquidity(1000, 10)
	if err != nil {
		t.Fatalf("failed to add liquidity: %v", err)
	}

	t.Run("Swap USDC to ETH", func(t *testing.T) {
		amountIn := 100.0
		minExpected := 0.5
		initialFeeA := pool.FeeReserveA

		amountOut, err := pool.ExecuteSwap("USDC", amountIn, minExpected)
		if err != nil {
			t.Fatalf("swap failed: %v", err)
		}
		if amountOut <= 0 {
			t.Fatal("expected positive output")
		}

		expectedFee := roundToPrecision(amountIn * SwapFeePercent)
		if !almostEqual(pool.FeeReserveA, initialFeeA+expectedFee) {
			t.Errorf("expected FeeReserveA to increase by %.6f, got %.6f", expectedFee, pool.FeeReserveA-initialFeeA)
		}

		expectedReserveA := roundToPrecision(1000 + (amountIn - expectedFee))
		if !almostEqual(pool.ReserveA, expectedReserveA) {
			t.Errorf("unexpected ReserveA: got %.6f, want %.6f", pool.ReserveA, expectedReserveA)
		}
	})

	t.Run("Swap ETH to USDC", func(t *testing.T) {
		amountIn := 1.0
		minExpected := 50.0
		initialFeeB := pool.FeeReserveB

		amountOut, err := pool.ExecuteSwap("ETH", amountIn, minExpected)
		if err != nil {
			t.Fatalf("swap failed: %v", err)
		}
		if amountOut <= 0 {
			t.Fatal("expected positive output")
		}

		expectedFee := roundToPrecision(amountIn * SwapFeePercent)
		if !almostEqual(pool.FeeReserveB, initialFeeB+expectedFee) {
			t.Errorf("expected FeeReserveB to increase by %.6f, got %.6f", expectedFee, pool.FeeReserveB-initialFeeB)
		}
	})

	t.Run("Invalid token", func(t *testing.T) {
		_, err := pool.ExecuteSwap("INVALID", 50, 1)
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("Zero input", func(t *testing.T) {
		_, err := pool.ExecuteSwap("USDC", 0, 0.1)
		if err == nil {
			t.Error("expected error for zero input")
		}
	})
}


func TestDistributeFees(t *testing.T) {
	t.Run("DistributeFees with non-zero fee reserves", func(t *testing.T) {
		pool := &LiquidityPool{
			TokenA:      "ETH",
			TokenB:      "USDC",
			ReserveA:    1000,
			ReserveB:    2000,
			TotalShare:  100,
			FeeReserveA: 10,
			FeeReserveB: 20,
		}

		aFee, bFee, err := pool.DistributeFees()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if aFee != 10 || bFee != 20 {
			t.Errorf("expected distributed fees (10, 20), got (%v, %v)", aFee, bFee)
		}
		if pool.FeeReserveA != 0 || pool.FeeReserveB != 0 {
			t.Errorf("fee reserves not cleared after distribution")
		}
		if pool.ReserveA != 1010 || pool.ReserveB != 2020 {
			t.Errorf("reserves not updated correctly after distribution")
		}
	})

	t.Run("DistributeFees with zero fee reserves", func(t *testing.T) {
		pool := &LiquidityPool{
			TokenA:      "ETH",
			TokenB:      "USDC",
			ReserveA:    1000,
			ReserveB:    1000,
			TotalShare:  100,
			FeeReserveA: 0,
			FeeReserveB: 0,
		}

		aFee, bFee, err := pool.DistributeFees()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if aFee != 0 || bFee != 0 {
			t.Errorf("expected 0 distributed fees, got (%v, %v)", aFee, bFee)
		}
		if pool.ReserveA != 1000 || pool.ReserveB != 1000 {
			t.Errorf("reserves changed unexpectedly")
		}
	})

	t.Run("DistributeFees with floating-point precision", func(t *testing.T) {
		pool := &LiquidityPool{
			TokenA:      "ETH",
			TokenB:      "USDC",
			ReserveA:    500,
			ReserveB:    800,
			TotalShare:  80,
			FeeReserveA: 0.0000001234,
			FeeReserveB: 0.0000005678,
		}

		aFee, bFee, err := pool.DistributeFees()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if math.Abs(pool.FeeReserveA) > 1e-12 || math.Abs(pool.FeeReserveB) > 1e-12 {
			t.Errorf("fee reserves not cleared (residuals remain)")
		}
		if aFee <= 0 || bFee <= 0 {
			t.Errorf("expected tiny distributed fees, got %v and %v", aFee, bFee)
		}
	})
}
