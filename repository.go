package repository

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"strings"
)

const TableName = "Authorization"

func GetClient() *dynamodb.DynamoDB {
	// Create DynamoDB client
	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return dynamodb.New(s, s.Config.WithEndpoint("http://localhost:8000"))
}

type Node struct {
	OrganisationId uuid.UUID `json:"organisationId"`
	Id             uuid.UUID `json:"id"`
	Type           string    `json:"type"`
	Data           string    `json:"data"`
}

type nodeDto struct {
	GlobalId       string `json:"globalId"`
	TypeTarget     string `json:"typeTarget"`
	OrganisationId string `json:"organisationId"`
	Id             string `json:"id"`
	Type           string `json:"type"`
	Data           string `json:"data"`
}

const separator = '|'
const nodePrefix = "node_"

func (node *Node) createNodeDto() *nodeDto {
	return &nodeDto{
		GlobalId:       fmt.Sprintf("%s_%s", node.OrganisationId, node.Id),
		TypeTarget:     nodePrefix + strings.Join([]string{node.Type, node.Id.String()}, string(separator)),
		OrganisationId: node.OrganisationId.String(),
		Id:             node.Id.String(),
		Type:           node.Type,
		Data:           node.Data,
	}
}

func (n nodeDto) createNode() Node {
	return Node{
		OrganisationId: uuid.MustParse(n.OrganisationId),
		Id:             uuid.MustParse(n.Id),
		Type:           n.Type,
		Data:           n.Data,
	}
}

func InsertNode(node *Node) error {
	client := GetClient()

	dto := node.createNodeDto()
	av, err := dynamodbattribute.MarshalMap(dto)

	if err != nil {
		return err
	}

	_, err = client.PutItem(&dynamodb.PutItemInput{
		ConditionExpression: aws.String("attribute_not_exists(id)"),
		Item:                av,
		TableName:           aws.String(TableName),
	})

	if err != nil {
		return err
	}

	return nil
}

func GetNodes(organisationId uuid.UUID, nodeType string) ([]Node, error) {
	client := GetClient()
	output, err := client.Query(&dynamodb.QueryInput{
		IndexName:              aws.String("GSIApplicationTypeTarget"),
		TableName:              aws.String(TableName),
		KeyConditionExpression: aws.String("organisationId = :organisationId and begins_with(typeTarget, :type)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":organisationId": {
				S: aws.String(organisationId.String()),
			},
			":type": {
				S: aws.String(nodePrefix + nodeType),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	result := make([]Node, *output.Count)
	for i, item := range output.Items {
		var nodeDto = nodeDto{}
		err := dynamodbattribute.UnmarshalMap(item, &nodeDto)
		if err != nil {
			return nil, err
		}
		result[i] = nodeDto.createNode()
	}

	return result, nil
}
