package services

import (
	"fmt"
	"math"

	"github.com/bluesky-o/fairshare/internal/models"
)

func calculateSplits(req *models.CreateExpenseRequest) ([]models.SplitInput, error) {
	if len(req.Splits) == 0 {
		return nil, fmt.Errorf("at least one person must be included in the split")
	}

	switch req.SplitType {
		case "equal":
			return calculateEqualSplit(req.Amount, req.Splits)
		case "exact":
			return calculateExactSplit(req.Amount, req.Splits)
		case "percentage":
			return calculatePercentageSplit(req.Amount, req.Splits)
		default:
			return nil, fmt.Errorf("invalid split_type: must be equal, exact, or percentage")
	}
}

func calculateEqualSplit(totalAmount float64, splits []models.SplitInput) ([]models.SplitInput, error) {
	count := len(splits)

	baseAmount := math.Floor((totalAmount/float64(count))*100) / 100

	result := make([]models.SplitInput, len(splits))
	runningTotal := 0.0

	for i, split := range splits {
		if i == len(splits)-1 {
			lastAmount := math.Round((totalAmount-runningTotal)*100) / 100
			result[i] = models.SplitInput{
				UserID:     split.UserID,
				OwedAmount: lastAmount,
			}
		} else {
			result[i] = models.SplitInput{
				UserID:     split.UserID,
				OwedAmount: baseAmount,
			}
			runningTotal += baseAmount
		}
	}

	return result, nil
}

func calculateExactSplit(totalAmount float64, splits []models.SplitInput) ([]models.SplitInput, error) {
	var splitTotal float64
	for _, split := range splits {
		if split.OwedAmount < 0 {
			return nil, fmt.Errorf("split amounts cannot be negative")
		}
		splitTotal += split.OwedAmount
	}

	if math.Abs(splitTotal-totalAmount) > 0.01 {
		return nil, fmt.Errorf(
			"split amounts (₹%.2f) must add up to total amount (₹%.2f)",
			splitTotal,
			totalAmount,
		)
	}

	return splits, nil
}

func calculatePercentageSplit(totalAmount float64, splits []models.SplitInput) ([]models.SplitInput, error) {
	var totalPercentage float64
	for _, split := range splits {
		if split.Percentage <= 0 {
			return nil, fmt.Errorf("percentages must be greater than 0")
		}
		totalPercentage += split.Percentage
	}

	if math.Abs(totalPercentage-100) > 0.01 {
		return nil, fmt.Errorf(
			"percentages must add up to 100 (got %.2f)",
			totalPercentage,
		)
	}	

	result := make([]models.SplitInput, len(splits))
	runningTotal := 0.0

	for i, split := range splits {
		if i == len(splits)-1 {
			lastAmount := math.Round((totalAmount-runningTotal)*100) / 100
			result[i] = models.SplitInput{
				UserID:     split.UserID,
				OwedAmount: lastAmount,
			}
		} else {
			amount := math.Round((totalAmount*split.Percentage/100)*100) / 100
			result[i] = models.SplitInput{
				UserID:     split.UserID,
				OwedAmount: amount,
			}
			runningTotal += amount
		}
	}

	return result, nil
}
