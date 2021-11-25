package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
)

const (
	activeStatus = "active"
	tableName    = "contactsFredy"
)

type UserRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type User struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	UserRequest
}

func HandleLambdaEvent(ur UserRequest) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(opts *config.LoadOptions) error {
		opts.Region = os.Getenv("AWS_REGION")
		return nil
	})
	if err != nil {
		log.Printf("error loading dynamo configuration: %v", err)
		return err
	}
	svc := dynamodb.NewFromConfig(cfg)
	p := dynamodb.NewListTablesPaginator(svc, nil, func(o *dynamodb.ListTablesPaginatorOptions) {
		o.StopOnDuplicateToken = true
	})

	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			log.Printf("error getting paget next: %v", err)
			return err
		}

		for _, tn := range out.TableNames {
			fmt.Println(tn)
		}
	}
	user := User{
		ID:          uuid.New().String(),
		Status:      activeStatus,
		UserRequest: ur,
	}
	log.Println(user)
	return nil
}

func main() {
	lambda.Start(HandleLambdaEvent)
}
