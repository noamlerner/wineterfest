package stats

import (
	"math"
	"testing"
	"wineterfest/datamodels"
)

func TestRatedTheMostWines(t *testing.T) {
	ratingsByUser := map[string][]*datamodels.WineRating{
		"user1": {{AnonymizedNumber: 1}, {AnonymizedNumber: 2}},
		"user2": {{AnonymizedNumber: 3}, {AnonymizedNumber: 4}, {AnonymizedNumber: 5}},
		"user3": {{AnonymizedNumber: 6}},
	}

	s := &Stats{}
	ratedTheMostWines(s, ratingsByUser)

	if s.RatedTheMostWines.Name != "user2" || s.RatedTheMostWines.Value != 3 {
		t.Errorf("Expected user2 with 3 ratings, got %s with %f", s.RatedTheMostWines.Name, s.RatedTheMostWines.Value)
	}
}

func TestWineAverages(t *testing.T) {
	ratingsByWine := map[int][]*datamodels.WineRating{
		1: {{AnonymizedNumber: 1, Rating: 4}, {AnonymizedNumber: 1, Rating: 5}},
		2: {{AnonymizedNumber: 2, Rating: 3}},
	}

	numToWine := map[int]*datamodels.Wine{
		1: {WineName: "Wine A", WinePrice: 10},
		2: {WineName: "Wine B", WinePrice: 20},
	}

	s := &Stats{}
	wineAverages(s, ratingsByWine, numToWine)

	if len(s.WineRanking) != 2 || s.WineRanking[0].Name.WineName != "Wine A" {
		t.Errorf("Expected Wine A to rank first, got %s", s.WineRanking[0].Name.WineName)
	}

	if len(s.WineValue) != 2 || s.WineValue[0].Name.WineName != "Wine A" {
		t.Errorf("Expected Wine A to have the best value, got %s", s.WineValue[0].Name.WineName)
	}
}

func TestUserCorrelationCoefficient(t *testing.T) {
	ratingsByUser := map[string][]*datamodels.WineRating{
		"user1": {
			{AnonymizedNumber: 1, PriceGuess: 15},
			{AnonymizedNumber: 2, PriceGuess: 20},
			{AnonymizedNumber: 3, PriceGuess: 30},
			{AnonymizedNumber: 4, PriceGuess: 50},
		},
		"user2": {
			{AnonymizedNumber: 1, PriceGuess: 10},
			{AnonymizedNumber: 2, PriceGuess: 10},
		},
		"user3": {
			{AnonymizedNumber: 1, PriceGuess: 20},
			{AnonymizedNumber: 2, PriceGuess: 15},
			{AnonymizedNumber: 3, PriceGuess: 23},
			{AnonymizedNumber: 4, PriceGuess: 11},
		},
	}

	numToWine := map[int]*datamodels.Wine{
		1: {WineName: "Wine A", WinePrice: 15},
		2: {WineName: "Wine B", WinePrice: 20},
		3: {WineName: "Wine C", WinePrice: 30},
		4: {WineName: "Wine D", WinePrice: 40},
	}

	s := &Stats{}
	userCorrelationCoefficient(s, ratingsByUser, numToWine)
}

func TestCalculateCorrelation(t *testing.T) {
	guesses := []float64{10, 20, 30}
	actualPrices := []float64{10, 20, 30}

	correlation := calculateCorrelation(guesses, actualPrices)
	if correlation == nil || math.Abs(*correlation-1) > 0.01 {
		t.Errorf("Expected correlation of 1, got %v", correlation)
	}

	guesses = []float64{10, 20, 30}
	actualPrices = []float64{30, 20, 10}

	correlation = calculateCorrelation(guesses, actualPrices)
	if correlation == nil || math.Abs(*correlation+1) > 0.01 {
		t.Errorf("Expected correlation of -1, got %v", correlation)
	}
}

func TestMean(t *testing.T) {
	data := []float64{10, 20, 30}
	expected := 20.0

	result := mean(data)
	if math.Abs(result-expected) > 0.01 {
		t.Errorf("Expected mean of 20, got %f", result)
	}
}
