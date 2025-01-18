package datamodels

import "wineterfest/utils"

type WineRating struct {
	AnonymizedNumber int     `json:"anonymizedNumber"  dynamodbav:"n"`
	Rating           int     `json:"rating"  dynamodbav:"wineRating"`
	WineUser         string  `json:"wineuser"  dynamodbav:"u"`
	PriceGuess       float64 `json:"priceGuess" dynamodbav:"price"`
	TimeStampMilli   int64   `json:"timeStamp"  dynamodbav:"tsmilli"`
}

type User struct {
	Username string `json:"username"`
}

type Wine struct {
	WineName         string  `json:"wineName"  dynamodbav:"wineName"`
	WinePrice        float64 `json:"winePrice"  dynamodbav:"winePrice"`
	AnonymizedNumber int     `json:"anonymizedNumber" dynamodbav:"n"`
	Username         string  `json:"username" dynamodbav:"u"`
	BroughtWith      string  `json:"broughtWith" dynamodbav:"broughtWith"`
}

func (w *Wine) BroughtBy() string {
	if w.BroughtWith == "" {
		return w.Username
	}
	return w.Username + ", " + w.BroughtWith
}

func (w *Wine) Normalize() *Wine {
	w.WineName = utils.Normalize(w.WineName)
	w.BroughtWith = utils.Normalize(w.BroughtWith)
	w.Username = utils.Normalize(w.Username)
	return w
}

func (w *User) Normalize() *User {
	w.Username = utils.Normalize(w.Username)
	return w
}

func (w *WineRating) Normalize() *WineRating {
	w.WineUser = utils.Normalize(w.WineUser)
	return w
}
