package repository

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"os"
	"strings"
	//"github.com/google/uuid"
)

const TableName = "Authorization"

func GetClient() *dynamodb.DynamoDB {
	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	return dynamodb.New(s, s.Config.WithEndpoint("http://localhost:8000"))
}

type Node struct {
	OrganisationId string `json:"organisationId"`
	Id             string `json:"id"`
	Type           string `json:"type"`
	Data           string `json:"data"`
}

type nodeDto struct {
	GlobalId   string `json:"globalId"`
	TypeTarget string `json:"typeTarget"`
	Node
}

const separator = '|'
const nodePrefix = "node_"

func createNodeDto(node *Node) *nodeDto {
	return &nodeDto{
		GlobalId:   fmt.Sprintf("%s_%s", node.OrganisationId, node.Id),
		TypeTarget: nodePrefix + strings.Join([]string{node.Type, node.Id}, string(separator)),
		Node:       *node,
	}
}

func CreateNode(node *Node) {
	client := GetClient()

	dto := createNodeDto(node)
	av, err := dynamodbattribute.MarshalMap(dto)

	fmt.Println(av)

	if err != nil {
		fmt.Println("Got error marshalling node:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	_, err = client.PutItem(&dynamodb.PutItemInput{
		//ConditionExpression: aws.String("attribute_not_exists(id)"),
		Item:      av,
		TableName: aws.String(TableName),
	})

	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func GetNodes(organisationId string, nodeType string) []Node {
	client := GetClient()
	output, err := client.Query(&dynamodb.QueryInput{
		IndexName:              aws.String("GSIApplicationTypeTarget"),
		TableName:              aws.String(TableName),
		KeyConditionExpression: aws.String("organisationId = :organisationId and begins_with(typeTarget, :type)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":organisationId": {
				S: aws.String(organisationId),
			},
			":type": {
				S: aws.String(nodePrefix + nodeType),
			},
		},
	})

	if err != nil {
		fmt.Println("Got error calling Query:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	result := make([]Node, *output.Count)
	for i, item := range output.Items {
		err := dynamodbattribute.UnmarshalMap(item, &result[i])
		if err != nil {
			fmt.Println("Can't unmarshal node")
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	return result
}
