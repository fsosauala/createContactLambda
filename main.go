package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const (
	activeStatus = "active"
	tableName    = "contactsFredy"
)

type (
	UserRequest struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}
	User struct {
		key
		UserRequest
	}
	key struct {
		ID     string `dynamodbav:"id" json:"id"`
		Status string `dynamodbav:"status" json:"status"`
	}
)

func HandleLambdaEvent(ur UserRequest) error {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, func(opts *config.LoadOptions) error {
		opts.Region = os.Getenv("AWS_REGION")
		return nil
	})
	if err != nil {
		log.Printf("error loading dynamo configuration: %v", err)
		return err
	}
	svc := dynamodb.NewFromConfig(cfg)
	err = checkOrCreateDatabase(ctx, svc)
	if err != nil {
		log.Printf("error retrieving database information: %v", err)
		return err
	}
	userInserted, err := insertContact(ctx, svc, ur)
	if err != nil {
		log.Printf("error inserting contact: %v", err)
		return err
	}
	userRetrieved, err := retrieveContact(ctx, svc, userInserted.ID, activeStatus)
	if err != nil {
		log.Printf("error getting user information: %v", err)
		return err
	}
	log.Printf("Compare users.\nInserted: %v\nRetrieved: %v", userInserted, userRetrieved)
	return nil
}

func main() {
	lambda.Start(HandleLambdaEvent)
}

func checkOrCreateDatabase(ctx context.Context, svc *dynamodb.Client) error {
	p := dynamodb.NewListTablesPaginator(svc, nil, func(o *dynamodb.ListTablesPaginatorOptions) {
		o.StopOnDuplicateToken = true
	})

	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			log.Printf("error getting paget next: %v", err)
			return err
		}

		for _, tn := range out.TableNames {
			if tn == tableName {
				return nil
			}
		}
	}
	return createTable(ctx, svc)
}

func createTable(ctx context.Context, svc *dynamodb.Client) error {
	_, err := svc.CreateTable(ctx, &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("status"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("status"),
				KeyType:       types.KeyTypeRange,
			},
		},
		TableName:   aws.String(tableName),
		BillingMode: types.BillingModePayPerRequest,
	})
	return err
}

func insertContact(ctx context.Context, svc *dynamodb.Client, r UserRequest) (User, error) {
	key := key{
		ID:     uuid.New().String(),
		Status: activeStatus,
	}
	user := User{
		key:         key,
		UserRequest: r,
	}
	input := userToDynamoType(user)

	_, err := svc.PutItem(ctx, input)
	return user, err
}

func userToDynamoType(user User) *dynamodb.PutItemInput {
	item := userToAttributeMap(user)
	return &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	}
}

func retrieveContact(ctx context.Context, svc *dynamodb.Client, userID, status string) (User, error) {
	key := key{
		ID:     userID,
		Status: status,
	}
	avs, err := attributevalue.MarshalMap(key)
	if err != nil {
		panic(err)
	}

	out, err := svc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       avs,
	})
	if err != nil {
		return User{}, err
	}
	var u User
	err = attributevalue.UnmarshalMap(out.Item, &u)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func userToAttributeMap(user User) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{
			Value: user.ID,
		},
		"firstName": &types.AttributeValueMemberS{
			Value: user.FirstName,
		},
		"lastName": &types.AttributeValueMemberS{
			Value: user.LastName,
		},
		"status": &types.AttributeValueMemberS{
			Value: user.Status,
		},
	}
}
