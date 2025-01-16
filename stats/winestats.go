package stats

import (
	"fmt"
	"gonum.org/v1/gonum/stat"
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

type CrowdWineRating struct {
	Wine       *datamodels.Wine
	Rating     float64
	NumRatings int
}
type JsonStats struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Table       [][]string `json:"table"`
	GraphData   [][]int    `json:"graphData"`
}

func Calc(allWines []*datamodels.Wine, allRatings []*datamodels.WineRating) []JsonStats {
	ratingsByUser := make(map[string][]*datamodels.WineRating)
	ratingsByWine := make(map[int][]*datamodels.WineRating)
	for _, rating := range allRatings {
		if rating.AnonymizedNumber == -1 {
			continue
		}
		ratingsByUser[rating.WineUser] = append(ratingsByUser[rating.WineUser], rating)
		ratingsByWine[rating.AnonymizedNumber] = append(ratingsByWine[rating.AnonymizedNumber], rating)
	}
	numToWine := make(map[int]*datamodels.Wine)
	for _, wine := range allWines {
		numToWine[wine.AnonymizedNumber] = wine
	}

	wineRankings := generateWineRankings(ratingsByWine, numToWine)

	s := []JsonStats{}
	s = ap(s, howAUserRates(ratingsByUser))
	s = ap(s, bestWine(wineRankings))
	s = ap(s, topValueWine(wineRankings))
	s = ap(s, userCorrelationCoefficient(ratingsByUser, numToWine))
	s = ap(s, trueToTheCrowd(wineRankings, ratingsByUser))
	s = ap(s, controversialWine(ratingsByWine, numToWine))
	controversialWine(ratingsByWine, numToWine)
	return s
}

func ap(s []JsonStats, stats ...JsonStats) []JsonStats {
	return append(s, stats...)
}

func howAUserRates(ratingsByUser map[string][]*datamodels.WineRating) JsonStats {
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

func bestWine(wineRankings []CrowdWineRating) JsonStats {
	stat := JsonStats{
		Title:       "Best Wine",
		Description: "Wine rankings",
	}

	stat.Table = make([][]string, 0, len(wineRankings)+1)
	stat.Table = append(stat.Table, []string{
		"Wine #", "Wine Name", "Rated", "# of Ratings", "Cost", "Brought By",
	})
	for _, wine := range wineRankings {
		stat.Table = append(stat.Table, []string{
			strconv.Itoa(wine.Wine.AnonymizedNumber),
			wine.Wine.WineName,
			fmt.Sprintf("%.2f", wine.Rating),
			strconv.Itoa(wine.NumRatings),
			fmt.Sprintf("$%.2f", wine.Wine.WinePrice),
			wine.Wine.BroughtBy(),
		})
	}
	return stat
}

func controversialWine(wineRatings map[int][]*datamodels.WineRating, numToWine map[int]*datamodels.Wine) JsonStats {
	stdevs := make([]Stat[*datamodels.Wine], 0, len(wineRatings))
	for _, wine := range wineRatings {
		allRatings := make([]float64, 0, len(wineRatings))
		stdevs = append(stdevs, Stat[*datamodels.Wine]{})

		for _, rating := range wine {
			allRatings = append(allRatings, float64(rating.Rating))
			stdevs[len(stdevs)-1].Name = numToWine[rating.AnonymizedNumber]
		}
		dev := stat.PopStdDev(allRatings, nil)
		stdevs[len(stdevs)-1].FloatValue = dev
		stdevs[len(stdevs)-1].IntValue = len(wine)
	}

	sortFloat(stdevs)

	jStat := JsonStats{
		Title:       "Most Controversial Wine",
		Description: "Most loved and hated",
	}

	jStat.Table = make([][]string, 0, len(stdevs)+1)
	jStat.Table = append(jStat.Table, []string{
		"Wine #", "Wine Name", "Controversy Score", "#Ratings", "Brought By",
	})
	for _, wine := range stdevs {
		jStat.Table = append(jStat.Table, []string{
			strconv.Itoa(wine.Name.AnonymizedNumber),
			wine.Name.WineName,
			fmt.Sprintf("%.2f", wine.FloatValue*100),
			strconv.Itoa(wine.IntValue),
			wine.Name.BroughtBy(),
		})
	}
	return jStat
}

func generateWineRankings(ratingsByWine map[int][]*datamodels.WineRating, numToWine map[int]*datamodels.Wine) []CrowdWineRating {
	wineRankings := make([]CrowdWineRating, 0, len(ratingsByWine))
	for num, ratings := range ratingsByWine {
		if numToWine[num].WinePrice == 0.0 {
			continue
		}
		sum := 0.0
		for _, rating := range ratings {
			sum += float64(rating.Rating)
		}
		wineRankings = append(wineRankings, CrowdWineRating{
			Wine:       numToWine[num],
			Rating:     sum / float64(len(ratingsByWine[num])),
			NumRatings: len(ratingsByWine[num]),
		})
	}
	sort.Slice(wineRankings, func(i, j int) bool {
		if wineRankings[i].Rating == wineRankings[j].Rating {
			return wineRankings[i].NumRatings > wineRankings[j].NumRatings
		}
		return wineRankings[i].Rating > wineRankings[j].Rating
	})
	return wineRankings
}

func topValueWine(wineRankings []CrowdWineRating) JsonStats {
	stat := JsonStats{
		Title:       "Top Value",
		Description: "Best Bang for you Buck",
	}

	value := make([]Stat[*datamodels.Wine], 0, len(wineRankings))

	for _, rating := range wineRankings {
		if rating.Wine.WinePrice == 0.0 {
			continue
		}

		value = append(value, Stat[*datamodels.Wine]{
			Name:       rating.Wine,
			FloatValue: math.Pow(math.E, rating.Rating) / rating.Wine.WinePrice,
			IntValue:   rating.NumRatings,
		})
	}
	sortFloat(value)

	stat.Table = make([][]string, 0, len(value)+1)
	stat.Table = append(stat.Table, []string{
		"Wine #", "Wine Name", "Value Score", "#Ratings", "Cost", "Brought By",
	})
	for _, wine := range value {
		stat.Table = append(stat.Table, []string{
			strconv.Itoa(wine.Name.AnonymizedNumber),
			wine.Name.WineName,
			fmt.Sprintf("%.2f", wine.FloatValue),
			strconv.Itoa(wine.IntValue),
			fmt.Sprintf("$%.2f", wine.Name.WinePrice),
			wine.Name.BroughtBy(),
		})
	}
	return stat
}

func trueToTheCrowd(wineRankings []CrowdWineRating, ratingsByUser map[string][]*datamodels.WineRating) JsonStats {
	stat := JsonStats{
		Title:       "True to the Crowd",
		Description: "Who's taste best aligned with the crowd's opinion?",
	}

	numToCrowdRanking := make(map[int]CrowdWineRating, len(wineRankings))
	for _, wine := range wineRankings {
		numToCrowdRanking[wine.Wine.AnonymizedNumber] = wine
	}

	correlations := make([]Stat[string], 0, len(ratingsByUser))
	for user, ratings := range ratingsByUser {
		guesses := make([]float64, 0, len(ratings))
		actualRating := make([]float64, 0, len(ratings))
		for _, rating := range ratings {
			guesses = append(guesses, float64(rating.Rating))
			actualRating = append(actualRating, numToCrowdRanking[rating.AnonymizedNumber].Rating)
		}
		correlation := calculateCorrelation(guesses, actualRating)
		if correlation == nil {
			continue
		}
		correlations = append(correlations, Stat[string]{
			Name:       user,
			FloatValue: *correlation,
			IntValue:   len(ratingsByUser[user]),
		})
	}

	stat.Table = make([][]string, 0, len(ratingsByUser)+1)
	stat.Table = append(stat.Table, []string{
		"Name", "Wine Rating Correlation", "Number of Ratings",
	})
	stat.GraphData = make([][]int, 0, len(ratingsByUser))
	sortFloat(correlations)
	for i, correlation := range correlations {
		stat.Table = append(stat.Table, []string{
			correlation.Name,
			fmt.Sprintf("%d", int(correlation.FloatValue*100)),
			fmt.Sprintf("%d", correlation.IntValue),
		})
		stat.GraphData = append(stat.GraphData, []int{
			i, int(correlation.FloatValue),
		})
	}
	return stat
}

func userCorrelationCoefficient(ratingsByUser map[string][]*datamodels.WineRating, numToWine map[int]*datamodels.Wine) JsonStats {
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
			fmt.Sprintf("%d", int(correlation.FloatValue*100)),
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
