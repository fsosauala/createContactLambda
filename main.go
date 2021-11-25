package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
)

const (
	activeStatus = "active"
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
	log.Printf("REGION: %s", os.Getenv("AWS_REGION"))
	log.Println("ALL ENV VARS:")
	for _, element := range os.Environ() {
		log.Println(element)
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
