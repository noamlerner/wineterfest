package stats

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"wineterfest/datamodels"
)

type Stat[T any] struct {
	Name       T
	FloatValue float64
	IntValue   int
}
type JsonStats struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Table       [][]string `json:"table"`
}

func Calc(allWines []datamodels.Wine, allRatings []datamodels.WineRating) []JsonStats {

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
	s := []JsonStats{}

	s = ap(s, averageRatings(ratingsByUser))
	s = ap(s, bestWine(ratingsByWine, numToWine))
	s = ap(s, topValueWine(ratingsByWine, numToWine))
	s = ap(s, userCorrelationCoefficient(ratingsByUser, numToWine))
	return s
}

func ap(s []JsonStats, stats ...JsonStats) []JsonStats {
	return append(s, stats...)
}

func averageRatings(ratingsByUser map[string][]datamodels.WineRating) JsonStats {
	stat := JsonStats{
		Title:       "Average Wine Ratings by Users",
		Description: "Are you a generous rater?",
		Table:       make([][]string, 0, len(ratingsByUser)+1),
	}
	stat.Table = append(stat.Table, []string{"User", "Number of Ratings", "Average Rating"})

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
			Name: user,
			// Number of Ratings
			IntValue: len(ratingsByUser[user]),
			// Average Rating
			FloatValue: sum / float64(len(ratingsByUser[user])),
		})
	}
	sortInt(averageRatingsPerUser)
	for _, user := range averageRatingsPerUser {
		stat.Table = append(stat.Table, []string{
			user.Name,
			fmt.Sprintf("%d", user.IntValue),
			fmt.Sprintf("%f", user.FloatValue),
		})
	}
	return stat
}

func bestWine(ratingsByWine map[int][]datamodels.WineRating, numToWine map[int]datamodels.Wine) JsonStats {
	stat := JsonStats{
		Title:       "Best Wine",
		Description: "Wine rankings",
	}

	wineRankings := make([]Stat[datamodels.Wine], 0, len(ratingsByWine))

	for num, ratings := range ratingsByWine {
		if numToWine[num].WinePrice == 0.0 {
			continue
		}
		sum := 0.0
		for _, rating := range ratings {
			sum += float64(rating.Rating)
		}
		wineRankings = append(wineRankings, Stat[datamodels.Wine]{
			Name:       numToWine[num],
			FloatValue: sum / float64(len(ratingsByWine[num])),
			IntValue:   len(ratingsByWine[num]),
		})
	}
	sortFloat(wineRankings)

	stat.Table = make([][]string, 0, len(ratingsByWine)+1)
	stat.Table = append(stat.Table, []string{
		"Wine #", "Wine Name", "Rated", "# of Ratings", "Cost", "Brought By",
	})
	for _, wine := range wineRankings {
		stat.Table = append(stat.Table, []string{
			strconv.Itoa(wine.Name.AnonymizedNumber),
			wine.Name.WineName,
			fmt.Sprintf("%.2f", wine.FloatValue),
			strconv.Itoa(wine.IntValue),
			fmt.Sprintf("$%.2f", wine.Name.WinePrice),
			wine.Name.Username,
		})
	}
	return stat
}

func topValueWine(ratingsByWine map[int][]datamodels.WineRating, numToWine map[int]datamodels.Wine) JsonStats {
	stat := JsonStats{
		Title:       "Top Value",
		Description: "Best Bang for you Buck",
	}

	wineRankings := make([]Stat[datamodels.Wine], 0, len(ratingsByWine))

	for num, ratings := range ratingsByWine {
		if numToWine[num].WinePrice == 0.0 {
			continue
		}
		sum := 0.0
		for _, rating := range ratings {
			sum += float64(rating.Rating)
		}
		wineRankings = append(wineRankings, Stat[datamodels.Wine]{
			Name:       numToWine[num],
			FloatValue: math.Pow(sum, 4) / numToWine[num].WinePrice,
			IntValue:   len(ratingsByWine[num]),
		})
	}
	sortFloat(wineRankings)

	stat.Table = make([][]string, 0, len(ratingsByWine)+1)
	stat.Table = append(stat.Table, []string{
		"Wine #", "Wine Name", "Value Score", "#Ratings", "Cost", "Brought By",
	})
	for _, wine := range wineRankings {
		stat.Table = append(stat.Table, []string{
			strconv.Itoa(wine.Name.AnonymizedNumber),
			wine.Name.WineName,
			fmt.Sprintf("%.2f", wine.FloatValue),
			strconv.Itoa(wine.IntValue),
			fmt.Sprintf("$%.2f", wine.Name.WinePrice),
			wine.Name.Username,
		})
	}
	return stat
}

func userCorrelationCoefficient(ratingsByUser map[string][]datamodels.WineRating, numToWine map[int]datamodels.Wine) JsonStats {
	stat := JsonStats{
		Title:       "Price Guessing Correlation",
		Description: "How do your taste buds stack up?",
	}

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
			Name:       user,
			FloatValue: *correlation,
			IntValue:   len(ratingsByUser[user]),
		})
	}
	sortFloat(correlations)

	stat.Table = make([][]string, 0, len(correlations)+1)
	stat.Table = append(stat.Table, []string{
		"Name", "Price Guessing Correlation", "Number Of Guesses",
	})
	for _, correlation := range correlations {
		stat.Table = append(stat.Table, []string{
			correlation.Name,
			fmt.Sprintf("%.2f", correlation.FloatValue),
			fmt.Sprintf("%d", correlation.IntValue),
		})
	}
	return stat
}

func sortFloat[T any](sli []Stat[T]) {
	sort.Slice(sli, func(i, j int) bool {
		if sli[i].FloatValue == sli[j].FloatValue {
			return sli[i].IntValue > sli[j].IntValue
		}
		return sli[i].FloatValue > sli[j].FloatValue
	})
}

func sortInt[T any](sli []Stat[T]) {
	sort.Slice(sli, func(i, j int) bool {
		if sli[i].IntValue == sli[j].IntValue {
			return sli[i].FloatValue > sli[j].FloatValue
		}
		return sli[i].IntValue > sli[j].IntValue
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
