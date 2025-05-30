package swap

import (
    "testing"
)

func TestGetSwapRate(t *testing.T) {
    amountIn := 10.0
    reserveA := 1000.0
    reserveB := 5000.0
    fromToken := "USDC"
    toToken := "ETH"

    result, err := GetSwapRate(amountIn, reserveA, reserveB, fromToken, toToken)
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }

    if result == nil {
        t.Fatal("Expected result, got nil")
    }

    if result.FromAmount != amountIn {
        t.Errorf("Expected FromAmount %v, got %v", amountIn, result.FromAmount)
    }

    if result.FromToken != fromToken || result.ToToken != toToken {
        t.Errorf("Token names mismatch. Got FromToken: %s, ToToken: %s", result.FromToken, result.ToToken)
    }

    if result.ToAmount <= 0 {
        t.Errorf("Expected positive ToAmount, got %v", result.ToAmount)
    }

    expectedFee := amountIn * 0.003
    expectedAmountAfterFee := amountIn - expectedFee
    if result.AmountAfterFee != expectedAmountAfterFee {
        t.Errorf("Expected AmountAfterFee %v, got %v", expectedAmountAfterFee, result.AmountAfterFee)
    }
}


func TestQuote(t *testing.T) {
    result, err := Quote(10, 1000, 5000, "USDC", "ETH")
    if err != nil {
        t.Fatalf("Quote failed: %v", err)
    }

    expected := 50.0
    if result.ToAmount != expected {
        t.Errorf("Expected %v, got %v", expected, result.ToAmount)
    }

    t.Logf("Quoted %v %s -> %v %s", result.FromAmount, result.FromToken, result.ToAmount, result.ToToken)
}


func TestCalculatePriceImpact(t *testing.T) {
    reserveA := 1000.0
    reserveB := 5000.0
    amountIn := 10.0

    priceImpact, err := CalculatePriceImpact(amountIn, reserveA, reserveB)
    if err != nil {
        t.Fatalf("CalculatePriceImpact returned error: %v", err)
    }

    if priceImpact <= 0 {
        t.Errorf("Expected positive price impact, got %v", priceImpact)
    }

    // Test invalid inputs
    _, err = CalculatePriceImpact(-5, reserveA, reserveB)
    if err == nil {
        t.Error("Expected error for negative amountIn, got nil")
    }

    _, err = CalculatePriceImpact(amountIn, -100, reserveB)
    if err == nil {
        t.Error("Expected error for negative reserveA, got nil")
    }

    _, err = CalculatePriceImpact(amountIn, reserveA, 0)
    if err == nil {
        t.Error("Expected error for zero reserveB, got nil")
    }
}

