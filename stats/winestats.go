package stats

import (
	"fmt"
	"math"
	"sort"
	"wineterfest/datamodels"
)

type Stat[T any] struct {
	Name  T
	Value float64
}

type Stats struct {
	// User -> Number of Ratings
	RatedTheMostWines Stat[[]string]
	// User -> Average Wine Rating
	AverageWineRatings []Stat[string]
	// WineName -> Average Rating
	WineRanking []Stat[datamodels.Wine]
	// WineName -> Value
	WineValue []Stat[datamodels.Wine]
	// User -> PriceGuessCorrelation
	PriceGuessingCorrelations []Stat[string]
}
type JsonStats struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Items       []string `json:"items"`
}

func Calc(allWines []datamodels.Wine, allRatings []datamodels.WineRating) *Stats {
	s := &Stats{}

	ratingsByUser := make(map[string][]datamodels.WineRating)
	ratingsByWine := make(map[int][]datamodels.WineRating)
	for _, rating := range allRatings {
		if rating.AnonymizedNumber == -1 {
			continue
		}
		ratingsByUser[rating.WineUser] = append(ratingsByUser[rating.WineUser], rating)
		ratingsByWine[rating.AnonymizedNumber] = append(ratingsByWine[rating.AnonymizedNumber], rating)
	}
	numToWine := make(map[int]datamodels.Wine)
	for _, wine := range allWines {
		numToWine[wine.AnonymizedNumber] = wine
	}

	averageRatings(ratingsByUser, s)
	ratedTheMostWines(s, ratingsByUser)
	wineAverages(s, ratingsByWine, numToWine)
	userCorrelationCoefficient(s, ratingsByUser, numToWine)
	return s
}

func averageRatings(ratingsByUser map[string][]datamodels.WineRating, s *Stats) {
	averageRatingsPerUser := make([]Stat[string], 0, len(ratingsByUser))
	for user, ratings := range ratingsByUser {
		if len(ratings) == 0 {
			continue
		}
		sum := 0.0
		for _, rating := range ratings {
			sum += float64(rating.Rating)
		}
		averageRatingsPerUser = append(averageRatingsPerUser, Stat[string]{
			Name:  user,
			Value: sum / float64(len(ratingsByUser[user])),
		})
	}
	s.AverageWineRatings = averageRatingsPerUser
}

func wineAverages(s *Stats, ratingsByWine map[int][]datamodels.WineRating, numToWine map[int]datamodels.Wine) {
	wineRankings := make([]Stat[datamodels.Wine], 0, len(ratingsByWine))
	wineValues := make([]Stat[datamodels.Wine], 0, len(ratingsByWine))
	for num, ratings := range ratingsByWine {
		if numToWine[num].WinePrice == 0.0 {
			continue
		}
		sum := 0.0
		for _, rating := range ratings {
			sum += float64(rating.AnonymizedNumber)
		}
		wineRankings = append(wineRankings, Stat[datamodels.Wine]{
			Name:  numToWine[num],
			Value: sum,
		})
		wineValues = append(wineValues, Stat[datamodels.Wine]{
			Name:  numToWine[num],
			Value: sum / numToWine[num].WinePrice,
		})
	}
	sort.Slice(wineRankings, func(i, j int) bool {
		return wineRankings[i].Value > wineRankings[j].Value
	})
	s.WineRanking = wineRankings

	sort.Slice(s.WineValue, func(i, j int) bool {
		return s.WineValue[i].Value > s.WineValue[j].Value
	})
	s.WineValue = wineValues
}

func ratedTheMostWines(s *Stats, ratingsByUser map[string][]datamodels.WineRating) {
	stat := Stat[[]string]{}
	for user, ratings := range ratingsByUser {
		if len(ratings) > int(stat.Value) {
			stat.Name = []string{user}
			stat.Value = float64(len(ratings))
		} else if len(ratings) == int(stat.Value) {
			stat.Name = append(stat.Name, user)
		}
	}
	s.RatedTheMostWines = stat
}

func userCorrelationCoefficient(s *Stats, ratingsByUser map[string][]datamodels.WineRating, numToWine map[int]datamodels.Wine) {
	correlations := make([]Stat[string], 0, len(ratingsByUser))
	for user, ratings := range ratingsByUser {
		guesses := make([]float64, 0, len(ratings))
		actualPrices := make([]float64, 0, len(ratings))
		for _, rating := range ratings {
			guesses = append(guesses, float64(rating.PriceGuess))
			actualPrices = append(actualPrices, numToWine[rating.AnonymizedNumber].WinePrice)
		}
		correlation := calculateCorrelation(guesses, actualPrices)
		if correlation == nil {
			continue
		}
		correlations = append(correlations, Stat[string]{
			Name:  user,
			Value: *correlation,
		})
	}
	sort.Slice(correlations, func(i, j int) bool {
		return correlations[i].Value > correlations[j].Value
	})

	s.PriceGuessingCorrelations = correlations
}

func calculateCorrelation(guesses, actualPrices []float64) *float64 {
	if len(guesses) != len(actualPrices) || len(guesses) == 0 {
		return nil
	}
	// Calculate means
	meanGuess := mean(guesses)
	meanActual := mean(actualPrices)

	// Calculate covariance and variances
	var covariance, varianceGuess, varianceActual float64
	for i := range guesses {
		guessDiff := guesses[i] - meanGuess
		actualDiff := actualPrices[i] - meanActual
		covariance += guessDiff * actualDiff
		varianceGuess += guessDiff * guessDiff
		varianceActual += actualDiff * actualDiff
	}

	// Calculate Pearson correlation
	if varianceGuess == 0 || varianceActual == 0 {
		return nil
	}
	f := covariance / math.Sqrt(varianceGuess*varianceActual)
	return &f
}

func mean(data []float64) float64 {
	var sum float64
	for _, value := range data {
		sum += value
	}
	return sum / float64(len(data))
}

func (s *Stats) ToJson() []JsonStats {
	var jsonStats []JsonStats

	ratedTheMost := JsonStats{
		Title:       "Rated The Most Wines",
		Description: "User with the highest number of wine ratings (the drunkest?):",
		Items:       []string{fmt.Sprintf("%s: %d ratings", s.RatedTheMostWines.Name, s.RatedTheMostWines.Value)},
	}
	for _, user := range s.RatedTheMostWines.Name {
		ratedTheMost.Items = append(ratedTheMost.Items, user)
	}
	jsonStats = append(jsonStats, ratedTheMost)

	// Convert AverageWineRatings
	averageRatings := JsonStats{
		Title:       "Average Wine Ratings by Users",
		Description: "Are you a generous rater?",
	}
	for _, stat := range s.AverageWineRatings {
		averageRatings.Items = append(averageRatings.Items, fmt.Sprintf("%s: %.2f stars", stat.Name, stat.Value))
	}
	jsonStats = append(jsonStats, averageRatings)

	// Convert WineRanking
	wineRanking := JsonStats{
		Title:       "Wine Rankings",
		Description: "Which wines did we like best?",
	}
	for _, stat := range s.WineRanking {
		wineRanking.Items = append(wineRanking.Items, fmt.Sprintf("%s: %.2f stars", stat.Name, stat.Value))
	}
	jsonStats = append(jsonStats, wineRanking)

	// Convert WineValue
	wineValue := JsonStats{
		Title:       "Best Bang for you Buck",
		Description: "Average Rating/ Price",
	}
	for _, stat := range s.WineValue {
		wineValue.Items = append(wineValue.Items, fmt.Sprintf("%s: %.2f points", stat.Name, stat.Value))
	}
	jsonStats = append(jsonStats, wineValue)

	// Convert PriceGuessingCorrelations
	priceGuessing := JsonStats{
		Title:       "Price Guessing Correlation",
		Description: "How do your taste buds stack up?",
	}
	for _, stat := range s.PriceGuessingCorrelations {
		priceGuessing.Items = append(priceGuessing.Items, fmt.Sprintf("%s: %.2f correlation", stat.Name, stat.Value))
	}
	jsonStats = append(jsonStats, priceGuessing)

	return jsonStats
}
