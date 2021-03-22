package repository

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"testing"
)

func GetClient() *dynamodb.DynamoDB {
	// Create DynamoDB client
	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return dynamodb.New(s, s.Config.WithEndpoint("http://localhost:8000"))
}

func TestGetNodes(t *testing.T) {
	repository := CreateRepository(GetClient(), "test")
	node := Node{
		OrganisationId: uuid.New(),
		Id:             uuid.New(),
		Type:           "ROLE",
		Data:           "Branch manager",
	}
	err := repository.InsertNode(&node)
	if err != nil {
		t.Fatalf("Failed to insert node %v with error %v", node, err)
	}

	result, err := repository.GetNodes(node.OrganisationId, node.Type)
	if err != nil {
		t.Fatalf("Failed to get nodes with error %v", err)
	}

	if node != result[0] {
		t.Errorf("Expect %s, got %s", node, result[0])
	}
}

func ExampleString() {
	s := "world"
	f := func(a string) bool {
		return &a == &s
	}
	fmt.Println(f(s))
	// Output: true1
}
