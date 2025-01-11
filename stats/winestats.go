package stats

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"wineterfest/datamodels"
)

type Stat[T any] struct {
	Name   T
	Value  float64
	Value2 int
}

type Stats struct {
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
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Table       [][]string `json:"table"`
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
			Name:   user,
			Value:  sum / float64(len(ratingsByUser[user])),
			Value2: len(ratingsByUser[user]),
		})
	}
	s.AverageWineRatings = averageRatingsPerUser
	sort.Slice(s.AverageWineRatings, func(i, j int) bool {
		return s.AverageWineRatings[i].Value2 > s.AverageWineRatings[j].Value2
	})
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
			sum += float64(rating.Rating)
		}
		wineRankings = append(wineRankings, Stat[datamodels.Wine]{
			Name:   numToWine[num],
			Value:  sum / float64(len(ratingsByWine[num])),
			Value2: len(ratingsByWine[num]),
		})
		wineValues = append(wineValues, Stat[datamodels.Wine]{
			Name:   numToWine[num],
			Value:  sum * sum / numToWine[num].WinePrice,
			Value2: len(ratingsByWine[num]),
		})
	}
	sortSlice(wineRankings)
	s.WineRanking = wineRankings

	sortSlice(wineValues)
	s.WineValue = wineValues
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
			Name:   user,
			Value:  *correlation,
			Value2: len(ratingsByUser[user]),
		})
	}
	sortSlice(correlations)

	s.PriceGuessingCorrelations = correlations
}

func sortSlice[T any](sli []Stat[T]) {
	sort.Slice(sli, func(i, j int) bool {
		if sli[i].Value == sli[j].Value {
			return sli[i].Value2 > sli[j].Value2
		}
		return sli[i].Value > sli[j].Value
	})
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

	// Convert AverageWineRatings
	averageRatings := JsonStats{
		Title:       "Average Wine Ratings by Users",
		Description: "Are you a generous rater?",
		Table: [][]string{
			{"User", "Number of Ratings", "Average Rating"},
		},
	}
	for _, stat := range s.AverageWineRatings {
		averageRatings.Table = append(averageRatings.Table, []string{
			stat.Name,
			strconv.Itoa(stat.Value2),
			fmt.Sprintf("%.2f", stat.Value),
		})
	}
	jsonStats = append(jsonStats, averageRatings)

	// Convert WineRanking
	wineRanking := JsonStats{
		Title:       "Wine Rankings",
		Description: "Which wines did we like best?",
		Table: [][]string{
			{"Wine #", "Wine Name", "Rated", "# of Ratings", "Cost", "Brought By"},
		},
	}
	for _, stat := range s.WineRanking {
		wineRanking.Table = append(wineRanking.Table, []string{
			strconv.Itoa(stat.Name.AnonymizedNumber),
			stat.Name.WineName,
			fmt.Sprintf("%.2f", stat.Value),
			strconv.Itoa(stat.Value2),
			fmt.Sprintf("$%.2f", stat.Name.WinePrice),
			stat.Name.Username,
		})
	}
	jsonStats = append(jsonStats, wineRanking)

	// Convert WineValue
	wineValue := JsonStats{
		Title:       "Best Bang for you Buck",
		Description: "Average Rating/ Price",
		Table: [][]string{
			{"Wine #", "Wine Name", "Value Score", "#Ratings", "Cost", "Brought By"},
		},
	}
	for _, stat := range s.WineValue {
		wineValue.Table = append(wineValue.Table, []string{
			strconv.Itoa(stat.Name.AnonymizedNumber),
			stat.Name.WineName,
			fmt.Sprintf("%.2f", stat.Value),
			strconv.Itoa(int(stat.Value2)),
			fmt.Sprintf("$%.2f", stat.Name.WinePrice),
			stat.Name.Username,
		})
	}
	jsonStats = append(jsonStats, wineValue)

	// Convert PriceGuessingCorrelations
	priceGuessing := JsonStats{
		Title:       "Price Guessing Correlation",
		Description: "How do your taste buds stack up?",
		Table: [][]string{
			{
				"Name", "Price Guessing Correlation", "Number Of Guesses",
			},
		},
	}
	for _, stat := range s.PriceGuessingCorrelations {
		priceGuessing.Table = append(priceGuessing.Table, []string{
			stat.Name,
			fmt.Sprintf("%.2f", stat.Value),
			fmt.Sprintf("%d", stat.Value2),
		})
	}
	jsonStats = append(jsonStats, priceGuessing)

	return jsonStats
}
