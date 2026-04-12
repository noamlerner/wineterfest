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

	return ap(nil,
		howAUserRates(ratingsByUser),
		mostRated(ratingsByWine, numToWine),
		userCorrelationCoefficient(ratingsByUser, numToWine),
		bestPriceGuess(ratingsByUser, numToWine),
		mostContrarian(wineRankings, ratingsByUser),
		trueToTheCrowd(wineRankings, ratingsByUser),
		controversialWine(ratingsByWine, numToWine),
		topValueWine(wineRankings),
		bestWine(wineRankings),
	)
}

func ap(s []JsonStats, stats ...JsonStats) []JsonStats {
	return append(s, stats...)
}

func howAUserRates(ratingsByUser map[string][]*datamodels.WineRating) JsonStats {
	stat := JsonStats{
		Title:       "Your rating stats",
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
		Title:       "Price Guessing",
		Description: "Bring out th Sommelier",
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

func mostRated(ratingsByWine map[int][]*datamodels.WineRating, numToWine map[int]*datamodels.Wine) JsonStats {
	entries := make([]Stat[*datamodels.Wine], 0, len(ratingsByWine))
	for num, ratings := range ratingsByWine {
		wine, ok := numToWine[num]
		if !ok {
			continue
		}
		sum := 0.0
		for _, r := range ratings {
			sum += float64(r.Rating)
		}
		entries = append(entries, Stat[*datamodels.Wine]{
			Name:       wine,
			IntValue:   len(ratings),
			FloatValue: sum / float64(len(ratings)),
		})
	}
	sortInt(entries)

	jStat := JsonStats{
		Title:       "Most Rated",
		Description: "The wines everyone had an opinion on",
		Table:       make([][]string, 0, len(entries)+1),
	}
	jStat.Table = append(jStat.Table, []string{"Wine #", "Wine Name", "# of Ratings", "Avg Rating", "Brought By"})
	for _, e := range entries {
		jStat.Table = append(jStat.Table, []string{
			strconv.Itoa(e.Name.AnonymizedNumber),
			e.Name.WineName,
			strconv.Itoa(e.IntValue),
			fmt.Sprintf("%.2f", e.FloatValue),
			e.Name.BroughtBy(),
		})
	}
	return jStat
}

func bestPriceGuess(ratingsByUser map[string][]*datamodels.WineRating, numToWine map[int]*datamodels.Wine) JsonStats {
	type guessEntry struct {
		user   string
		wine   *datamodels.Wine
		guess  float64
		actual float64
		diff   float64
	}

	var entries []guessEntry
	for user, ratings := range ratingsByUser {
		for _, r := range ratings {
			wine, ok := numToWine[r.AnonymizedNumber]
			if !ok || wine.WinePrice == 0 {
				continue
			}
			entries = append(entries, guessEntry{
				user:   user,
				wine:   wine,
				guess:  r.PriceGuess,
				actual: wine.WinePrice,
				diff:   math.Abs(r.PriceGuess - wine.WinePrice),
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].diff == entries[j].diff {
			return entries[i].actual > entries[j].actual
		}
		return entries[i].diff < entries[j].diff
	})

	jStat := JsonStats{
		Title:       "Best Price Guess",
		Description: "Closest single guess to the actual price",
		Table:       make([][]string, 0, len(entries)+1),
	}
	jStat.Table = append(jStat.Table, []string{"User", "Wine Name", "Their Guess", "Actual Price", "Difference"})
	for _, e := range entries {
		jStat.Table = append(jStat.Table, []string{
			e.user,
			e.wine.WineName,
			fmt.Sprintf("$%.2f", e.guess),
			fmt.Sprintf("$%.2f", e.actual),
			fmt.Sprintf("$%.2f", e.diff),
		})
	}
	return jStat
}

func mostContrarian(wineRankings []CrowdWineRating, ratingsByUser map[string][]*datamodels.WineRating) JsonStats {
	crowdAvg := make(map[int]float64, len(wineRankings))
	for _, wr := range wineRankings {
		crowdAvg[wr.Wine.AnonymizedNumber] = wr.Rating
	}

	contrarians := make([]Stat[string], 0, len(ratingsByUser))
	for user, ratings := range ratingsByUser {
		var totalDev float64
		var counted int
		for _, r := range ratings {
			avg, ok := crowdAvg[r.AnonymizedNumber]
			if !ok {
				continue
			}
			totalDev += math.Abs(float64(r.Rating) - avg)
			counted++
		}
		if counted == 0 {
			continue
		}
		contrarians = append(contrarians, Stat[string]{
			Name:       user,
			FloatValue: totalDev / float64(counted),
			IntValue:   counted,
		})
	}
	sortFloat(contrarians)

	jStat := JsonStats{
		Title:       "Most Contrarian",
		Description: "Whose ratings deviated most from the crowd",
		Table:       make([][]string, 0, len(contrarians)+1),
	}
	jStat.Table = append(jStat.Table, []string{"Name", "Avg Deviation from Crowd", "# of Ratings"})
	for _, c := range contrarians {
		jStat.Table = append(jStat.Table, []string{
			c.Name,
			fmt.Sprintf("%.2f", c.FloatValue),
			strconv.Itoa(c.IntValue),
		})
	}
	return jStat
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
