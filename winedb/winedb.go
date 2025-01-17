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
	"strconv"
	"time"
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

func (cl *Client) MyWineRatings(ctx context.Context, user *datamodels.User) ([]*datamodels.WineRating, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(ratingsTableName),
		KeyConditionExpression: aws.String("#pk = :pkValue"),
		ExpressionAttributeNames: map[string]string{
			"#pk": usernamePropertyKey,
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pkValue": &types.AttributeValueMemberS{Value: user.Username},
		},
	}

	// Execute the query
	result, err := cl.CL.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}

	// Unmarshal the result into a slice of WineRating
	ratings := make([]*datamodels.WineRating, 0, len(result.Items))
	for _, item := range result.Items {
		rating := &datamodels.WineRating{}
		if err := attributevalue.UnmarshalMap(item, rating); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}
		if rating.AnonymizedNumber == -1 {
			continue
		}
		ratings = append(ratings, rating.Normalize())
	}
	return ratings, nil
}

func (cl *Client) AllWines(ctx context.Context) ([]*datamodels.Wine, error) {
	// Prepare the scan input.
	input := &dynamodb.ScanInput{
		TableName: aws.String(winesTableName),
	}

	var allItems []*datamodels.Wine
	paginator := dynamodb.NewScanPaginator(cl.CL, input)

	// Paginate through all items.
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page: %w", err)
		}

		for _, item := range page.Items {
			wine := &datamodels.Wine{}
			if err := attributevalue.UnmarshalMap(item, wine); err != nil {
				return nil, fmt.Errorf("failed to unmarshal item: %w", err)
			}
			allItems = append(allItems, wine.Normalize())
		}
	}

	return allItems, nil
}

func (cl *Client) AllRatings(ctx context.Context) ([]*datamodels.WineRating, error) {
	// Prepare the scan input.
	input := &dynamodb.ScanInput{
		TableName: aws.String(ratingsTableName),
	}

	var allItems []*datamodels.WineRating
	paginator := dynamodb.NewScanPaginator(cl.CL, input)

	// Paginate through all items.
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page: %w", err)
		}

		for _, item := range page.Items {
			wine := &datamodels.WineRating{}
			if err := attributevalue.UnmarshalMap(item, wine); err != nil {
				return nil, fmt.Errorf("failed to unmarshal item: %w", err)
			}
			allItems = append(allItems, wine.Normalize())
		}
	}

	return allItems, nil
}

func (cl *Client) CreateWineRating(ctx context.Context, w *datamodels.WineRating) error {
	w.TimeStampMilli = time.Now().UnixMilli()
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

func (cl *Client) GetWine(ctx context.Context, num int) (*datamodels.Wine, error) {
	// Define the GetItem input
	input := &dynamodb.GetItemInput{
		TableName: aws.String(winesTableName),
		Key: map[string]types.AttributeValue{
			wineNumberPropertyKey: &types.AttributeValueMemberN{Value: strconv.Itoa(num)},
		},
	}

	// Fetch the item from the table
	result, err := cl.CL.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	if result == nil || result.Item == nil {
		return nil, nil
	}
	wine := &datamodels.Wine{}
	if err := attributevalue.UnmarshalMap(result.Item, wine); err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	// If the Item is nil, the item does not exist
	return wine.Normalize(), nil
}

func (cl *Client) CreateUser(ctx context.Context, user *datamodels.User) error {
	_, err := cl.CL.PutItem(ctx, &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			usernamePropertyKey: &types.AttributeValueMemberS{
				Value: user.Username,
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

func (cl *Client) UserExists(ctx context.Context, user *datamodels.User) bool {
	// Define the GetItem input
	input := &dynamodb.GetItemInput{
		TableName: aws.String(ratingsTableName),
		Key: map[string]types.AttributeValue{
			usernamePropertyKey: &types.AttributeValueMemberS{
				Value: user.Username,
			},
			wineNumberPropertyKey: &types.AttributeValueMemberN{
				Value: "-1",
			},
		},
	}

	// Fetch the item from the table
	result, err := cl.CL.GetItem(ctx, input)
	if err != nil {
		return false
	}
	if result == nil || result.Item == nil {
		return false
	}
	return true
}

func (cl *Client) UsersWines(ctx context.Context, user *datamodels.User) ([]*datamodels.Wine, error) {
	wines, err := cl.AllWines(ctx)
	if err != nil {
		return nil, err
	}

	myWines := make([]*datamodels.Wine, 0, len(wines))
	for _, wine := range wines {
		if wine.Username == user.Username {
			myWines = append(myWines, wine)
		}
	}
	return myWines, nil
}
