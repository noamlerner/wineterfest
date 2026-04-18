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

	userAvg := userAverageRatings(ratingsByUser)
	wineRankings := generateWineRankings(ratingsByWine, numToWine, userAvg)
	looAvg := leaveOneOutWeightedAvg(ratingsByWine, userAvg)

	return ap(nil,
		howAUserRates(ratingsByUser, userAvg),
		userCorrelationCoefficient(ratingsByUser, numToWine),
		bestPriceGuess(ratingsByUser, numToWine),
		mostContrarian(looAvg, ratingsByUser, userAvg),
		trueToTheCrowd(looAvg, ratingsByUser, userAvg),
		controversialWine(ratingsByWine, numToWine, userAvg),
		topValueWine(wineRankings),
		bestWine(wineRankings),
		typeRankings(ratingsByWine, numToWine, userAvg),
		userFavoriteType(ratingsByUser, numToWine, userAvg),
	)
}

func ap(s []JsonStats, stats ...JsonStats) []JsonStats {
	return append(s, stats...)
}

func howAUserRates(ratingsByUser map[string][]*datamodels.WineRating, userAvg map[string]float64) JsonStats {
	globalMean := userAvg[""]
	st := JsonStats{
		Title:       "Your rating stats",
		Description: "Are you a generous rater?",
		Table:       make([][]string, 0, len(ratingsByUser)+1),
	}
	st.Table = append(st.Table, []string{"User", "# of Ratings", "Avg Rating", "Scaling Factor"})

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
			Name:     user,
			IntValue: len(ratings),
			FloatValue: sum / float64(len(ratings)),
		})
	}
	sortInt(averageRatingsPerUser)
	for _, user := range averageRatingsPerUser {
		scalingFactor := 1.0
		if user.FloatValue != 0 {
			scalingFactor = globalMean / user.FloatValue
		}
		st.Table = append(st.Table, []string{
			user.Name,
			fmt.Sprintf("%d", user.IntValue),
			fmt.Sprintf("%.2f", user.FloatValue),
			fmt.Sprintf("%.2f×", scalingFactor),
		})
	}
	return st
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

func controversialWine(wineRatings map[int][]*datamodels.WineRating, numToWine map[int]*datamodels.Wine, userAvg map[string]float64) JsonStats {
	globalMean := userAvg[""]
	stdevs := make([]Stat[*datamodels.Wine], 0, len(wineRatings))
	for _, wine := range wineRatings {
		allRatings := make([]float64, 0, len(wine))
		stdevs = append(stdevs, Stat[*datamodels.Wine]{})

		for _, rating := range wine {
			allRatings = append(allRatings, biasAdjusted(float64(rating.Rating), userAvg[rating.WineUser], globalMean))
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

// leaveOneOutWeightedAvg returns, for each user, the bias-adjusted crowd average
// for each wine they rated — computed from everyone else's ratings only.
func leaveOneOutWeightedAvg(ratingsByWine map[int][]*datamodels.WineRating, userAvg map[string]float64) map[string]map[int]float64 {
	globalMean := userAvg[""]
	out := make(map[string]map[int]float64)
	for wineNum, ratings := range ratingsByWine {
		for _, r := range ratings {
			user := r.WineUser
			if out[user] == nil {
				out[user] = make(map[int]float64)
			}
			var sum float64
			var count int
			for _, other := range ratings {
				if other.WineUser == user {
					continue
				}
				sum += biasAdjusted(float64(other.Rating), userAvg[other.WineUser], globalMean)
				count++
			}
			if count == 0 {
				continue
			}
			out[user][wineNum] = sum / float64(count)
		}
	}
	return out
}

// userAverageRatings returns each user's mean rating and the global mean.
func userAverageRatings(ratingsByUser map[string][]*datamodels.WineRating) map[string]float64 {
	avgs := make(map[string]float64, len(ratingsByUser))
	var globalSum float64
	var globalCount int
	for user, ratings := range ratingsByUser {
		if len(ratings) == 0 {
			continue
		}
		var sum float64
		for _, r := range ratings {
			sum += float64(r.Rating)
		}
		avgs[user] = sum / float64(len(ratings))
		globalSum += sum
		globalCount += len(ratings)
	}
	if globalCount > 0 {
		globalMean := globalSum / float64(globalCount)
		// Store global mean under empty key so callers can reference it.
		avgs[""] = globalMean
	}
	return avgs
}

// biasAdjusted scales a raw rating by (globalMean / userMean) so generous/stingy
// raters are normalised before wines are ranked.
func biasAdjusted(raw float64, userMean, globalMean float64) float64 {
	if userMean == 0 {
		return raw
	}
	return raw * (globalMean / userMean)
}

func generateWineRankings(ratingsByWine map[int][]*datamodels.WineRating, numToWine map[int]*datamodels.Wine, userAvg map[string]float64) []CrowdWineRating {
	globalMean := userAvg[""]
	wineRankings := make([]CrowdWineRating, 0, len(ratingsByWine))
	for num, ratings := range ratingsByWine {
		if numToWine[num].WinePrice == 0.0 {
			continue
		}
		sum := 0.0
		for _, rating := range ratings {
			sum += biasAdjusted(float64(rating.Rating), userAvg[rating.WineUser], globalMean)
		}
		wineRankings = append(wineRankings, CrowdWineRating{
			Wine:       numToWine[num],
			Rating:     sum / float64(len(ratings)),
			NumRatings: len(ratings),
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

func trueToTheCrowd(looAvg map[string]map[int]float64, ratingsByUser map[string][]*datamodels.WineRating, userAvg map[string]float64) JsonStats {
	globalMean := userAvg[""]
	stat := JsonStats{
		Title:       "True to the Crowd",
		Description: "Who's taste best aligned with the crowd's opinion?",
	}

	correlations := make([]Stat[string], 0, len(ratingsByUser))
	for user, ratings := range ratingsByUser {
		userWineAvgs, ok := looAvg[user]
		if !ok {
			continue
		}
		guesses := make([]float64, 0, len(ratings))
		actualRating := make([]float64, 0, len(ratings))
		for _, rating := range ratings {
			crowdScore, ok := userWineAvgs[rating.AnonymizedNumber]
			if !ok {
				continue
			}
			guesses = append(guesses, biasAdjusted(float64(rating.Rating), userAvg[user], globalMean))
			actualRating = append(actualRating, crowdScore)
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

func mostContrarian(looAvg map[string]map[int]float64, ratingsByUser map[string][]*datamodels.WineRating, userAvg map[string]float64) JsonStats {
	globalMean := userAvg[""]
	contrarians := make([]Stat[string], 0, len(ratingsByUser))
	for user, ratings := range ratingsByUser {
		userWineAvgs := looAvg[user]
		var totalDev float64
		var counted int
		for _, r := range ratings {
			crowdScore, ok := userWineAvgs[r.AnonymizedNumber]
			if !ok {
				continue
			}
			adjusted := biasAdjusted(float64(r.Rating), userAvg[user], globalMean)
			totalDev += math.Abs(adjusted - crowdScore)
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

func typeRankings(ratingsByWine map[int][]*datamodels.WineRating, numToWine map[int]*datamodels.Wine, userAvg map[string]float64) JsonStats {
	globalMean := userAvg[""]
	type typeAccum struct {
		sum   float64
		count int
	}
	byType := make(map[string]*typeAccum)

	for num, ratings := range ratingsByWine {
		wine, ok := numToWine[num]
		if !ok || wine.WineType == "" {
			continue
		}
		if byType[wine.WineType] == nil {
			byType[wine.WineType] = &typeAccum{}
		}
		for _, r := range ratings {
			byType[wine.WineType].sum += biasAdjusted(float64(r.Rating), userAvg[r.WineUser], globalMean)
			byType[wine.WineType].count++
		}
	}

	entries := make([]Stat[string], 0, len(byType))
	for t, acc := range byType {
		if acc.count == 0 {
			continue
		}
		entries = append(entries, Stat[string]{
			Name:       t,
			FloatValue: acc.sum / float64(acc.count),
			IntValue:   acc.count,
		})
	}
	sortFloat(entries)

	jStat := JsonStats{
		Title:       "Wine Type Rankings",
		Description: "Which style the crowd rated highest on average",
		Table:       make([][]string, 0, len(entries)+1),
	}
	jStat.Table = append(jStat.Table, []string{"Wine Type", "Avg Rating", "# of Ratings"})
	for _, e := range entries {
		jStat.Table = append(jStat.Table, []string{
			e.Name,
			fmt.Sprintf("%.2f", e.FloatValue),
			strconv.Itoa(e.IntValue),
		})
	}
	return jStat
}

func userFavoriteType(ratingsByUser map[string][]*datamodels.WineRating, numToWine map[int]*datamodels.Wine, userAvg map[string]float64) JsonStats {
	globalMean := userAvg[""]
	type typeAccum struct {
		sum   float64
		count int
	}

	type userEntry struct {
		user     string
		topType  string
		topAvg   float64
		numTypes int
	}

	entries := make([]userEntry, 0, len(ratingsByUser))
	for user, ratings := range ratingsByUser {
		byType := make(map[string]*typeAccum)
		for _, r := range ratings {
			wine, ok := numToWine[r.AnonymizedNumber]
			if !ok || wine.WineType == "" {
				continue
			}
			if byType[wine.WineType] == nil {
				byType[wine.WineType] = &typeAccum{}
			}
			byType[wine.WineType].sum += biasAdjusted(float64(r.Rating), userAvg[user], globalMean)
			byType[wine.WineType].count++
		}
		if len(byType) == 0 {
			continue
		}
		var topType string
		var topAvg float64
		for t, acc := range byType {
			avg := acc.sum / float64(acc.count)
			if avg > topAvg || topType == "" {
				topAvg = avg
				topType = t
			}
		}
		entries = append(entries, userEntry{
			user:     user,
			topType:  topType,
			topAvg:   topAvg,
			numTypes: len(byType),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].topAvg > entries[j].topAvg
	})

	jStat := JsonStats{
		Title:       "Your Taste Profile",
		Description: "Each person's highest-rated wine type",
		Table:       make([][]string, 0, len(entries)+1),
	}
	jStat.Table = append(jStat.Table, []string{"User", "Favorite Type", "Avg Rating for Type"})
	for _, e := range entries {
		jStat.Table = append(jStat.Table, []string{
			e.user,
			e.topType,
			fmt.Sprintf("%.2f", e.topAvg),
		})
	}
	return jStat
}

func mean(data []float64) float64 {
	var sum float64
	for _, value := range data {
		sum += value
	}
	return sum / float64(len(data))
}
