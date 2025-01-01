package datamodels

type WineRating struct {
	AnonymizedNumber int    `json:"anonymizedNumber"  dynamodbav:"n"`
	Rating           int    `json:"rating"  dynamodbav:"r"`
	WineUser         string `json:"wineuser"  dynamodbav:"u"`
}

type User struct {
	Username string `json:"username"`
}

type Wine struct {
	WineName         string  `json:"wineName"  dynamodbav:"t"`
	WinePrice        float64 `json:"winePrice"  dynamodbav:"t"`
	AnonymizedNumber int     `json:"anonymizedNumber" dynamodbav:"n"`
	Username         string  `json:"username" dynamodbav:"u"`
}
