package swap

import (
    "errors"
)

type SwapRateResult struct {
    FromAmount       float64
    ToAmount         float64
    FromToken        string
    ToToken          string
    AmountAfterFee   float64
    FeePercentage    float64
}

type QuoteResult struct {
    FromAmount float64 // input amount
    ToAmount   float64 // calculated output
    FromToken  string  // e.g., "USDC"
    ToToken    string  // e.g., "ETH"
}



func GetSwapRate(amountIn, reserveA, reserveB float64, fromToken, toToken string) (*SwapRateResult, error) {
    if amountIn <= 0 {
        return nil, errors.New("input amount must be positive")
    }
    if reserveA <= 0 || reserveB <= 0 {
        return nil, errors.New("invalid pool reserves")
    }

    feePercent := 0.003
    amountInWithFee := amountIn * (1 - feePercent)
    numerator := amountInWithFee * reserveB
    denominator := reserveA + amountInWithFee
    amountOut := numerator / denominator

    return &SwapRateResult{
        FromAmount:     amountIn,
        ToAmount:       amountOut,
        FromToken:      fromToken,
        ToToken:        toToken,
        AmountAfterFee: amountInWithFee,
        FeePercentage:  feePercent * 100,
    }, nil
}


func Quote(fromAmount, reserveA, reserveB float64, fromToken, toToken string) (*QuoteResult, error) {
    if fromAmount <= 0 {
        return nil, errors.New("input amount must be positive")
    }
    if reserveA <= 0 || reserveB <= 0 {
        return nil, errors.New("invalid pool reserves")
    }

    toAmount := (fromAmount * reserveB) / reserveA

    result := &QuoteResult{
        FromAmount: fromAmount,
        ToAmount:   toAmount,
        FromToken:  fromToken,
        ToToken:    toToken,
    }

    return result, nil
}

func CalculatePriceImpact(amountIn, reserveA, reserveB float64) (float64, error) {
    if amountIn <= 0 {
        return 0, errors.New("input amount must be positive")
    }
    if reserveA <= 0 || reserveB <= 0 {
        return 0, errors.New("invalid pool reserves")
    }


    priceBefore := reserveB / reserveA

    
    swapResult, err := GetSwapRate(amountIn, reserveA, reserveB, "", "")
    if err != nil {
        return 0, err
    }
    amountOut := swapResult.ToAmount

    
    priceAfter := (reserveB - amountOut) / (reserveA + amountIn)

    
    priceImpact := ((priceBefore - priceAfter) / priceBefore) * 100

    return priceImpact, nil
}