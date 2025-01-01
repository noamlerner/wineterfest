package datamodels

type WineRating struct {
	AnonymizedNumber int    `json:"anonymizedNumber"  dynamodbav:"n"`
	Rating           int    `json:"rating"  dynamodbav:"wineRating"`
	WineUser         string `json:"wineuser"  dynamodbav:"u"`
}

type User struct {
	Username string `json:"username"`
}

type Wine struct {
	WineName         string  `json:"wineName"  dynamodbav:"wineName"`
	WinePrice        float64 `json:"winePrice"  dynamodbav:"winePrice"`
	AnonymizedNumber int     `json:"anonymizedNumber" dynamodbav:"n"`
	Username         string  `json:"username" dynamodbav:"u"`
}
