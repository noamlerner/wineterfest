package winedb

import (
	"context"
	"fmt"
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

func (cl *Client) MyWineRatings(ctx context.Context, username string) ([]datamodels.WineRating, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(ratingsTableName),
		KeyConditionExpression: aws.String("#pk = :pkValue"),
		ExpressionAttributeNames: map[string]string{
			"#pk": usernamePropertyKey,
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pkValue": &types.AttributeValueMemberS{Value: username},
		},
	}

	// Execute the query
	result, err := cl.CL.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}

	// Unmarshal the result into a slice of WineRating
	ratings := make([]datamodels.WineRating, 0, len(result.Items)-1)
	for _, item := range result.Items {
		var rating datamodels.WineRating
		if err := attributevalue.UnmarshalMap(item, &rating); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}
		if rating.AnonymizedNumber == -1 {
			continue
		}
		ratings = append(ratings, rating)
	}
	return ratings, nil
}

func (cl *Client) AllWines(ctx context.Context) ([]datamodels.Wine, error) {
	// Prepare the scan input.
	input := &dynamodb.ScanInput{
		TableName: aws.String(winesTableName),
	}

	var allItems []datamodels.Wine
	paginator := dynamodb.NewScanPaginator(cl.CL, input)

	// Paginate through all items.
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page: %w", err)
		}

		for _, item := range page.Items {
			var wine datamodels.Wine
			if err := attributevalue.UnmarshalMap(item, &wine); err != nil {
				return nil, fmt.Errorf("failed to unmarshal item: %w", err)
			}
			allItems = append(allItems, wine)
		}
	}

	return allItems, nil
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
