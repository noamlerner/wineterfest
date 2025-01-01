package winedb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"log"
	"wineterfest/datamodels"
)

const (
	winesTableName   = "wines"
	ratingsTableName = "ratings"

	wineNumberPropertyKey = "n"
	usernamePropertyKey   = "u"
)

type Client struct {
	CL *dynamodb.Client
}

func Conn() *Client {
	// Load the default AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create a DynamoDB client
	svc := dynamodb.NewFromConfig(cfg)

	return &Client{CL: svc}
}

func (cl *Client) CreateWineRating(ctx context.Context, w *datamodels.WineRating) error {
	marshalMap, err := attributevalue.MarshalMap(w)
	if err != nil {
		return err
	}
	_, err = cl.CL.PutItem(ctx, &dynamodb.PutItemInput{
		Item:      marshalMap,
		TableName: aws.String(ratingsTableName),
	})
	return err
}

func (cl *Client) CreateWine(ctx context.Context, w *datamodels.Wine) error {
	marshalMap, err := attributevalue.MarshalMap(w)
	if err != nil {
		return err
	}
	_, err = cl.CL.PutItem(ctx, &dynamodb.PutItemInput{
		Item:                marshalMap,
		TableName:           aws.String(winesTableName),
		ConditionExpression: aws.String("attribute_not_exists(#num)"),
		ExpressionAttributeNames: map[string]string{
			"#num": wineNumberPropertyKey,
		},
	})
	return err
}

func (cl *Client) CreateUser(ctx context.Context, user string) error {
	_, err := cl.CL.PutItem(ctx, &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			usernamePropertyKey: &types.AttributeValueMemberS{
				Value: user,
			},
			wineNumberPropertyKey: &types.AttributeValueMemberN{
				Value: "-1",
			},
		},
		TableName:           aws.String(ratingsTableName),
		ConditionExpression: aws.String("attribute_not_exists(#username)"),
		ExpressionAttributeNames: map[string]string{
			"#username": usernamePropertyKey,
		},
	})
	return err
}
